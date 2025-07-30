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
    registry.url = updateDto.url
    registry.username = updateDto.username
    if (updateDto.password) {
      registry.password = updateDto.password
    }
    registry.project = updateDto.project

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

    if (registry.url.startsWith('localhost') || registry.url.startsWith('127.0.0.1')) {
      return `http://${registry.url}`
    }

    return registry.url.startsWith('http') ? registry.url : `https://${registry.url}`
  }

  /**
   * Finds a registry that matches the given image name
   * First tries database, then creates a temporary registry config for public access
   */
  private async findRegistryByImageName(imageName: string, organizationId?: string): Promise<DockerRegistry | null> {
    // Parse the image to extract potential registry hostname
    const imageParts = imageName.split('/')

    // Check if the first part looks like a registry hostname (contains . or :)
    const hasRegistryPrefix = imageParts.length > 1 && (imageParts[0].includes('.') || imageParts[0].includes(':'))

    if (hasRegistryPrefix) {
      // Image has registry prefix, try to find matching registry in database first
      const whereCondition = organizationId
        ? [{ organizationId }, { organizationId: IsNull() }]
        : [{ organizationId: IsNull() }]

      const registries = await this.dockerRegistryRepository.find({
        where: whereCondition,
      })

      // Try to find a registry that matches the image hostname
      for (const registry of registries) {
        const strippedUrl = registry.url.replace(/^(https?:\/\/)/, '')
        if (imageName.startsWith(strippedUrl)) {
          return registry
        }
      }

      // Not found in database, create temporary registry config for public access
      return this.createTemporaryRegistryConfig(imageParts[0])
    } else {
      // Image has no registry prefix (e.g., "alpine:3.21")
      // Create temporary Docker Hub config
      return this.createTemporaryRegistryConfig('docker.io')
    }
  }

  /**
   * Creates a temporary registry configuration for public access
   */
  private createTemporaryRegistryConfig(hostname: string): DockerRegistry {
    const registry = new DockerRegistry()
    registry.id = `temp-${hostname}`
    registry.name = `Temporary ${hostname}`
    registry.url = hostname === 'docker.io' ? 'https://registry-1.docker.io' : `https://${hostname}`
    registry.username = ''
    registry.password = ''
    registry.project = ''
    registry.isDefault = false
    registry.registryType = RegistryType.INTERNAL
    return registry
  }

  /**
   * Gets an anonymous token for Docker Hub
   */
  private async getDockerHubToken(repository: string): Promise<string | null> {
    try {
      const tokenUrl = `https://auth.docker.io/token?service=registry.docker.io&scope=repository:${repository}:pull`
      const response = await axios.get(tokenUrl, { timeout: 10000 })
      return response.data.token
    } catch (error) {
      this.logger.warn(`Failed to get Docker Hub token: ${error.message}`)
      return null
    }
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

  /**
   * Gets comprehensive image details including digest, size, and entrypoint without pulling the image
   * Automatically detects the registry from the image name
   */
  async getImageDetails(
    image: string,
    organizationId?: string,
  ): Promise<{
    digest: string
    sizeGB: number
    entrypoint: string[]
    cmd: string[]
    env: string[]
    workingDir?: string
    user?: string
  }> {
    try {
      // Extract tag
      const lastColonIndex = image.lastIndexOf(':')
      const fullPath = image.substring(0, lastColonIndex)
      const tag = image.substring(lastColonIndex + 1)

      // Find the registry for this image (tries database first, then creates temporary config)
      const registry = await this.findRegistryByImageName(image, organizationId)

      const registryUrl = this.getRegistryUrl(registry)

      // Remove registry prefix if present in the image name
      let repoPath: string

      // Extract hostname from registry URL for comparison
      const registryHost = registryUrl.replace(/^https?:\/\//, '')

      if (fullPath.startsWith(registryHost + '/')) {
        // Image includes registry hostname, strip it
        const projectAndRepo = fullPath.substring(registryHost.length + 1) // +1 for the slash
        // For Harbor format like: bbox-stage/backup-sandbox-75148d5a
        const parts = projectAndRepo.split('/')
        // Use project/repo directly as repoPath
        repoPath = projectAndRepo
      } else {
        // Image name without registry prefix, use as-is
        repoPath = fullPath

        // Special handling for Docker Hub - add library/ prefix for single-name images
        if (registry.url.includes('registry-1.docker.io') && !repoPath.includes('/')) {
          repoPath = `library/${repoPath}`
        }
      }

      // Get the manifest using GET request to retrieve full body
      const manifestUrl = `${registryUrl}/v2/${repoPath}/manifests/${tag}`

      // Build headers - handle different auth methods
      const headers: any = {
        Accept: 'application/vnd.docker.distribution.manifest.v2+json',
      }

      if (registry.username && registry.password) {
        // Use basic auth for configured registries
        const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')
        headers.Authorization = `Basic ${encodedCredentials}`
      } else if (registry.url.includes('registry-1.docker.io')) {
        // Get anonymous token for Docker Hub
        const dockerHubRepo = repoPath.includes('/') ? repoPath : `library/${repoPath}`
        const token = await this.getDockerHubToken(dockerHubRepo)
        if (token) {
          headers.Authorization = `Bearer ${token}`
        }
      }

      const manifestResponse = await axios({
        method: 'get',
        url: manifestUrl,
        headers,
        validateStatus: (status) => status < 500,
        timeout: 30000,
      })

      if (manifestResponse.status >= 300) {
        throw new Error(`Failed to get manifest for image ${image}: ${manifestResponse.statusText}`)
      }

      // Extract the digest from headers
      const digest = manifestResponse.headers['docker-content-digest']
      if (!digest) {
        throw new Error(`Docker content digest not found for image ${image}`)
      }

      let manifest = manifestResponse.data

      // Handle manifest lists (multi-platform images)
      if (
        manifest.mediaType === 'application/vnd.oci.image.index.v1+json' ||
        manifest.mediaType === 'application/vnd.docker.distribution.manifest.list.v2+json'
      ) {
        this.logger.debug(`Image ${image} is a manifest list, selecting platform-specific manifest`)

        // Find linux/amd64 platform (only architecture we support)
        const platformManifest = manifest.manifests?.find(
          (m: any) => m.platform?.architecture === 'amd64' && m.platform?.os === 'linux',
        )

        if (!platformManifest) {
          throw new Error(`No linux/amd64 platform found for image ${image}. Only amd64 architecture is supported.`)
        }

        // Fetch the actual platform-specific manifest
        const platformManifestUrl = `${registryUrl}/v2/${repoPath}/manifests/${platformManifest.digest}`
        const platformResponse = await axios({
          method: 'get',
          url: platformManifestUrl,
          headers,
          validateStatus: (status) => status < 500,
          timeout: 30000,
        })

        if (platformResponse.status >= 300) {
          throw new Error(`Failed to get platform manifest for image ${image}: ${platformResponse.statusText}`)
        }

        manifest = platformResponse.data
        this.logger.debug(`Successfully fetched platform-specific manifest for ${image}`)
      }

      // Calculate total size from all layers
      const totalSize = manifest.layers?.reduce((sum: number, layer: any) => sum + (layer.size || 0), 0) || 0
      const sizeGB = totalSize / (1024 * 1024 * 1024)

      // Get the config blob to extract entrypoint and other details
      const configDigest = manifest.config?.digest
      if (!configDigest) {
        // Return basic info if config is not available
        return {
          digest,
          sizeGB,
          entrypoint: [],
          cmd: [],
          env: [],
        }
      }

      const configUrl = `${registryUrl}/v2/${repoPath}/blobs/${configDigest}`

      // Build headers for config request - handle different auth methods
      const configHeaders: any = {}
      if (registry.username && registry.password) {
        // Use basic auth for configured registries
        const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')
        configHeaders.Authorization = `Basic ${encodedCredentials}`
      } else if (registry.url.includes('registry-1.docker.io')) {
        // Get anonymous token for Docker Hub
        const dockerHubRepo = repoPath.includes('/') ? repoPath : `library/${repoPath}`
        const token = await this.getDockerHubToken(dockerHubRepo)
        if (token) {
          configHeaders.Authorization = `Bearer ${token}`
        }
      }

      const configResponse = await axios({
        method: 'get',
        url: configUrl,
        headers: configHeaders,
        validateStatus: (status) => status < 500,
        timeout: 30000,
      })

      if (configResponse.status >= 300) {
        this.logger.warn(`Failed to get config blob for image ${image}: ${configResponse.statusText}`)
        // Return basic info without config details
        return {
          digest,
          sizeGB,
          entrypoint: [],
          cmd: [],
          env: [],
        }
      }

      const config = configResponse.data

      return {
        digest,
        sizeGB,
        entrypoint: config.config?.Entrypoint || [],
        cmd: config.config?.Cmd || [],
        env: config.config?.Env || [],
        workingDir: config.config?.WorkingDir,
        user: config.config?.User,
      }
    } catch (error) {
      this.logger.error(`Error getting image details for ${image}: ${error.message}`)
      throw new Error(`Failed to get image details for ${image}: ${error.message}`)
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
