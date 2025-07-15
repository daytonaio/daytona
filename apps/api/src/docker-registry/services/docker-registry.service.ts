/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Inject, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { In, IsNull, Repository } from 'typeorm'
import { DockerRegistry } from '../entities/docker-registry.entity'
import { CreateDockerRegistryDto } from '../dto/create-docker-registry.dto'
import { UpdateDockerRegistryDto } from '../dto/update-docker-registry.dto'
import { ApiOAuth2 } from '@nestjs/swagger'
import { RegistryPushAccessDto } from '../../sandbox/dto/registry-push-access-dto'
import {
  DOCKER_REGISTRY_PROVIDER,
  IDockerRegistryProvider,
} from './../../docker-registry/providers/docker-registry.provider.interface'
import { RegistryType } from './../../docker-registry/enums/registry-type.enum'
import axios from 'axios'

@Injectable()
@ApiOAuth2(['openid', 'profile', 'email'])
export class DockerRegistryService {
  private readonly logger = new Logger(DockerRegistryService.name)

  constructor(
    @InjectRepository(DockerRegistry)
    private readonly dockerRegistryRepository: Repository<DockerRegistry>,
    @Inject(DOCKER_REGISTRY_PROVIDER)
    private readonly dockerRegistryProvider: IDockerRegistryProvider,
  ) {}

  async create(createDto: CreateDockerRegistryDto, organizationId?: string): Promise<DockerRegistry> {
    //  set some limit to the number of registries
    if (organizationId) {
      const registries = await this.dockerRegistryRepository.find({
        where: { organizationId },
      })
      if (registries.length >= 100) {
        throw new ForbiddenException('You have reached the maximum number of registries')
      }
    }

    const registry = this.dockerRegistryRepository.create({
      ...createDto,
      organizationId,
    })
    return this.dockerRegistryRepository.save(registry)
  }

  async findAll(organizationId: string): Promise<DockerRegistry[]> {
    return this.dockerRegistryRepository.find({
      where: { organizationId },
      order: {
        createdAt: 'DESC',
      },
    })
  }

  async findOne(registryId: string): Promise<DockerRegistry | null> {
    return this.dockerRegistryRepository.findOne({
      where: { id: registryId },
    })
  }

  async findOneOrFail(registryId: string): Promise<DockerRegistry> {
    return this.dockerRegistryRepository.findOneOrFail({
      where: { id: registryId },
    })
  }

  async update(registryId: string, updateDto: UpdateDockerRegistryDto): Promise<DockerRegistry> {
    const registry = await this.dockerRegistryRepository.findOne({
      where: { id: registryId },
    })

    if (!registry) {
      throw new NotFoundException(`Docker registry with ID ${registryId} not found`)
    }

    registry.name = updateDto.name
    registry.username = updateDto.username
    if (updateDto.password) {
      registry.password = updateDto.password
    }

    return this.dockerRegistryRepository.save(registry)
  }

  async remove(registryId: string): Promise<void> {
    const registry = await this.dockerRegistryRepository.findOne({
      where: { id: registryId },
    })

    if (!registry) {
      throw new NotFoundException(`Docker registry with ID ${registryId} not found`)
    }

    await this.dockerRegistryRepository.remove(registry)
  }

  async setDefault(registryId: string): Promise<DockerRegistry> {
    const registry = await this.dockerRegistryRepository.findOne({
      where: { id: registryId },
    })

    if (!registry) {
      throw new NotFoundException(`Docker registry with ID ${registryId} not found`)
    }

    await this.unsetDefaultRegistry()

    registry.isDefault = true
    return this.dockerRegistryRepository.save(registry)
  }

  private async unsetDefaultRegistry(): Promise<void> {
    await this.dockerRegistryRepository.update({ isDefault: true }, { isDefault: false })
  }

  async getDefaultInternalRegistry(): Promise<DockerRegistry | null> {
    return this.dockerRegistryRepository.findOne({
      where: { isDefault: true, registryType: RegistryType.INTERNAL },
    })
  }

