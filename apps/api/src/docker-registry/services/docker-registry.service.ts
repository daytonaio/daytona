/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ForbiddenException, Inject, Injectable, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { FindOptionsWhere, In, IsNull, Repository } from 'typeorm'
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
import { OrganizationService } from '../../organization/services/organization.service'

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

  // TODO: transactional
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
   * If `organizationId` is not provided, the default *shared* snapshot registry is returned (if exists).
   *
   * If `organizationId` is provided and shared infrastructure is blocked for the organization, the default snapshot registry for the organization is returned (if exists).
   *
   * If shared infrastructure is not blocked for the organization, the default *shared* snapshot registry is returned (if exists) as a fallback if no organization snapshot registry exists.
   */
  async getDefaultSnapshotRegistry(organizationId?: string): Promise<DockerRegistry | null> {
    const baseFindOptions: FindOptionsWhere<DockerRegistry> = {
      isDefault: true,
      registryType: RegistryType.SNAPSHOT,
    }

    if (!organizationId) {
      // Return the default shared registry (if exists)
      return this.dockerRegistryRepository.findOne({
        where: {
          ...baseFindOptions,
          organizationId: IsNull(),
        },
      })
    }

    const organization = await this.organizationService.findOne(organizationId)
    if (!organization) {
      throw new NotFoundException('Organization not found')
    }

    const orgRegistry = await this.dockerRegistryRepository.findOne({
      where: {
        ...baseFindOptions,
        organizationId,
      },
    })

    if (orgRegistry) {
      // Prefer default organization registry
      return orgRegistry
    }

    if (!organization.blockSharedInfrastructure) {
      // Return the default shared registry (if exists)
      return this.dockerRegistryRepository.findOne({
        where: {
          ...baseFindOptions,
          organizationId: IsNull(),
        },
      })
    }

    return null
  }

  async getDefaultTransientRegistry(): Promise<DockerRegistry | null> {
    return this.dockerRegistryRepository.findOne({
      where: { isDefault: true, registryType: RegistryType.TRANSIENT },
    })
  }

  async getAvailableBackupRegistry(preferredRegionId: string): Promise<DockerRegistry | null> {
    const registries = await this.dockerRegistryRepository.find({
      where: { registryType: RegistryType.BACKUP, isDefault: true },
    })

    if (registries.length === 0) {
      return null
    }

    // Filter registries by preferred region
    const preferredRegionRegistries = registries.filter((registry) => registry.regionId === preferredRegionId)

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
    if (registry.url.startsWith('localhost:') || registry.url.startsWith('registry:')) {
      return `http://${registry.url}`
    }

    if (registry.url.startsWith('localhost') || registry.url.startsWith('127.0.0.1')) {
      return `http://${registry.url}`
    }

    return registry.url.startsWith('http') ? registry.url : `https://${registry.url}`
  }
}
