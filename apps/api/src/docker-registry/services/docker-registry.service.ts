/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Inject, Injectable, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { FindOptionsWhere, In, IsNull, Repository } from 'typeorm'
import { DockerRegistry } from '../entities/docker-registry.entity'
import { CreateDockerRegistryInternalDto } from '../dto/create-docker-registry.internal.dto'
import { UpdateDockerRegistryInternalDto } from '../dto/update-docker-registry.internal.dto'
import { ApiOAuth2 } from '@nestjs/swagger'
import { RegistryPushAccessDto } from '../../sandbox/dto/registry-push-access-dto'
import {
  DOCKER_REGISTRY_PROVIDER,
  IDockerRegistryProvider,
} from './../../docker-registry/providers/docker-registry.provider.interface'
import { RegistryType } from './../../docker-registry/enums/registry-type.enum'
import { OrganizationService } from '../../organization/services/organization.service'
import { Organization } from '../../organization/entities/organization.entity'
import { RegionService } from '../../region/services/region.service'

@Injectable()
@ApiOAuth2(['openid', 'profile', 'email'])
export class DockerRegistryService {
  constructor(
    @InjectRepository(DockerRegistry)
    private readonly dockerRegistryRepository: Repository<DockerRegistry>,
    @Inject(DOCKER_REGISTRY_PROVIDER)
    private readonly dockerRegistryProvider: IDockerRegistryProvider,
    @Inject(OrganizationService)
    private readonly organizationService: OrganizationService,
    private readonly regionService: RegionService,
  ) {}

