/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, LessThan, Not, Repository } from 'typeorm'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { Image } from '../entities/image.entity'
import { ImageState } from '../enums/image-state.enum'
import { DockerProvider } from '../docker/docker-provider'
import { ImageNode } from '../entities/image-node.entity'
import { Node } from '../entities/node.entity'
import { NodeState } from '../enums/node-state.enum'
import { ImageNodeState } from '../enums/image-node-state.enum'
import { NodeApiFactory } from '../runner-api/runnerApi'
import { v4 as uuidv4 } from 'uuid'
import { NodeNotReadyError } from '../errors/node-not-ready.error'
import { RegistryType } from '../../docker-registry/enums/registry-type.enum'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { OrganizationService } from '../../organization/services/organization.service'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { BuildInfo } from '../entities/build-info.entity'
import { fromAxiosError } from '../../common/utils/from-axios-error'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
@Injectable()
export class ImageManager {
  private readonly logger = new Logger(ImageManager.name)
  //  generate a unique instance id used to ensure only one instance of the worker is handing the
  //  image activation
  private readonly instanceId = uuidv4()

  constructor(
    @InjectRedis() private readonly redis: Redis,
    @InjectRepository(Image)
    private readonly imageRepository: Repository<Image>,
    @InjectRepository(ImageNode)
    private readonly imageNodeRepository: Repository<ImageNode>,
    @InjectRepository(Node)
    private readonly nodeRepository: Repository<Node>,
    @InjectRepository(BuildInfo)
    private readonly buildInfoRepository: Repository<BuildInfo>,
    private readonly dockerRegistryService: DockerRegistryService,
    private readonly dockerProvider: DockerProvider,
    private readonly nodeApiFactory: NodeApiFactory,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly organizationService: OrganizationService,
  ) {}

  @Cron(CronExpression.EVERY_5_SECONDS)
  async syncNodeImages() {
    const lockKey = 'sync-node-images-lock'
    if (await this.redisLockProvider.lock(lockKey, 30)) {
      return
    }

    const skip = (await this.redis.get('sync-node-images-skip')) || 0

    const totalNodes = await this.nodeRepository.count({
      where: {
        state: NodeState.READY,
        unschedulable: false,
      },
    })

    const images = await this.imageRepository.find({
      where: {
        state: ImageState.ACTIVE,
      },
      order: {
        createdAt: 'ASC',
      },
      take: 100,
      skip: Number(skip),
    })

    if (images.length === 0) {
      await this.redis.set('sync-node-images-skip', 0)
      return
    }

    await this.redis.set('sync-node-images-skip', Number(skip) + images.length)

    const imageNodes = await this.imageNodeRepository.count({
      where: {
        imageRef: In(images.map((image) => image.internalName)),
        state: ImageNodeState.READY,
      },
    })

    if (imageNodes === totalNodes) {
      return
    }

    await Promise.all(
      images.map((image) => {
        this.propagateImageToNodes(image.internalName).catch((err) => {
          this.logger.error(`Error propagating image ${image.id} to nodes: ${err}`)
        })
      }),
    )

    await this.redisLockProvider.unlock(lockKey)
  }

  @Cron(CronExpression.EVERY_10_SECONDS)
  async syncNodeImageStates() {
    //  this approach is not ideal, as if the number of nodes is large, this will take a long time
    //  also, if some images stuck in a "pulling" state, they will infest the queue
    //  todo: find a better approach

    const lockKey = 'sync-node-image-states-lock'
    if (await this.redisLockProvider.lock(lockKey, 30)) {
      return
    }

    const nodeImages = await this.imageNodeRepository
      .createQueryBuilder('imageNode')
      .where({
        state: In([ImageNodeState.PULLING_IMAGE, ImageNodeState.BUILDING_IMAGE, ImageNodeState.REMOVING]),
      })
      .orderBy('RANDOM()')
      .take(100)
      .getMany()

    await Promise.allSettled(
      nodeImages.map((imageNode) => {
        return this.syncNodeImageState(imageNode).catch((err) => {
          if (err.code !== 'ECONNRESET') {
            this.logger.error(`Error syncing node image state ${imageNode.id}: ${fromAxiosError(err)}`)
            this.imageNodeRepository.update(imageNode.id, {
              state: ImageNodeState.ERROR,
              errorReason: fromAxiosError(err).message,
            })
          }
        })
      }),
    )

    await this.redisLockProvider.unlock(lockKey)
  }