  async getDefaultTransientRegistry(): Promise<DockerRegistry | null> {
    return this.dockerRegistryRepository.findOne({
      where: { isDefault: true, registryType: RegistryType.TRANSIENT },
    })
  }

  async findOneBySnapshotImageName(imageName: string, organizationId?: string): Promise<DockerRegistry | null> {
    const whereCondition = organizationId
      ? [
          { organizationId, registryType: In([RegistryType.INTERNAL, RegistryType.ORGANIZATION]) },
          { organizationId: IsNull(), registryType: In([RegistryType.INTERNAL, RegistryType.ORGANIZATION]) },
        ]
      : [{ organizationId: IsNull(), registryType: In([RegistryType.INTERNAL, RegistryType.ORGANIZATION]) }]

    const registries = await this.dockerRegistryRepository.find({
      where: whereCondition,
    })

    // Try to find a registry that matches the snapshot image name pattern
    for (const registry of registries) {
      const strippedUrl = registry.url.replace(/^(https?:\/\/)/, '')
      if (imageName.startsWith(strippedUrl)) {
        return registry
      }
    }

    return null
  }

  async getRegistryPushAccess(organizationId: string, userId: string): Promise<RegistryPushAccessDto> {
    const transientRegistry = await this.getDefaultTransientRegistry()
    if (!transientRegistry) {
      throw new Error('No default transient registry configured')
    }

    const uniqueId = crypto.randomUUID().replace(/-/g, '').slice(0, 12)
    const robotName = `temp-push-robot-${uniqueId}`
    const expiresAt = new Date()
    expiresAt.setHours(expiresAt.getHours() + 1) // Token valid for 1 hour

    const url = this.getRegistryUrl(transientRegistry) + '/api/v2.0/robots'

    try {
      const response = await this.dockerRegistryProvider.createRobotAccount(
        url,
        {
          username: transientRegistry.username,
          password: transientRegistry.password,
        },
        {
          name: robotName,
          description: `Temporary push access for user ${userId} in organization ${organizationId}`,
          duration: 3600,
          level: 'project',
          permissions: [
            {
              kind: 'project',
              namespace: transientRegistry.project,
              access: [{ resource: 'repository', action: 'push' }],
            },
          ],
        },
      )

      return {
        username: response.name,
        secret: response.secret,
        registryId: transientRegistry.id,
        registryUrl: new URL(url).host,
        project: transientRegistry.project,
        expiresAt: expiresAt.toISOString(),
      }
    } catch (error) {
      let errorMessage = `Failed to generate push token: ${error.message}`
      if (error.response) {
        errorMessage += ` - ${error.response.data.message || error.response.statusText}`
      }
      throw new Error(errorMessage)
    }
  }

  async removeImage(imageName: string, registryId: string): Promise<void> {
    const registry = await this.findOne(registryId)
    if (!registry) {
      throw new Error('Registry not found')
    }

    // Parse fully qualified image name
    // Example: harbor-test.internal.daytona.app/daytona/busybox:1.36.1
    const [nameWithTag, tag] = imageName.split(':')

    // Remove registry hostname if present
    const parts = nameWithTag.split('/')
    let project: string
    let repository: string

    if (parts.length >= 3 && parts[0].includes('.')) {
      // Format: hostname/project/repository
      project = parts[1]
      repository = parts.slice(2).join('/')
    } else if (parts.length === 2) {
      // Format: project/repository
      ;[project, repository] = parts
    } else {
      throw new Error('Invalid image name format. Expected: [registry]/project/repository[:tag]')
    }

    try {
      await this.dockerRegistryProvider.deleteArtifact(
        this.getRegistryUrl(registry),
        {
          username: registry.username,
          password: registry.password,
        },
        { project, repository, tag },
      )
    } catch (error) {
      const message = error.response?.data?.message || error.message
      throw new Error(`Failed to remove image ${imageName}: ${message}`)
    }
  }