  async create(createDto: CreateDockerRegistryInternalDto, organizationId?: string): Promise<DockerRegistry> {
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

  async findAll(organizationId?: string, regionName?: string): Promise<DockerRegistry[]> {
    if (organizationId && regionName) {
      return this.findAllByRegionName(organizationId, regionName)
    } else if (organizationId) {
      return this.findAllByOrganizationId(organizationId)
    } else {
      return this.dockerRegistryRepository.find({
        order: {
          createdAt: 'DESC',
        },
      })
    }
  }

  async findAllByRegionName(organizationId: string, regionName: string): Promise<DockerRegistry[]> {
    const region = await this.regionService.findOneByNameAndOrganization(regionName, organizationId)

    if (!region) {
      throw new NotFoundException('Region not found')
    }

    return this.dockerRegistryRepository.find({
      where: {
        regionId: region.id,
      },
      order: {
        createdAt: 'DESC',
      },
    })
  }

  async findAllByOrganizationId(organizationId: string): Promise<DockerRegistry[]> {
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

  async getOrganizationId(registryId: string): Promise<string | null> {
    const registry = await this.dockerRegistryRepository.findOne({
      where: { id: registryId },
      select: ['organizationId'],
    })

    if (!registry) {
      throw new NotFoundException(`Docker registry with ID ${registryId} not found`)
    }

    return registry.organizationId
  }

  async update(registryId: string, updateDto: UpdateDockerRegistryInternalDto): Promise<DockerRegistry> {
    const registry = await this.dockerRegistryRepository.findOne({
      where: { id: registryId },
    })

    if (!registry) {
      throw new NotFoundException(`Docker registry with ID ${registryId} not found`)
    }

    if (updateDto.name) {
      registry.name = updateDto.name
    }
    if (updateDto.url) {
      registry.url = updateDto.url
    }
    if (updateDto.username) {
      registry.username = updateDto.username
    }
    if (updateDto.password) {
      registry.password = updateDto.password
    }
    if (updateDto.project) {
      registry.project = updateDto.project
    }
    if (updateDto.isActive) {
      registry.isActive = updateDto.isActive
    }
    if (updateDto.isFallback) {
      registry.isFallback = updateDto.isFallback
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

  /**
   * If `organizationIdOrEntity` is not provided, an available *shared* snapshot registry is returned (if exists).
   *
   * If `organizationIdOrEntity` is provided and shared infrastructure is blocked for the organization, an available organization snapshot registry is returned (if exists).
   *
   * If shared infrastructure is not blocked for the organization, an available *shared* snapshot registry is returned (if exists) as a fallback if no organization snapshot registry exists.
   */
  async getAvailableSnapshotRegistry(organizationIdOrEntity?: string | Organization): Promise<DockerRegistry | null> {
    const baseFindOptions: FindOptionsWhere<DockerRegistry> = {
      isActive: true,
      registryType: RegistryType.SNAPSHOT,
    }

    const findAvailableRegistries = (organizationId?: string) => {
      return this.dockerRegistryRepository.find({
        where: { ...baseFindOptions, organizationId: organizationId ?? IsNull() },
      })
    }

    const getRandomRegistry = (registries: DockerRegistry[]): DockerRegistry | null => {
      if (registries.length > 0) {
        const randomIndex = Math.floor(Math.random() * registries.length)
        return registries[randomIndex]
      }
      return null
    }

    if (!organizationIdOrEntity) {
      const sharedRegistries = await findAvailableRegistries()
      return getRandomRegistry(sharedRegistries)
    }

    const organizationId =
      organizationIdOrEntity === undefined
        ? undefined
        : typeof organizationIdOrEntity === 'string'
          ? organizationIdOrEntity
          : organizationIdOrEntity.id

    const organization =
      organizationIdOrEntity === undefined
        ? undefined
        : typeof organizationIdOrEntity === 'string'
          ? await this.organizationService.findOne(organizationIdOrEntity)
          : organizationIdOrEntity

    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    // Prefer organization registries
    const orgRegistries = await findAvailableRegistries(organizationId)
    const orgRegistry = getRandomRegistry(orgRegistries)
    if (orgRegistry) {
      return orgRegistry
    }

    if (!organization.blockSharedInfrastructure) {
      const sharedRegistries = await findAvailableRegistries()
      return getRandomRegistry(sharedRegistries)
    }

    return null
  }

  /**
   * Gets an available transient registry (if exists).
   *
   * Note: Transient registries are considered *shared* infrastructure.
   */
  async getAvailableTransientRegistry(): Promise<DockerRegistry | null> {
    const baseFindOptions: FindOptionsWhere<DockerRegistry> = {
      isActive: true,
      registryType: RegistryType.TRANSIENT,
    }

    const registries = await this.dockerRegistryRepository.find({
      where: baseFindOptions,
    })

    if (registries.length > 0) {
      const randomIndex = Math.floor(Math.random() * registries.length)
      return registries[randomIndex]
    }

    return null
  }

  /**
   * Gets the default backup registry for an organization, with preference for a specific region.
   *
   * The selection logic follows this priority order:
   * 1. Organization registry in the preferred region
   * 2. Organization registry with an unset region (not reserved for a specific region)
   * 3. Global fallback registry (if organization allows shared infrastructure)
   */
  async getAvailableBackupRegistry(
    preferredRegionId: string,
    organizationIdOrEntity?: string | Organization,
  ): Promise<DockerRegistry | null> {
    const baseFindOptions: FindOptionsWhere<DockerRegistry> = {
      isActive: true,
      registryType: RegistryType.BACKUP,
    }

    const findAvailableRegistries = (organizationId?: string, isFallback?: boolean) => {
      return this.dockerRegistryRepository.find({
        where: {
          ...baseFindOptions,
          organizationId: organizationId ?? IsNull(),
          ...(isFallback !== undefined ? { isFallback } : {}),
        },
      })
    }

    const getRandomRegistry = (registries: DockerRegistry[]): DockerRegistry | null => {
      if (registries.length > 0) {
        const randomIndex = Math.floor(Math.random() * registries.length)
        return registries[randomIndex]
      }
      return null
    }

    if (!organizationIdOrEntity) {
      const sharedRegistries = await findAvailableRegistries()
      return getRandomRegistry(sharedRegistries)
    }

    const organizationId =
      organizationIdOrEntity === undefined
        ? undefined
        : typeof organizationIdOrEntity === 'string'
          ? organizationIdOrEntity
          : organizationIdOrEntity.id

    const organization =
      organizationIdOrEntity === undefined
        ? undefined
        : typeof organizationIdOrEntity === 'string'
          ? await this.organizationService.findOne(organizationIdOrEntity)
          : organizationIdOrEntity

    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    const orgRegistries = await findAvailableRegistries(organizationId)

    const preferredRegionRegistry = getRandomRegistry(orgRegistries.filter((r) => r.regionId === preferredRegionId))
    if (preferredRegionRegistry) {
      return preferredRegionRegistry
    }

    // Fallback to organization registries with an unset region (not reserved for a specific region)
    const fallbackOrgRegistry = getRandomRegistry(orgRegistries.filter((r) => !r.regionId))
    if (fallbackOrgRegistry) {
      return fallbackOrgRegistry
    }

    if (!organization.blockSharedInfrastructure) {
      // Find shared registry marked as fallback
      const sharedRegistries = await findAvailableRegistries(undefined, true)
      return getRandomRegistry(sharedRegistries)
    }

    return null
  }

  /**
   * Finds a registry by matching the image name or internal name against registry URLs.
   *
   * If `organizationId` is not provided, only *shared* registries are searched.
   *
   * If `organizationId` is provided and shared infrastructure is blocked for the organization,
   * only organization registries are searched.
   *
   * If shared infrastructure is not blocked for the organization, organization registries
   * are searched first, with *shared* registries as a fallback if no match is found.
   */
  private async findOneBySnapshotImageNameOrInternalName(
    imageNameOrInternalName: string,
    registryType: RegistryType | RegistryType[],
    organizationId?: string,
  ): Promise<DockerRegistry | null> {
    const registryTypes = Array.isArray(registryType) ? registryType : [registryType]

    const baseFindOptions: FindOptionsWhere<DockerRegistry> = {
      registryType: In(registryTypes),
    }

    if (!organizationId) {
      // Search only in shared registries
      const registries = await this.dockerRegistryRepository.find({
        where: {
          ...baseFindOptions,
          organizationId: IsNull(),
        },
      })

      return this.findMatchingRegistry(registries, imageNameOrInternalName)
    }

    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    // First, try to find in organization registries
    const orgRegistries = await this.dockerRegistryRepository.find({
      where: {
        ...baseFindOptions,
        organizationId,
      },
    })

    const orgMatch = this.findMatchingRegistry(orgRegistries, imageNameOrInternalName)
    if (orgMatch) {
      return orgMatch
    }

    if (!organization.blockSharedInfrastructure) {
      // Fall back to shared registries
      const sharedRegistries = await this.dockerRegistryRepository.find({
        where: {
          ...baseFindOptions,
          organizationId: IsNull(),
        },
      })

      return this.findMatchingRegistry(sharedRegistries, imageNameOrInternalName)
    }

    return null
  }

  private findMatchingRegistry(registries: DockerRegistry[], imageNameOrInternalName: string): DockerRegistry | null {
    for (const registry of registries) {
      const strippedUrl = registry.url.replace(/^(https?:\/\/)/, '')
      if (imageNameOrInternalName.startsWith(strippedUrl)) {
        return registry
      }
    }
    return null
  }

  async findSnapshotRegistryBySnapshotInternalName(
    snapshotInternalName: string,
    organizationId?: string,
  ): Promise<DockerRegistry | null> {
    return this.findOneBySnapshotImageNameOrInternalName(snapshotInternalName, RegistryType.SNAPSHOT, organizationId)
  }

  async findRegistryBySnapshotImageName(
    snapshotImageName: string,
    registryType: RegistryType | RegistryType[],
    organizationId?: string,
  ): Promise<DockerRegistry | null> {
    return this.findOneBySnapshotImageNameOrInternalName(snapshotImageName, registryType, organizationId)
  }

  async getRegistryPushAccess(organizationId: string, userId: string): Promise<RegistryPushAccessDto> {
    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    if (organization.blockSharedInfrastructure) {
      throw new ForbiddenException('Using a shared transient registry is not allowed for this organization')
    }

    const transientRegistry = await this.getAvailableTransientRegistry()
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
    if (registry.url.startsWith('localhost:') || registry.url.startsWith('registry:')) {
      return `http://${registry.url}`
    }

    if (registry.url.startsWith('localhost') || registry.url.startsWith('127.0.0.1')) {
      return `http://${registry.url}`
    }

    return registry.url.startsWith('http') ? registry.url : `https://${registry.url}`
  }
}
