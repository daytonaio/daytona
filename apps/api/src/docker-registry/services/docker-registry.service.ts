/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Inject, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { EntityManager, FindOptionsWhere, IsNull, Repository } from 'typeorm'
import { DockerRegistry } from '../entities/docker-registry.entity'
import { CreateDockerRegistryInternalDto } from '../dto/create-docker-registry-internal.dto'
import { UpdateDockerRegistryDto } from '../dto/update-docker-registry.dto'
import { ApiOAuth2 } from '@nestjs/swagger'
import { RegistryPushAccessDto } from '../../sandbox/dto/registry-push-access-dto'
import {
  DOCKER_REGISTRY_PROVIDER,
  IDockerRegistryProvider,
} from './../../docker-registry/providers/docker-registry.provider.interface'
import { RegistryType } from './../../docker-registry/enums/registry-type.enum'
import { parseDockerImage, checkDockerfileHasRegistryPrefix } from '../../common/utils/docker-image.util'
import axios from 'axios'
import { OnAsyncEvent } from '../../common/decorators/on-async-event.decorator'
import { RegionEvents } from '../../region/constants/region-events.constant'
import { RegionCreatedEvent } from '../../region/events/region-created.event'
import { RegionDeletedEvent } from '../../region/events/region-deleted.event'
import { RegionService } from '../../region/services/region.service'
import {
  RegionSnapshotManagerCredsRegeneratedEvent,
  RegionSnapshotManagerUpdatedEvent,
} from '../../region/events/region-snapshot-manager-creds.event'
import { OrganizationEvents } from '../../organization/constants/organization-events.constant'
import { OrganizationDeletedEvent } from '../../organization/events/organization-deleted.event'

const AXIOS_TIMEOUT_MS = 3000
const DOCKER_HUB_REGISTRY = 'registry-1.docker.io'
const DOCKER_HUB_URL = 'docker.io'

/**
 * Normalizes Docker Hub URLs to 'docker.io' for storage.
 * Empty URLs are assumed to be Docker Hub.
 */
function normalizeRegistryUrl(url: string): string {
  if (!url || url.trim() === '' || url.toLowerCase().includes('docker.io')) {
    return DOCKER_HUB_URL
  }
  // Strip trailing slashes for consistent matching
  return url.trim().replace(/\/+$/, '')
}

export interface ImageDetails {
  digest: string
  sizeGB: number
  entrypoint: string[]
  cmd: string[]
  env: string[]
  workingDir?: string
  user?: string
}

@Injectable()
@ApiOAuth2(['openid', 'profile', 'email'])
export class DockerRegistryService {
  private readonly logger = new Logger(DockerRegistryService.name)

  constructor(
    @InjectRepository(DockerRegistry)
    private readonly dockerRegistryRepository: Repository<DockerRegistry>,
    @Inject(DOCKER_REGISTRY_PROVIDER)
    private readonly dockerRegistryProvider: IDockerRegistryProvider,
    private readonly regionService: RegionService,
  ) {}

  async create(
    createDto: CreateDockerRegistryInternalDto,
    organizationId?: string,
    isFallback = false,
    entityManager?: EntityManager,
  ): Promise<DockerRegistry> {
    const repository = entityManager ? entityManager.getRepository(DockerRegistry) : this.dockerRegistryRepository

    //  set some limit to the number of registries
    if (organizationId) {
      const registries = await repository.find({
        where: { organizationId },
      })
      if (registries.length >= 100) {
        throw new ForbiddenException('You have reached the maximum number of registries')
      }
    }

    const registry = repository.create({
      ...createDto,
      url: normalizeRegistryUrl(createDto.url),
      region: createDto.regionId,
      organizationId,
      isFallback,
    })
    return repository.save(registry)
  }