  getRegistryUrl(registry: DockerRegistry): string {
    // Dev mode
    if (registry.url === 'registry:5000') {
      return 'http://registry:5000'
    }

    return registry.url.startsWith('http') ? registry.url : `https://${registry.url}`
  }

  /**
   * Checks if an image exists in the specified registry without pulling it
   */
  async checkImageExistsInRegistry(imageName: string, registry: DockerRegistry): Promise<boolean> {
    try {
      // extract tag
      const lastColonIndex = imageName.lastIndexOf(':')
      const fullPath = imageName.substring(0, lastColonIndex)
      const tag = imageName.substring(lastColonIndex + 1)

      const registryUrl = this.getRegistryUrl(registry)

      // Remove registry prefix if present in the image name
      let projectAndRepo = fullPath
      if (fullPath.startsWith(registryUrl)) {
        projectAndRepo = fullPath.substring(registryUrl.length + 1) // +1 for the slash
      }

      // For Harbor format like: harbor.host/bbox-stage/backup-sandbox-75148d5a
      const parts = projectAndRepo.split('/')

      const apiUrl = `${registryUrl}/v2/${parts[1]}/${parts[2]}/manifests/${tag}`
      const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')

      const response = await axios({
        method: 'get',
        url: apiUrl,
        headers: {
          Authorization: `Basic ${encodedCredentials}`,
        },
        validateStatus: (status) => status < 500,
        timeout: 30000,
      })

      if (response.status === 200) {
        this.logger.debug(`Image ${imageName} exists in registry`)
        return true
      }

      this.logger.debug(`Image ${imageName} does not exist in registry (status: ${response.status})`)
      return false
    } catch (error) {
      this.logger.error(`Error checking if image ${imageName} exists in registry: ${error.message}`)
      return false
    }
  }

  private async deleteRepositoryWithPrefix(
    repository: string,
    prefix: string,
    registry: DockerRegistry,
  ): Promise<void> {
    const registryUrl = this.getRegistryUrl(registry)
    const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')
    const repoPath = `${registry.project}/${prefix}${repository}`

    try {
      // Step 1: List all tags in the repository
      const tagsUrl = `${registryUrl}/v2/${repoPath}/tags/list`

      const tagsResponse = await axios({
        method: 'get',
        url: tagsUrl,
        headers: {
          Authorization: `Basic ${encodedCredentials}`,
        },
        validateStatus: (status) => status < 500,
        timeout: 30000,
      })

      if (tagsResponse.status === 404) {
        return
      }

      if (tagsResponse.status >= 300) {
        this.logger.error(`Error listing tags in repository ${repoPath}: ${tagsResponse.statusText}`)
        throw new Error(`Failed to list tags in repository ${repoPath}: ${tagsResponse.statusText}`)
      }

      const tags = tagsResponse.data.tags || []

      if (tags.length === 0) {
        this.logger.debug(`Repository ${repoPath} has no tags to delete`)
        return
      }

      // Step 2: Delete each tag
      for (const tag of tags) {
        try {
          // Get the digest for this tag
          const manifestUrl = `${registryUrl}/v2/${repoPath}/manifests/${tag}`

          const manifestResponse = await axios({
            method: 'head',
            url: manifestUrl,
            headers: {
              Authorization: `Basic ${encodedCredentials}`,
              Accept: 'application/vnd.docker.distribution.manifest.v2+json',
            },
            validateStatus: (status) => status < 500,
            timeout: 30000,
          })

          if (manifestResponse.status >= 300) {
            this.logger.warn(`Couldn't get manifest for tag ${tag}: ${manifestResponse.statusText}`)
            continue
          }

          const digest = manifestResponse.headers['docker-content-digest']
          if (!digest) {
            this.logger.warn(`Docker content digest not found for tag ${tag}`)
            continue
          }

          // Delete the manifest
          const deleteUrl = `${registryUrl}/v2/${repoPath}/manifests/${digest}`

          const deleteResponse = await axios({
            method: 'delete',
            url: deleteUrl,
            headers: {
              Authorization: `Basic ${encodedCredentials}`,
            },
            validateStatus: (status) => status < 500,
            timeout: 30000,
          })

          if (deleteResponse.status < 300) {
            this.logger.debug(`Deleted tag ${tag} from repository ${repoPath}`)
          } else {
            this.logger.warn(`Failed to delete tag ${tag}: ${deleteResponse.statusText}`)
          }
        } catch (error) {
          this.logger.warn(`Exception when deleting tag ${tag}: ${error.message}`)
          // Continue with other tags
        }
      }

      this.logger.debug(`Repository ${repoPath} cleanup completed`)
    } catch (error) {
      this.logger.error(`Exception when deleting repository ${repoPath}: ${error.message}`)
      throw error
    }
  }