  async syncNodeImageState(imageNode: ImageNode): Promise<void> {
    const node = await this.nodeRepository.findOne({
      where: {
        id: imageNode.nodeId,
      },
    })
    if (!node) {
      //  cleanup the image node record if the node is not found
      //  this can happen if the node is deleted from the database without cleaning up the image nodes
      await this.imageNodeRepository.delete(imageNode.id)
      this.logger.warn(
        `Node ${imageNode.nodeId} not found while trying to process image node ${imageNode.id}. Image node has been removed.`,
      )
      return
    }
    if (node.state !== NodeState.READY) {
      //  todo: handle timeout policy
      //  for now just remove the image node record if the node is not ready
      await this.imageNodeRepository.delete(imageNode.id)

      throw new NodeNotReadyError(`Node ${node.id} is not ready`)
    }

    switch (imageNode.state) {
      case ImageNodeState.PULLING_IMAGE:
        await this.handleImageNodeStatePullingImage(imageNode)
        break
      case ImageNodeState.BUILDING_IMAGE:
        await this.handleImageNodeStateBuildingImage(imageNode)
        break
      case ImageNodeState.REMOVING:
        await this.handleImageNodeStateRemovingImage(imageNode)
        break
    }
  }

  async propagateImageToNodes(internalImageName: string) {
    //  todo: remove try catch block and implement error handling
    try {
      const nodes = await this.nodeRepository.find({
        where: {
          state: NodeState.READY,
          unschedulable: false,
        },
      })

      const results = await Promise.allSettled(
        nodes.map(async (node) => {
          let imageNode = await this.imageNodeRepository.findOne({
            where: {
              imageRef: internalImageName,
              nodeId: node.id,
            },
          })

          try {
            if (imageNode && !imageNode.imageRef) {
              //  this should never happen
              this.logger.warn(`Internal image name not found for image node ${imageNode.id}`)
              return
            }

            if (!imageNode) {
              imageNode = new ImageNode()
              imageNode.imageRef = internalImageName
              imageNode.nodeId = node.id
              imageNode.state = ImageNodeState.PULLING_IMAGE
              await this.imageNodeRepository.save(imageNode)
              await this.propagateImageToNode(internalImageName, node)
            } else if (imageNode.state === ImageNodeState.PULLING_IMAGE) {
              await this.handleImageNodeStatePullingImage(imageNode)
            }
          } catch (err) {
            this.logger.error(`Error propagating image to node ${node.id}: ${fromAxiosError(err)}`)
            imageNode.state = ImageNodeState.ERROR
            imageNode.errorReason = err.message
            await this.imageNodeRepository.update(imageNode.id, imageNode)
          }
        }),
      )

      results.forEach((result) => {
        if (result.status === 'rejected') {
          this.logger.error(result.reason)
        }
      })
    } catch (err) {
      this.logger.error(err)
    }
  }

  async propagateImageToNode(internalImageName: string, node: Node) {
    const dockerRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()

    const imageApi = this.nodeApiFactory.createImageApi(node)

    let retries = 0
    while (retries < 10) {
      try {
        await imageApi.pullImage({
          image: internalImageName,
          registry: {
            url: dockerRegistry.url,
            username: dockerRegistry.username,
            password: dockerRegistry.password,
          },
        })
      } catch (err) {
        if (err.code !== 'ECONNRESET') {
          throw err
        }
      }
      retries++
      await new Promise((resolve) => setTimeout(resolve, retries * 1000))
    }
  }

  async handleImageNodeStatePullingImage(imageNode: ImageNode) {
    const node = await this.nodeRepository.findOneOrFail({
      where: {
        id: imageNode.nodeId,
      },
    })

    const imageApi = this.nodeApiFactory.createImageApi(node)
    const response = (await imageApi.imageExists(imageNode.imageRef)).data
    if (response.exists) {
      imageNode.state = ImageNodeState.READY
      await this.imageNodeRepository.save(imageNode)
      return
    }

    const timeoutMinutes = 60
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - imageNode.createdAt.getTime() > timeoutMs) {
      imageNode.state = ImageNodeState.ERROR
      imageNode.errorReason = 'Timeout while pulling image'
      await this.imageNodeRepository.save(imageNode)
      return
    }

