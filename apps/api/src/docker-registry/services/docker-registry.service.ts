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
import { parseDockerImage } from '../../common/utils/docker-image.util'
import axios from 'axios'
import type { AxiosRequestHeaders } from 'axios'
import { AxiosHeaders } from 'axios'

const timeoutMs = 3000

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
  ) {}

  async create(
    createDto: CreateDockerRegistryDto,
    organizationId?: string,
    isFallback?: boolean,
  ): Promise<DockerRegistry> {
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
      isFallback,
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

  async getAvailableBackupRegistry(preferredRegion: string): Promise<DockerRegistry | null> {
    const registries = await this.dockerRegistryRepository.find({
      where: { registryType: RegistryType.BACKUP, isDefault: true },
    })

    if (registries.length === 0) {
      return null
    }

    // Filter registries by preferred region
    const preferredRegionRegistries = registries.filter((registry) => registry.region === preferredRegion)

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

  private async getDockerHubToken(repository: string): Promise<string | null> {
    try {
      const tokenUrl = `https://auth.docker.io/token?service=registry-1.docker.io&scope=repository:${repository}:pull`
      const response = await axios.get(tokenUrl, { timeout: 10000 })
      return response.data.token
    } catch (error) {
      this.logger.warn(`Failed to get Docker Hub token: ${error.message}`)
      return null
    }
  }

  async checkImageExistsInRegistry(imageName: string, registry: DockerRegistry): Promise<boolean> {
    try {
      const parsedImage = parseDockerImage(imageName)
      if (!parsedImage.project || !parsedImage.tag) {
        throw new Error('Invalid image name format. Expected: [registry]/project/repository:tag')
      }

      const registryUrl = this.getRegistryUrl(registry)
      const apiUrl = `${registryUrl}/v2/${parsedImage.project}/${parsedImage.repository}/manifests/${parsedImage.tag}`
      const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')

      const response = await axios({
        method: 'get',
        url: apiUrl,
        headers: {
          Authorization: `Basic ${encodedCredentials}`,
        },
        validateStatus: (status) => status < 500,
        timeout: timeoutMs,
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

  async getImageDetails(image: string, organizationId?: string): Promise<ImageDetails> {
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
      const acceptHeader = [
        'application/vnd.docker.distribution.manifest.v2+json',
        'application/vnd.docker.distribution.manifest.list.v2+json',
        'application/vnd.oci.image.index.v1+json',
        'application/vnd.oci.image.manifest.v1+json',
      ].join(', ')

      const baseHeaders = new AxiosHeaders()
      baseHeaders.set('Accept', acceptHeader)

      let bearerToken: string | null = null

      // Pre-populate auth if we already know how
      if (registry.username && registry.password) {
        const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')
        baseHeaders.set('Authorization', `Basic ${encodedCredentials}`)
      } else if (registry.url.includes('registry-1.docker.io')) {
        // Get anonymous token for Docker Hub
        const dockerHubRepo = repoPath.includes('/') ? repoPath : `library/${repoPath}`
        bearerToken = await this.getDockerHubToken(dockerHubRepo)
        if (bearerToken) {
          baseHeaders.set('Authorization', `Bearer ${bearerToken}`)
        }
      }

      const sendWithHeaders = async (url: string, headers: AxiosRequestHeaders | typeof baseHeaders) =>
        axios({ method: 'get', url, headers, validateStatus: (s) => s < 500, timeout: timeoutMs })

      let manifestResponse = await sendWithHeaders(manifestUrl, baseHeaders)

      // Handle Bearer challenge (e.g., AWS ECR Public)
      if (manifestResponse.status === 401 && manifestResponse.headers['www-authenticate']) {
        const authHeader = String(manifestResponse.headers['www-authenticate'])
        const challenge = parseWwwAuthenticate(authHeader)
        if (challenge?.scheme?.toLowerCase() === 'bearer' && challenge.realm) {
          try {
            const token = await fetchBearerToken(challenge, {
              repoPath,
              registryHost: registryHost,
            })
            if (token) {
              bearerToken = token
              const headersWithBearer = AxiosHeaders.from(baseHeaders)
              headersWithBearer.set('Authorization', `Bearer ${token}`)
              manifestResponse = await sendWithHeaders(manifestUrl, headersWithBearer)
            }
          } catch {
            // fall through to normal error handling
          }
        }
      }

      if (manifestResponse.status >= 300) {
        throw new Error(`Failed to get manifest for image ${image}: ${manifestResponse.statusText}`)
      }

      // Extract the digest from headers
      let digest = manifestResponse.headers['docker-content-digest']

      // If digest not in headers, calculate it from the manifest body
      if (!digest) {
        const crypto = require('crypto')
        const manifestStr = JSON.stringify(manifestResponse.data)
        digest = 'sha256:' + crypto.createHash('sha256').update(manifestStr).digest('hex')
      }

      let manifest = manifestResponse.data

      // Handle manifest lists (multi-platform images)
      if (
        manifest.mediaType === 'application/vnd.oci.image.index.v1+json' ||
        manifest.mediaType === 'application/vnd.docker.distribution.manifest.list.v2+json'
      ) {
        // Find linux/amd64 platform (only architecture we support)
        const platformManifest = manifest.manifests?.find(
          (m: any) => m.platform?.architecture === 'amd64' && m.platform?.os === 'linux',
        )

        if (!platformManifest) {
          throw new Error(`No linux/amd64 platform found for image ${image}. Only amd64 architecture is supported.`)
        }

        // Fetch the actual platform-specific manifest
        const platformManifestUrl = `${registryUrl}/v2/${repoPath}/manifests/${platformManifest.digest}`
        const platformHeaders = AxiosHeaders.from(baseHeaders)
        if (bearerToken) {
          platformHeaders.set('Authorization', `Bearer ${bearerToken}`)
        }
        const platformResponse = await axios({
          method: 'get',
          url: platformManifestUrl,
          headers: platformHeaders,
          validateStatus: (status) => status < 500,
          timeout: timeoutMs,
        })

        if (platformResponse.status >= 300) {
          throw new Error(`Failed to get platform manifest for image ${image}: ${platformResponse.statusText}`)
        }

        manifest = platformResponse.data
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

      // Build headers for config request - reuse Bearer token if available
      const configHeaders = new AxiosHeaders()
      if (bearerToken) {
        configHeaders.set('Authorization', `Bearer ${bearerToken}`)
      } else if (registry.username && registry.password) {
        const encodedCredentials = Buffer.from(`${registry.username}:${registry.password}`).toString('base64')
        configHeaders.set('Authorization', `Basic ${encodedCredentials}`)
      } else if (registry.url.includes('registry-1.docker.io')) {
        const dockerHubRepo = repoPath.includes('/') ? repoPath : `library/${repoPath}`
        const token = await this.getDockerHubToken(dockerHubRepo)
        if (token) {
          configHeaders.set('Authorization', `Bearer ${token}`)
        }
      }

      const configResponse = await axios({
        method: 'get',
        url: configUrl,
        headers: configHeaders,
        validateStatus: (status) => status < 500,
        timeout: timeoutMs,
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
        timeout: timeoutMs,
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
            timeout: timeoutMs,
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
            timeout: timeoutMs,
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
        timeout: timeoutMs,
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
        timeout: timeoutMs,
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

// Parses a WWW-Authenticate header for Bearer challenges
function parseWwwAuthenticate(header: string): {
  scheme: string
  realm?: string
  service?: string
  scope?: string
} | null {
  if (!header) return null
  const [schemePart, paramsPart] = header.split(/\s+/, 2)
  if (!schemePart) return null
  const scheme = schemePart.trim()
  const params: Record<string, string> = {}
  if (paramsPart) {
    for (const kv of paramsPart.split(',')) {
      const idx = kv.indexOf('=')
      if (idx > -1) {
        const key = kv.slice(0, idx).trim()
        let value = kv.slice(idx + 1).trim()
        if (value.startsWith('"') && value.endsWith('"')) {
          value = value.slice(1, -1)
        }
        params[key] = value
      }
    }
  }
  return { scheme, realm: params.realm, service: params.service, scope: params.scope }
}

// Fetches a Bearer token using the auth challenge parameters (works for Docker Registry and AWS ECR Public)
async function fetchBearerToken(
  challenge: { realm?: string; service?: string; scope?: string },
  ctx: { repoPath: string; registryHost: string },
): Promise<string | null> {
  if (!challenge.realm) return null
  const params = new URLSearchParams()
  if (challenge.service) params.set('service', challenge.service)
  // If scope not provided by challenge, construct a default repository:repo:pull
  const scope = challenge.scope || `repository:${ctx.repoPath}:pull`
  params.set('scope', scope)
  try {
    const url = `${challenge.realm}?${params.toString()}`
    const resp = await axios.get(url, { timeout: 10000, validateStatus: (s) => s < 500 })
    if (resp.status >= 300) return null
    return resp.data?.token || resp.data?.access_token || null
  } catch {
    return null
  }
}