  async deleteSandboxRepository(repository: string, registry: DockerRegistry): Promise<void> {
    try {
      // Delete both backup and snapshot repositories - necessary due to renaming
      await this.deleteRepositoryWithPrefix(repository, 'backup-', registry)
      await this.deleteRepositoryWithPrefix(repository, 'snapshot-', registry)
    } catch (error) {
      this.logger.error(`Failed to delete repositories for ${repository}: ${error.message}`)
      throw error
    }
  }

  async deleteBackupImageFromRegistry(imageName: string, registry: DockerRegistry): Promise<void> {
    // Extract tag
    const lastColonIndex = imageName.lastIndexOf(':')
    const fullPath = imageName.substring(0, lastColonIndex)
    const tag = imageName.substring(lastColonIndex + 1)

    const registryUrl = this.getRegistryUrl(registry)

    // Remove registry prefix if present in the image name
    let projectAndRepo = fullPath
    if (fullPath.startsWith(registryUrl)) {
      projectAndRepo = fullPath.substring(registryUrl.length + 1) // +1 for the slash
    }

    // For Harbor format like: harbor.host/bbox-stage/backup-sandbox-75148d5a
    const parts = projectAndRepo.split('/')

    // Construct repository path (everything after the registry host)
    const repoPath = parts.slice(1).join('/')

    // First, get the digest for the tag using the manifests endpoint
    const manifestUrl = `${registryUrl}/v2/${repoPath}/manifests/${tag}`
    const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')

    try {
      // Get the digest from the headers
      const manifestResponse = await axios({
        method: 'head', // Using HEAD request to only fetch headers
        url: manifestUrl,
        headers: {
          Authorization: `Basic ${encodedCredentials}`,
          Accept: 'application/vnd.docker.distribution.manifest.v2+json',
        },
        validateStatus: (status) => status < 500,
        timeout: 30000,
      })

      if (manifestResponse.status >= 300) {
        this.logger.error(`Error getting manifest for image ${imageName}: ${manifestResponse.statusText}`)
        throw new Error(`Failed to get manifest for image ${imageName}: ${manifestResponse.statusText}`)
      }

      // Extract the digest from headers
      const digest = manifestResponse.headers['docker-content-digest']
      if (!digest) {
        throw new Error(`Docker content digest not found for image ${imageName}`)
      }

      // Now delete the image using the digest
      const deleteUrl = `${registryUrl}/v2/${repoPath}/manifests/${digest}`

      const deleteResponse = await axios({
        method: 'delete',
        url: deleteUrl,
        headers: {
          Authorization: `Basic ${encodedCredentials}`,
        },
        validateStatus: (status) => status < 500,
        timeout: 30000,
      })

      if (deleteResponse.status < 300) {
        this.logger.debug(`Image ${imageName} removed from the registry`)
        return
      }

      this.logger.error(`Error removing image ${imageName} from registry: ${deleteResponse.statusText}`)
      throw new Error(`Failed to remove image ${imageName} from registry: ${deleteResponse.statusText}`)
    } catch (error) {
      this.logger.error(`Exception when deleting image ${imageName}: ${error.message}`)
      throw error
    }
  }
}
