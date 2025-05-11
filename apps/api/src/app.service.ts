/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnApplicationBootstrap } from '@nestjs/common'
import { DockerRegistryService } from './docker-registry/services/docker-registry.service'
import { RegistryType } from './docker-registry/enums/registry-type.enum'
import { OrganizationService } from './organization/services/organization.service'
import { UserService } from './user/user.service'
import { ApiKeyService } from './api-key/api-key.service'
import { EventEmitterReadinessWatcher } from '@nestjs/event-emitter'
import { ImageService } from './workspace/services/image.service'
import { SystemRole } from './user/enums/system-role.enum'
import { TypedConfigService } from './config/typed-config.service'

const DAYTONA_ADMIN_USER_ID = 'daytona-admin'

@Injectable()
export class AppService implements OnApplicationBootstrap {
  private readonly logger = new Logger(AppService.name)

  constructor(
    private readonly dockerRegistryService: DockerRegistryService,
    private readonly configService: TypedConfigService,
    private readonly userService: UserService,
    private readonly organizationService: OrganizationService,
    private readonly apiKeyService: ApiKeyService,
    private readonly eventEmitterReadinessWatcher: EventEmitterReadinessWatcher,
    private readonly imageService: ImageService,
  ) {}

  async onApplicationBootstrap() {
    await this.initializeAdminUser()
    await this.initializeTransientRegistry()
    await this.initializeInternalRegistry()
    await this.initializeDefaultImage()
  }

  private async initializeAdminUser(): Promise<void> {
    if (await this.userService.findOne(DAYTONA_ADMIN_USER_ID)) {
      return
    }

    await this.eventEmitterReadinessWatcher.waitUntilReady()
    const user = await this.userService.create({
      id: DAYTONA_ADMIN_USER_ID,
      name: 'Daytona Admin',
      personalOrganizationQuota: {
        totalCpuQuota: 0,
        totalMemoryQuota: 0,
        totalDiskQuota: 0,
        maxCpuPerWorkspace: 0,
        maxMemoryPerWorkspace: 0,
        maxDiskPerWorkspace: 0,
        maxConcurrentWorkspaces: 0,
        workspaceQuota: 0,
        imageQuota: 100,
        maxImageSize: 100,
        totalImageSize: 1000,
        volumeQuota: 0,
      },
      role: SystemRole.ADMIN,
    })
    const personalOrg = await this.organizationService.findPersonal(user.id)
    await this.apiKeyService.createApiKey(personalOrg.id, user.id, DAYTONA_ADMIN_USER_ID, [])
  }

  private async initializeTransientRegistry(): Promise<void> {
    const existingRegistry = await this.dockerRegistryService.getDefaultTransientRegistry()
    if (existingRegistry) {
      return
    }

    let registryUrl = this.configService.getOrThrow('transientRegistry.url')
    const registryAdmin = this.configService.getOrThrow('transientRegistry.admin')
    const registryPassword = this.configService.getOrThrow('transientRegistry.password')
    const registryProjectId = this.configService.getOrThrow('transientRegistry.projectId')

    if (!registryUrl || !registryAdmin || !registryPassword || !registryProjectId) {
      this.logger.warn('Registry configuration not found, skipping transient registry setup')
      return
    }

    registryUrl = registryUrl.replace(/^(https?:\/\/)/, '')

    this.logger.log('Initializing default transient registry...')

    await this.dockerRegistryService.create({
      name: 'Transient Registry',
      url: registryUrl,
      username: registryAdmin,
      password: registryPassword,
      project: registryProjectId,
      registryType: RegistryType.TRANSIENT,
      isDefault: true,
    })

    this.logger.log('Default transient registry initialized successfully')
  }

  private async initializeInternalRegistry(): Promise<void> {
    const existingRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
    if (existingRegistry) {
      return
    }

    let registryUrl = this.configService.getOrThrow('internalRegistry.url')
    const registryAdmin = this.configService.getOrThrow('internalRegistry.admin')
    const registryPassword = this.configService.getOrThrow('internalRegistry.password')
    const registryProjectId = this.configService.getOrThrow('internalRegistry.projectId')

    if (!registryUrl || !registryAdmin || !registryPassword || !registryProjectId) {
      this.logger.warn('Registry configuration not found, skipping internal registry setup')
      return
    }

    registryUrl = registryUrl.replace(/^(https?:\/\/)/, '')

    this.logger.log('Initializing default internal registry...')

    await this.dockerRegistryService.create({
      name: 'Internal Registry',
      url: registryUrl,
      username: registryAdmin,
      password: registryPassword,
      project: registryProjectId,
      registryType: RegistryType.INTERNAL,
      isDefault: true,
    })

    this.logger.log('Default internal registry initialized successfully')
  }

  private async initializeDefaultImage(): Promise<void> {
    const adminPersonalOrg = await this.organizationService.findPersonal(DAYTONA_ADMIN_USER_ID)

    try {
      const existingImage = await this.imageService.getImageByName(
        this.configService.getOrThrow('defaultImage'),
        adminPersonalOrg.id,
      )
      if (existingImage) {
        return
      }
    } catch {
      this.logger.log('Default image not found, creating...')
    }

    await this.imageService.createImage(
      adminPersonalOrg.id,
      {
        name: this.configService.getOrThrow('defaultImage'),
      },
      null,
      true,
    )
  }
}