  async findAll(organizationId: string, registryType: RegistryType): Promise<DockerRegistry[]> {
    return this.dockerRegistryRepository.find({
      where: { organizationId, registryType },
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
    registry.url = normalizeRegistryUrl(updateDto.url)
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

  /**
   * Returns an available internal registry for storing snapshots.
   *
   * If a snapshot manager _is_ configured for the region identified by the provided _regionId_, only an internal registry that matches the region snapshot manager can be returned.
   * If no matching internal registry is found, _null_ will be returned.
   *
   * If a snapshot manager _is not_ configured for the provided region, the default internal registry will be returned (if available).
   * If no default internal registry is found, _null_ will be returned.
   *
   * @param regionId - The ID of the region.
   */
  async getAvailableInternalRegistry(regionId: string): Promise<DockerRegistry | null> {
    const region = await this.regionService.findOne(regionId)
    if (!region) {
      return null
    }

    if (region.snapshotManagerUrl) {
      return this.dockerRegistryRepository.findOne({
        where: { region: regionId, registryType: RegistryType.INTERNAL },
      })
    }

    return this.dockerRegistryRepository.findOne({
      where: { isDefault: true, registryType: RegistryType.INTERNAL },
    })
  }

  /**
   * Returns an available transient registry for pushing snapshots.
   *
   * If a snapshot manager _is_ configured for the region identified by the provided _regionId_, only a transient registry that matches the region snapshot manager can be returned.
   * If no matching transient registry is found, _null_ will be returned.
   *
   * If a snapshot manager _is not_ configured for the provided region or no region is provided, the default transient registry will be returned (if available).
   * If no default transient registry is found, _null_ will be returned.
   *
   * @param regionId - (Optional) The ID of the region.
   */
  async getAvailableTransientRegistry(regionId?: string): Promise<DockerRegistry | null> {
    if (regionId) {
      const region = await this.regionService.findOne(regionId)
      if (!region) {
        return null
      }

      if (region.snapshotManagerUrl) {
        return this.dockerRegistryRepository.findOne({
          where: { region: regionId, registryType: RegistryType.TRANSIENT },
        })
      }
    }

    return this.dockerRegistryRepository.findOne({
      where: { isDefault: true, registryType: RegistryType.TRANSIENT },
    })
  }

  async getDefaultDockerHubRegistry(): Promise<DockerRegistry | null> {
    return this.dockerRegistryRepository.findOne({
      where: {
        organizationId: IsNull(),
        registryType: RegistryType.INTERNAL,
        url: DOCKER_HUB_URL,
        project: '',
      },
    })
  }

  /**
   * Returns an available backup registry for storing snapshots.
   *
   * If a snapshot manager _is_ configured for the region identified by the provided _preferredRegionId_, only a backup registry that matches the region snapshot manager can be returned.
   * If no matching backup registry is found, _null_ will be returned.
   *
   * If a snapshot manager _is not_ configured for the provided region, a backup registry in the preferred region will be returned (if available).
   * If no backup registry is found in the preferred region, a fallback backup registry will be returned (if available).
   * If no fallback backup registry is found, _null_ will be returned.
   *
   * @param preferredRegionId - The ID of the preferred region.
   */
  async getAvailableBackupRegistry(preferredRegionId: string): Promise<DockerRegistry | null> {
    const region = await this.regionService.findOne(preferredRegionId)
    if (!region) {
      return null
    }

    if (region.snapshotManagerUrl) {
      return this.dockerRegistryRepository.findOne({
        where: { region: preferredRegionId, registryType: RegistryType.BACKUP },
      })
    }

    const registries = await this.dockerRegistryRepository.find({
      where: { registryType: RegistryType.BACKUP, isDefault: true },
    })

    if (registries.length === 0) {
      return null
    }

    // Filter registries by preferred region
    const preferredRegionRegistries = registries.filter((registry) => registry.region === preferredRegionId)

    // If we have registries in the preferred region, randomly select one
    if (preferredRegionRegistries.length > 0) {
      const randomIndex = Math.floor(Math.random() * preferredRegionRegistries.length)
      return preferredRegionRegistries[randomIndex]
    }

    // If no registry found in preferred region, try to find a fallback registry
    const fallbackRegistries = registries.filter((registry) => registry.isFallback === true)

    if (fallbackRegistries.length > 0) {
      const randomIndex = Math.floor(Math.random() * fallbackRegistries.length)
      return fallbackRegistries[randomIndex]
    }

    // If no fallback registry found either, throw an error
    throw new Error('No backup registry available')
  }

  /**
   * Returns an internal registry that matches the snapshot ref.
   *
   * If no matching internal registry is found, _null_ will be returned.
   *
   * @param ref - The snapshot ref.
   * @param regionId - The ID of the region which needs access to the internal registry.
   */
  async findInternalRegistryBySnapshotRef(ref: string, regionId: string): Promise<DockerRegistry | null> {
    const region = await this.regionService.findOne(regionId)
    if (!region) {
      return null
    }

    let registries: DockerRegistry[]

    if (region.snapshotManagerUrl) {
      registries = await this.dockerRegistryRepository.find({
        where: {
          region: regionId,
          registryType: RegistryType.INTERNAL,
        },
      })
    } else {
      registries = await this.dockerRegistryRepository.find({
        where: {
          organizationId: IsNull(),
          registryType: RegistryType.INTERNAL,
        },
      })
    }

    return this.findRegistryByUrlMatch(registries, ref)
  }

  /**
   * Returns a source registry that matches the snapshot image name and can be used to pull the image.
   *
   * If no matching source registry is found, _null_ will be returned.
   *
   * @param imageName - The user-provided image.
   * @param regionId - The ID of the region which needs access to the source registry.
   */
  async findSourceRegistryBySnapshotImageName(
    imageName: string,
    regionId: string,
    organizationId?: string,
  ): Promise<DockerRegistry | null> {
    const region = await this.regionService.findOne(regionId)
    if (!region) {
      return null
    }

    const whereCondition: FindOptionsWhere<DockerRegistry>[] = []

    if (region.organizationId) {
      // registries manually added by the organization
      whereCondition.push({
        organizationId: region.organizationId,
        registryType: RegistryType.ORGANIZATION,
      })
    }

    if (organizationId) {
      whereCondition.push({
        organizationId: organizationId,
        registryType: RegistryType.ORGANIZATION,
      })
    }

    if (region.snapshotManagerUrl) {
      // internal registry associated with region snapshot manager
      whereCondition.push({
        region: regionId,
        registryType: RegistryType.INTERNAL,
      })
    } else {
      // shared internal registries
      whereCondition.push({
        organizationId: IsNull(),
        registryType: RegistryType.INTERNAL,
      })
    }

    const registries = await this.dockerRegistryRepository.find({
      where: whereCondition,
    })

    // Prioritize ORGANIZATION registries over others
    // This ensures user-configured credentials take precedence over shared internal ones
    const priority: Partial<Record<RegistryType, number>> = {
      [RegistryType.ORGANIZATION]: 0,
    }
    const sortedRegistries = [...registries].sort(
      (a, b) => (priority[a.registryType] ?? 1) - (priority[b.registryType] ?? 1),
    )

    return this.findRegistryByUrlMatch(sortedRegistries, imageName)
  }

  /**
   * Returns a transient registry that matches the snapshot image name.
   *
   * If no matching transient registry is found, _null_ will be returned.
   *
   * @param imageName - The user-provided image.
   * @param regionId - The ID of the region which needs access to the transient registry.
   */
  async findTransientRegistryBySnapshotImageName(imageName: string, regionId: string): Promise<DockerRegistry | null> {
    const region = await this.regionService.findOne(regionId)
    if (!region) {
      return null
    }

    let registries: DockerRegistry[]

    if (region.snapshotManagerUrl) {
      registries = await this.dockerRegistryRepository.find({
        where: {
          region: regionId,
          registryType: RegistryType.TRANSIENT,
        },
      })
    } else {
      registries = await this.dockerRegistryRepository.find({
        where: {
          organizationId: IsNull(),
          registryType: RegistryType.TRANSIENT,
        },
      })
    }

    return this.findRegistryByUrlMatch(registries, imageName)
  }

  async getRegistryPushAccess(
    organizationId: string,
    userId: string,
    regionId?: string,
  ): Promise<RegistryPushAccessDto> {
    const transientRegistry = await this.getAvailableTransientRegistry(regionId)
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

    const parsedImage = parseDockerImage(imageName)
    if (!parsedImage.project) {
      throw new Error('Invalid image name format. Expected: [registry]/project/repository[:tag]')
    }

    try {
      await this.dockerRegistryProvider.deleteArtifact(
        this.getRegistryUrl(registry),
        {
          username: registry.username,
          password: registry.password,
        },
        {
          project: parsedImage.project,
          repository: parsedImage.repository,
          tag: parsedImage.tag,
        },
      )
    } catch (error) {
      const message = error.response?.data?.message || error.message
      throw new Error(`Failed to remove image ${imageName}: ${message}`)
    }
  }

  getRegistryUrl(registry: DockerRegistry): string {
    // Dev mode
    if (registry.url.startsWith('localhost:') || registry.url.startsWith('registry:')) {
      return `http://${registry.url}`
    }

    if (registry.url.startsWith('localhost') || registry.url.startsWith('127.0.0.1')) {
      return `http://${registry.url}`
    }

    return registry.url.startsWith('http') ? registry.url : `https://${registry.url}`
  }

  public async findRegistryByImageName(
    imageName: string,
    regionId: string,
    organizationId?: string,
  ): Promise<DockerRegistry | null> {
    // Parse the image to extract potential registry hostname
    const parsedImage = parseDockerImage(imageName)

    if (parsedImage.registry) {
      // Image has registry prefix, try to find matching registry in database first
      const registry = await this.findSourceRegistryBySnapshotImageName(imageName, regionId, organizationId)
      if (registry) {
        return registry
      }
      // Not found in database, create temporary registry config for public access
      return this.createTemporaryRegistryConfig(parsedImage.registry)
    } else {
      // Image has no registry prefix (e.g., "alpine:3.21")
      // Create temporary Docker Hub config
      return this.createTemporaryRegistryConfig('docker.io')
    }
  }

  /**
   * Finds a registry with a URL that matches the start of the target string.
   *
   * @param registries - The list of registries to search.
   * @param targetString - The string to match against registry URLs.
   * @returns The matching registry, or null if no match is found.
   */
  private findRegistryByUrlMatch(registries: DockerRegistry[], targetString: string): DockerRegistry | null {
    for (const registry of registries) {
      const strippedUrl = registry.url.replace(/^(https?:\/\/)/, '').replace(/\/+$/, '')
      if (targetString.startsWith(strippedUrl)) {
        // Ensure match is at a proper boundary (followed by '/', ':', or end-of-string)
        // to prevent "registry.depot.dev" from matching "registry.depot.dev-evil.com/..."
        const nextChar = targetString[strippedUrl.length]
        if (nextChar === undefined || nextChar === '/' || nextChar === ':') {
          return registry
        }
      }
    }
    return null
  }

  private createTemporaryRegistryConfig(registryOrigin: string): DockerRegistry {
    const registry = new DockerRegistry()
    registry.id = `temp-${registryOrigin}`
    registry.name = `Temporary ${registryOrigin}`
    registryOrigin = registryOrigin.replace(/^(https?:\/\/)/, '')
    registry.url = `https://${registryOrigin}`
    registry.username = ''
    registry.password = ''
    registry.project = ''
    registry.isDefault = false
    registry.registryType = RegistryType.INTERNAL
    return registry
  }

  private async getDockerHubToken(repository: string): Promise<string | null> {
    try {
      const tokenUrl = `https://auth.docker.io/token?service=${DOCKER_HUB_REGISTRY}&scope=repository:${repository}:pull`
      const response = await axios.get(tokenUrl, { timeout: 10000 })
      return response.data.token
    } catch (error) {
      this.logger.warn(`Failed to get Docker Hub token: ${error.message}`)
      return null
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
        timeout: AXIOS_TIMEOUT_MS,
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

      if (tags.length > 500) {
        this.logger.warn(`Repository ${repoPath} has more than 500 tags, skipping cleanup`)
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
            timeout: AXIOS_TIMEOUT_MS,
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
            timeout: AXIOS_TIMEOUT_MS,
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
    const parsedImage = parseDockerImage(imageName)
    if (!parsedImage.project || !parsedImage.tag) {
      throw new Error('Invalid image name format. Expected: [registry]/project/repository:tag')
    }

    const registryUrl = this.getRegistryUrl(registry)
    const repoPath = `${parsedImage.project}/${parsedImage.repository}`

    // First, get the digest for the tag using the manifests endpoint
    const manifestUrl = `${registryUrl}/v2/${repoPath}/manifests/${parsedImage.tag}`
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
        timeout: AXIOS_TIMEOUT_MS,
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
        timeout: AXIOS_TIMEOUT_MS,
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

  /**
   * Gets source registries for building a Docker image from a Dockerfile
   * If the Dockerfile has images with registry prefixes, returns all user registries
   *
   * @param dockerfileContent - The Dockerfile content
   * @param organizationId - The organization ID
   * @returns Array of source registries (private registries + default Docker Hub)
   */
  async getSourceRegistriesForDockerfile(dockerfileContent: string, organizationId: string): Promise<DockerRegistry[]> {
    const sourceRegistries: DockerRegistry[] = []

    // Check if Dockerfile has any images with a registry prefix
    // If so, include all user's registries (we can't reliably match specific registries)
    if (checkDockerfileHasRegistryPrefix(dockerfileContent)) {
      const userRegistries = await this.findAll(organizationId, RegistryType.ORGANIZATION)
      sourceRegistries.push(...userRegistries)
    }

    // Add default Docker Hub registry only if user doesn't have their own Docker Hub credentials
    // The auth configs map is keyed by URL, so adding the default last would override user credentials
    const userHasDockerHubCreds = sourceRegistries.some((registry) => registry.url.includes('docker.io'))

    if (!userHasDockerHubCreds) {
      const defaultDockerHubRegistry = await this.getDefaultDockerHubRegistry()
      if (defaultDockerHubRegistry) {
        sourceRegistries.push(defaultDockerHubRegistry)
      }
    }

    return sourceRegistries
  }

  @OnAsyncEvent({
    event: RegionEvents.SNAPSHOT_MANAGER_CREDENTIALS_REGENERATED,
  })
  private async _handleRegionSnapshotManagerCredsRegenerated(
    payload: RegionSnapshotManagerCredsRegeneratedEvent,
  ): Promise<void> {
    const { regionId, snapshotManagerUrl, username, password, entityManager } = payload

    const em = entityManager ?? this.dockerRegistryRepository.manager

    const registries = await em.count(DockerRegistry, {
      where: { region: regionId, url: snapshotManagerUrl },
    })

    if (registries === 0) {
      throw new NotFoundException(`No registries found for region ${regionId} with URL ${snapshotManagerUrl}`)
    }

    await em.update(DockerRegistry, { region: regionId, url: snapshotManagerUrl }, { username, password })
  }

  @OnAsyncEvent({
    event: RegionEvents.SNAPSHOT_MANAGER_UPDATED,
  })
  private async _handleRegionSnapshotManagerUpdated(payload: RegionSnapshotManagerUpdatedEvent): Promise<void> {
    const {
      region,
      organizationId,
      snapshotManagerUrl,
      prevSnapshotManagerUrl,
      entityManager,
      newUsername,
      newPassword,
    } = payload

    const em = entityManager ?? this.dockerRegistryRepository.manager

    if (prevSnapshotManagerUrl) {
      // Update old registries associated with previous snapshot manager URL
      if (snapshotManagerUrl) {
        await em.update(
          DockerRegistry,
          {
            region: region.id,
            url: prevSnapshotManagerUrl,
          },
          {
            url: snapshotManagerUrl,
            username: newUsername,
            password: newPassword,
          },
        )
      } else {
        // If snapshot manager URL is removed, delete associated registries
        await em.delete(DockerRegistry, {
          region: region.id,
          url: prevSnapshotManagerUrl,
        })
      }

      return
    }

    const registries = await em.count(DockerRegistry, {
      where: { region: region.id, url: snapshotManagerUrl },
    })

    if (registries === 0) {
      await this._handleRegionCreatedEvent(
        new RegionCreatedEvent(entityManager, region, organizationId, newUsername, newPassword),
      )
    }
  }

  @OnAsyncEvent({
    event: RegionEvents.CREATED,
  })
  private async _handleRegionCreatedEvent(payload: RegionCreatedEvent): Promise<void> {
    const { entityManager, region, organizationId, snapshotManagerUsername, snapshotManagerPassword } = payload

    if (!region.snapshotManagerUrl || !snapshotManagerUsername || !snapshotManagerPassword) {
      return
    }

    await this.create(
      {
        name: `${region.name}-backup`,
        url: region.snapshotManagerUrl,
        username: snapshotManagerUsername,
        password: snapshotManagerPassword,
        registryType: RegistryType.BACKUP,
        regionId: region.id,
      },
      organizationId ?? undefined,
      false,
      entityManager,
    )

    await this.create(
      {
        name: `${region.name}-internal`,
        url: region.snapshotManagerUrl,
        username: snapshotManagerUsername,
        password: snapshotManagerPassword,
        registryType: RegistryType.INTERNAL,
        regionId: region.id,
      },
      organizationId ?? undefined,
      false,
      entityManager,
    )

    await this.create(
      {
        name: `${region.name}-transient`,
        url: region.snapshotManagerUrl,
        username: snapshotManagerUsername,
        password: snapshotManagerPassword,
        registryType: RegistryType.TRANSIENT,
        regionId: region.id,
      },
      organizationId ?? undefined,
      false,
      entityManager,
    )
  }

  @OnAsyncEvent({
    event: RegionEvents.DELETED,
  })
  async handleRegionDeletedEvent(payload: RegionDeletedEvent): Promise<void> {
    const { entityManager, region } = payload

    if (!region.snapshotManagerUrl) {
      return
    }

    const repository = entityManager.getRepository(DockerRegistry)
    await repository.delete({ region: region.id })
  }

  @OnAsyncEvent({
    event: OrganizationEvents.DELETED,
  })
  async handleOrganizationDeletedEvent(payload: OrganizationDeletedEvent): Promise<void> {
    const { entityManager, organizationId } = payload

    await entityManager.delete(DockerRegistry, { organizationId, registryType: RegistryType.ORGANIZATION })
  }
}