    const retryTimeoutMinutes = 10
    const retryTimeoutMs = retryTimeoutMinutes * 60 * 1000
    if (Date.now() - imageNode.createdAt.getTime() > retryTimeoutMs) {
      await this.retryImageNodePull(imageNode)
      return
    }
  }

  async handleImageNodeStateBuildingImage(imageNode: ImageNode) {
    const node = await this.nodeRepository.findOneOrFail({
      where: {
        id: imageNode.nodeId,
      },
    })

    const nodeWorkspaceApi = this.nodeApiFactory.createImageApi(node)
    const response = (await nodeWorkspaceApi.imageExists(imageNode.imageRef)).data
    if (response && response.exists) {
      imageNode.state = ImageNodeState.READY
      await this.imageNodeRepository.save(imageNode)
      return
    }
  }

  @Cron(CronExpression.EVERY_10_SECONDS)
  async checkImageCleanup() {
    const lockKey = 'check-image-cleanup-lock'
    if (await this.redisLockProvider.lock(lockKey, 30)) {
      return
    }

    //  get all images
    const images = await this.imageRepository.find({
      where: {
        state: ImageState.REMOVING,
      },
    })

    await Promise.all(
      images.map(async (image) => {
        await this.imageNodeRepository.update(
          {
            imageRef: image.internalName,
          },
          {
            state: ImageNodeState.REMOVING,
          },
        )

        await this.imageRepository.remove(image)
      }),
    )

    await this.redisLockProvider.unlock(lockKey)
  }

  @Cron(CronExpression.EVERY_10_SECONDS)
  async checkImageState() {
    //  the first time the image is created it needs to be validated and pushed to the internal registry
    //  before propagating to the nodes
    //  this cron job will process the image states until the image is active (or error)

    //  get all images
    const images = await this.imageRepository.find({
      where: {
        state: Not(In([ImageState.ACTIVE, ImageState.ERROR])),
      },
    })

    await Promise.all(
      images.map(async (image) => {
        const lockKey = `check-image-state-lock-${image.id}`
        if (await this.redisLockProvider.lock(lockKey, 720)) {
          return
        }

        try {
          switch (image.state) {
            case ImageState.BUILD_PENDING:
              await this.handleImageTagStateBuildPending(image)
              break
            case ImageState.BUILDING:
              await this.handleImageTagStateBuilding(image)
              break
            case ImageState.PENDING:
              await this.handleImageTagStatePending(image)
              break
            case ImageState.PULLING_IMAGE:
              await this.handleImageTagStatePullingImage(image)
              break
            case ImageState.PENDING_VALIDATION:
              //  temp workaround to avoid race condition between api instances
              if (!(await this.dockerProvider.imageExists(image.name))) {
                await this.redisLockProvider.unlock(lockKey)
                return
              }

              await this.handleImageTagStatePendingValidation(image)
              break
            case ImageState.VALIDATING:
              await this.handleImageTagStateValidating(image)
              break
            case ImageState.REMOVING:
              await this.handleImageTagStateRemoving(image)
              break
          }
        } catch (error) {
          if (error.code === 'ECONNRESET') {
            await this.redisLockProvider.unlock(lockKey)
            this.checkImageState()
            return
          }

          const message = error.message || String(error)
          await this.updateImageState(image.id, ImageState.ERROR, message)
        }

        await this.redisLockProvider.unlock(lockKey)
      }),
    )
  }

  @Cron(CronExpression.EVERY_30_MINUTES, {
    name: 'cleanup-local-images',
  })
  async cleanupLocalImages() {
    await this.dockerProvider.imagePrune()
  }

  async removeImageFromNode(node: Node, imageNode: ImageNode) {
    if (imageNode && !imageNode.imageRef) {
      //  this should never happen
      this.logger.warn(`Internal image name not found for image node ${imageNode.id}`)
      return
    }

    const imageApi = this.nodeApiFactory.createImageApi(node)
    const imageExists = (await imageApi.imageExists(imageNode.imageRef)).data
    if (imageExists.exists) {
      await imageApi.removeImage(imageNode.imageRef)
    }

    imageNode.state = ImageNodeState.REMOVING
    await this.imageNodeRepository.save(imageNode)
  }

  async handleImageNodeStateRemovingImage(imageNode: ImageNode) {
    const node = await this.nodeRepository.findOne({
      where: {
        id: imageNode.nodeId,
      },
    })
    if (!node) {
      //  generally this should not happen
      //  in case the node has been deleted from the database, delete the image node record
      const errorMessage = `Node not found while trying to remove image ${imageNode.imageRef} from node ${imageNode.nodeId}`
      this.logger.warn(errorMessage)

      this.imageNodeRepository.delete(imageNode.id).catch((err) => {
        this.logger.error(fromAxiosError(err))
      })
      return
    }
    if (!imageNode.imageRef) {
      //  this should never happen
      //  remove the image node record (it will be recreated again by the image propagation job)
      this.logger.warn(`Internal image name not found for image node ${imageNode.id}`)
      this.imageNodeRepository.delete(imageNode.id).catch((err) => {
        this.logger.error(fromAxiosError(err))
      })
      return
    }

    const imageApi = this.nodeApiFactory.createImageApi(node)
    const response = await imageApi.imageExists(imageNode.imageRef)
    if (response.data && !response.data.exists) {
      await this.imageNodeRepository.delete(imageNode.id)
    } else {
      //  just in case the image is still there
      imageApi.removeImage(imageNode.imageRef).catch((err) => {
        //  this should not happen, and is not critical
        //  if the node can not remote the image, just delete the node record
        this.imageNodeRepository.delete(imageNode.id).catch((err) => {
          this.logger.error(fromAxiosError(err))
        })
        //  and log the error for tracking
        const errorMessage = `Failed to do just in case remove image ${imageNode.imageRef} from node ${node.id}: ${fromAxiosError(err)}`
        this.logger.warn(errorMessage)
      })
    }
  }

  async handleImageTagStateRemoving(image: Image) {
    const imageNodeItems = await this.imageNodeRepository.find({
      where: {
        imageRef: image.internalName,
      },
    })

    if (imageNodeItems.length === 0) {
      await this.imageRepository.remove(image)
    }
  }

  async handleImageTagStateBuildPending(image: Image) {
    await this.updateImageState(image.id, ImageState.BUILDING)
  }

  async handleImageTagStateBuilding(image: Image) {
    // Check if build has timed out
    const timeoutMinutes = 30
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - image.createdAt.getTime() > timeoutMs) {
      await this.updateImageState(image.id, ImageState.ERROR, 'Timeout while building image')
      return
    }

    // Get build info
    if (!image.buildInfo) {
      await this.updateImageState(image.id, ImageState.ERROR, 'Missing build information')
      return
    }

    try {
      // Find a node to build the image on
      const node = await this.nodeRepository.findOne({
        where: { state: NodeState.READY, unschedulable: Not(true) },
        order: { createdAt: 'ASC' },
      })

      // TODO: get only nodes where the base image is available (extract from buildInfo)

      if (!node) {
        // No ready nodes available, retry later
        return
      }

      const registry = await this.dockerRegistryService.getDefaultInternalRegistry()

      const nodeImageApi = this.nodeApiFactory.createImageApi(node)

      const tag = image.name.split(':')[1] // Tag existance had already been validated
      const imageIdWithTag = `${image.id}:${tag}`

      await nodeImageApi.buildImage({
        image: imageIdWithTag, // Name doesn't matter for runner, it uses the image ID when pushing to internal registry
        registry: {
          url: registry.url,
          username: registry.username,
          password: registry.password,
        },
        organizationId: image.organizationId,
        dockerfile: image.buildInfo.dockerfileContent,
        context: image.buildInfo.contextHashes,
        pushToInternalRegistry: true,
      })

      // save ImageNode

      const internalImageName = `${registry.url}/${registry.project}/${imageIdWithTag}`

      image.internalName = internalImageName
      await this.imageRepository.save(image)

      // Wait for 30 seconds because of Harbor's delay at making newly created images available
      await new Promise((resolve) => setTimeout(resolve, 30000))

      // Move to next state
      await this.updateImageState(image.id, ImageState.PENDING)
    } catch (err) {
      if (err.code === 'ECONNRESET') {
        // Connection reset, retry later
        return
      }

      this.logger.error(`Error building image ${image.name}: ${fromAxiosError(err)}`)
      await this.updateImageState(image.id, ImageState.ERROR, fromAxiosError(err).message)
    }
  }

  async handleImageTagStatePending(image: Image) {
    let dockerRegistry: DockerRegistry

    await this.updateImageState(image.id, ImageState.PULLING_IMAGE)

    let localImageName = image.name

    if (image.buildInfo) {
      //  get the default internal registry
      dockerRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
      localImageName = image.internalName
    } else {
      //  find docker registry based on image name and organization id
      dockerRegistry = await this.dockerRegistryService.findOneByImageName(image.name, image.organizationId)
    }

    // Use the dockerRegistry for pulling the image
    await this.dockerProvider.pullImage(localImageName, dockerRegistry)
  }

  async handleImageTagStatePullingImage(image: Image) {
    const localImageName = image.buildInfo ? image.internalName : image.name
    // Check timeout first
    const timeoutMinutes = 15
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - image.createdAt.getTime() > timeoutMs) {
      await this.updateImageState(image.id, ImageState.ERROR, 'Timeout while pulling image')
      return
    }

    const imageExists = await this.dockerProvider.imageExists(localImageName)
    if (!imageExists) {
      //  retry until the image exists (or eventually timeout)
      return
    }

    //  sleep for 30 seconds
    //  workaround for docker image not being ready, but exists
    await new Promise((resolve) => setTimeout(resolve, 30000))

    //  get the organization
    const organization = await this.organizationService.findOne(image.organizationId)
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${image.organizationId} not found`)
    }

    // Check image size
    const imageInfo = await this.dockerProvider.getImageInfo(localImageName)
    const MAX_SIZE_GB = organization.maxImageSize

    if (imageInfo.sizeGB > MAX_SIZE_GB) {
      await this.updateImageState(
        image.id,
        ImageState.ERROR,
        `Image size (${imageInfo.sizeGB.toFixed(2)}GB) exceeds maximum allowed size of ${MAX_SIZE_GB}GB`,
      )
      return
    }

    //  check if the organization has reached the image size quota
    const totalImageSizeUsed = await this.imageRepository.sum('size', {
      organizationId: image.organizationId,
    })
    if (totalImageSizeUsed + imageInfo.sizeGB > organization.totalImageSize) {
      await this.updateImageState(
        image.id,
        ImageState.ERROR,
        `Total image size quota (${organization.totalImageSize.toFixed(2)}GB) exceeded`,
      )
      return
    }

    image.size = imageInfo.sizeGB
    image.state = ImageState.PENDING_VALIDATION

    // Ensure entrypoint is set
    if (!image.entrypoint) {
      if (imageInfo.entrypoint) {
        if (Array.isArray(imageInfo.entrypoint)) {
          image.entrypoint = imageInfo.entrypoint
        } else {
          image.entrypoint = [imageInfo.entrypoint]
        }
      } else {
        image.entrypoint = ['sleep', 'infinity']
      }
    }

    await this.imageRepository.save(image)
  }

  async handleImageTagStatePendingValidation(image: Image) {
    try {
      await this.updateImageState(image.id, ImageState.VALIDATING)

      await this.validateImageRuntime(image.id)

      if (!image.buildInfo) {
        // Imanges that went through the build process are already in the internal registry
        await this.pushImageToInternalRegistry(image.id)
      }
      await this.propagateImageToNodes(image.internalName)
      await this.updateImageState(image.id, ImageState.ACTIVE)

      // Best effort removal of old image from transient registry
      const registry = await this.dockerRegistryService.findOneByImageName(image.name, image.organizationId)
      if (registry && registry.registryType === RegistryType.TRANSIENT) {
        try {
          await this.dockerRegistryService.removeImage(image.name, registry.id)
        } catch (error) {
          if (error.statusCode === 404) {
            //  image not found, just return
            return
          }
          this.logger.error('Failed to remove old image:', fromAxiosError(error))
        }
      }
    } catch (error) {
      // workaround when app nodes don't use a single docker host instance
      if (error.statusCode === 404 || error.message?.toLowerCase().includes('no such image')) {
        return
      }
      await this.updateImageState(image.id, ImageState.ERROR, error.message)
    }
  }

  async handleImageTagStateValidating(image: Image) {
    //  check the timeout
    const timeoutMinutes = 10
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - image.createdAt.getTime() > timeoutMs) {
      await this.updateImageState(image.id, ImageState.ERROR, 'Timeout while validating image')
      return
    }
  }

  async validateImageRuntime(imageId: string): Promise<void> {
    const image = await this.imageRepository.findOneOrFail({
      where: {
        id: imageId,
      },
    })

    let containerId: string | null = null

    try {
      const localImageName = image.buildInfo ? image.internalName : image.name

      // Create and start the container
      containerId = await this.dockerProvider.create(localImageName, image.entrypoint)

      // Wait for 1 minute while checking container state
      const startTime = Date.now()
      const checkDuration = 60 * 1000 // 1 minute in milliseconds

      while (Date.now() - startTime < checkDuration) {
        const isRunning = await this.dockerProvider.isRunning(containerId)
        if (!isRunning) {
          throw new Error('Container exited')
        }
        await new Promise((resolve) => setTimeout(resolve, 2000))
      }
    } catch (error) {
      this.logger.debug('Error validating image runtime:', error)
      throw error
    } finally {
      // Cleanup: Destroy the container
      if (containerId) {
        try {
          await this.dockerProvider.remove(containerId)
        } catch (cleanupError) {
          this.logger.error('Error cleaning up container:', fromAxiosError(cleanupError))
        }
      }
    }
  }

  async pushImageToInternalRegistry(imageId: string) {
    const image = await this.imageRepository.findOneOrFail({
      where: {
        id: imageId,
      },
    })

    const registry = await this.dockerRegistryService.getDefaultInternalRegistry()
    if (!registry) {
      throw new Error('No default internal registry configured')
    }

    //  get tag from image name
    const tag = image.name.split(':')[1]
    const internalImageName = `${registry.url}/${registry.project}/${image.id}:${tag}`

    image.internalName = internalImageName
    await this.imageRepository.save(image)

    // Tag the image with the internal registry name
    await this.dockerProvider.tagImage(image.name, internalImageName)

    // Push the newly tagged image
    await this.dockerProvider.pushImage(internalImageName, registry)
  }

  async retryImageNodePull(imageNode: ImageNode) {
    const node = await this.nodeRepository.findOneOrFail({
      where: {
        id: imageNode.nodeId,
      },
    })

    const imageApi = this.nodeApiFactory.createImageApi(node)

    const dockerRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
    //  await this.redis.setex(lockKey, 360, this.instanceId)

    await imageApi.pullImage({
      image: imageNode.imageRef,
      registry: {
        url: dockerRegistry.url,
        username: dockerRegistry.username,
        password: dockerRegistry.password,
      },
    })
  }

  private async updateImageState(imageId: string, state: ImageState, errorReason?: string) {
    const image = await this.imageRepository.findOneOrFail({
      where: {
        id: imageId,
      },
    })
    image.state = state
    if (errorReason) {
      image.errorReason = errorReason
    }
    await this.imageRepository.save(image)
  }

  @Cron(CronExpression.EVERY_HOUR)
  async cleanupOldBuildInfoImageNodes() {
    const lockKey = 'cleanup-old-buildinfo-images-lock'
    if (await this.redisLockProvider.lock(lockKey, 300)) {
      return
    }

    try {
      const oneDayAgo = new Date()
      oneDayAgo.setDate(oneDayAgo.getDate() - 1)

      // Find all BuildInfo entities that haven't been used in over a day
      const oldBuildInfos = await this.buildInfoRepository.find({
        where: {
          lastUsedAt: LessThan(oneDayAgo),
        },
      })

      if (oldBuildInfos.length === 0) {
        return
      }

      const imageRefs = oldBuildInfos.map((buildInfo) => buildInfo.imageRef)

      const result = await this.imageNodeRepository.update(
        { imageRef: In(imageRefs) },
        { state: ImageNodeState.REMOVING },
      )

      if (result.affected > 0) {
        this.logger.debug(`Marked ${result.affected} ImageNodes for removal due to unused BuildInfo`)
      }
    } catch (error) {
      this.logger.error(`Failed to mark old BuildInfo ImageNodes for removal: ${fromAxiosError(error)}`)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }
}
