/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnApplicationBootstrap, OnApplicationShutdown } from '@nestjs/common'
import { DockerRegistryService } from './docker-registry/services/docker-registry.service'
import { RegistryType } from './docker-registry/enums/registry-type.enum'
import { OrganizationService } from './organization/services/organization.service'
import { UserService } from './user/user.service'
import { ApiKeyService } from './api-key/api-key.service'
import { EventEmitterReadinessWatcher } from '@nestjs/event-emitter'
import { SnapshotService } from './sandbox/services/snapshot.service'
import { SystemRole } from './user/enums/system-role.enum'
import { TypedConfigService } from './config/typed-config.service'
import { SchedulerRegistry } from '@nestjs/schedule'
import { RegionService } from './region/services/region.service'
import { RunnerService } from './sandbox/services/runner.service'
import { RunnerAdapterFactory } from './sandbox/runner-adapter/runnerAdapter'
import { RegionType } from './region/enums/region-type.enum'

export const DAYTONA_ADMIN_USER_ID = 'daytona-admin'

@Injectable()
export class AppService implements OnApplicationBootstrap, OnApplicationShutdown {
  private readonly logger = new Logger(AppService.name)

  constructor(
    private readonly dockerRegistryService: DockerRegistryService,
    private readonly configService: TypedConfigService,
    private readonly userService: UserService,
    private readonly organizationService: OrganizationService,
    private readonly apiKeyService: ApiKeyService,
    private readonly eventEmitterReadinessWatcher: EventEmitterReadinessWatcher,
    private readonly snapshotService: SnapshotService,
    private readonly schedulerRegistry: SchedulerRegistry,
    private readonly regionService: RegionService,
    private readonly runnerService: RunnerService,
    private readonly runnerAdapterFactory: RunnerAdapterFactory,
  ) {}

  async onApplicationShutdown(signal?: string) {
    this.logger.log(`Received shutdown signal: ${signal}. Shutting down gracefully...`)
    await this.stopAllCronJobs()
  }

  async onApplicationBootstrap() {
    if (this.configService.get('disableCronJobs') || this.configService.get('maintananceMode')) {
      await this.stopAllCronJobs()
    }

    await this.eventEmitterReadinessWatcher.waitUntilReady()

    await this.initializeDefaultRegion()
    await this.initializeDefaultRunner()
    await this.initializeAdminUser()
    await this.initializeTransientRegistry()
    await this.initializeBackupRegistry()
    await this.initializeInternalRegistry()
    await this.initializeBackupRegistry()
    await this.initializeDefaultSnapshot()
  }

  private async stopAllCronJobs(): Promise<void> {
    for (const cronName of this.schedulerRegistry.getCronJobs().keys()) {
      this.logger.debug(`Stopping cron job: ${cronName}`)
      this.schedulerRegistry.deleteCronJob(cronName)
    }
  }

  private async initializeDefaultRegion(): Promise<void> {
    const existingRegion = await this.regionService.findOne(this.configService.getOrThrow('defaultRegion.id'))
    if (existingRegion) {
      return
    }

    this.logger.log('Initializing default region...')

    await this.regionService.create(
      {
        id: this.configService.getOrThrow('defaultRegion.id'),
        name: this.configService.getOrThrow('defaultRegion.name'),
        enforceQuotas: this.configService.getOrThrow('defaultRegion.enforceQuotas'),
        regionType: RegionType.SHARED,
      },
      null,
    )

    this.logger.log(`Default region created successfully: ${this.configService.getOrThrow('defaultRegion.name')}`)
  }

  private async initializeDefaultRunner(): Promise<void> {
    if (!this.configService.get('defaultRunner.domain')) {
      return
    }

    const existingRunner = await this.runnerService.findOneByDomain(
      this.configService.getOrThrow('defaultRunner.domain'),
    )
    if (existingRunner) {
      return
    }

    this.logger.log(`Creating default runner: ${this.configService.getOrThrow('defaultRunner.domain')}`)

    const { runner } = await this.runnerService.create({
      apiUrl: this.configService.getOrThrow('defaultRunner.apiUrl'),
      proxyUrl: this.configService.getOrThrow('defaultRunner.proxyUrl'),
      apiKey: this.configService.getOrThrow('defaultRunner.apiKey'),
      cpu: this.configService.getOrThrow('defaultRunner.cpu'),
      memoryGiB: this.configService.getOrThrow('defaultRunner.memory'),
      diskGiB: this.configService.getOrThrow('defaultRunner.disk'),
      gpu: this.configService.getOrThrow('defaultRunner.gpu'),
      gpuType: this.configService.getOrThrow('defaultRunner.gpuType'),
      regionId: this.configService.getOrThrow('defaultRegion.id'),
      class: this.configService.getOrThrow('defaultRunner.class'),
      domain: this.configService.getOrThrow('defaultRunner.domain'),
      version: this.configService.get('defaultRunner.version') || '0',
      name: this.configService.getOrThrow('defaultRunner.name'),
    })

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    this.logger.log(`Waiting for runner ${runner.domain} to be healthy...`)
    for (let i = 0; i < 30; i++) {
      try {
        await runnerAdapter.healthCheck()
        this.logger.log(`Runner ${runner.domain} is healthy`)
        break
      } catch {
        // ignore
      }
      await new Promise((resolve) => setTimeout(resolve, 1000))
    }

    this.logger.log(`Default runner created successfully: ${this.configService.getOrThrow('defaultRunner.domain')}`)
  }

  private async initializeAdminUser(): Promise<void> {
    if (await this.userService.findOne(DAYTONA_ADMIN_USER_ID)) {
      return
    }

    const user = await this.userService.create({
      id: DAYTONA_ADMIN_USER_ID,
      name: 'Daytona Admin',
      personalOrganizationQuota: {
        totalCpuQuota: this.configService.getOrThrow('admin.totalCpuQuota'),
        totalMemoryQuota: this.configService.getOrThrow('admin.totalMemoryQuota'),
        totalDiskQuota: this.configService.getOrThrow('admin.totalDiskQuota'),
        maxCpuPerSandbox: this.configService.getOrThrow('admin.maxCpuPerSandbox'),
        maxMemoryPerSandbox: this.configService.getOrThrow('admin.maxMemoryPerSandbox'),
        maxDiskPerSandbox: this.configService.getOrThrow('admin.maxDiskPerSandbox'),
        snapshotQuota: this.configService.getOrThrow('admin.snapshotQuota'),
        maxSnapshotSize: this.configService.getOrThrow('admin.maxSnapshotSize'),
        volumeQuota: this.configService.getOrThrow('admin.volumeQuota'),
      },
      personalOrganizationDefaultRegionId: this.configService.getOrThrow('defaultRegion.id'),
      role: SystemRole.ADMIN,
    })
    const personalOrg = await this.organizationService.findPersonal(user.id)
    const { value } = await this.apiKeyService.createApiKey(
      personalOrg.id,
      user.id,
      DAYTONA_ADMIN_USER_ID,
      [],
      undefined,
      this.configService.getOrThrow('admin.apiKey'),
    )
    this.logger.log(
      `
=========================================
=========================================
Admin user created with API key: ${value}
=========================================
=========================================`,
    )
  }

  private async initializeTransientRegistry(): Promise<void> {
    const existingRegistry = await this.dockerRegistryService.getDefaultTransientRegistry()
    if (existingRegistry) {
      return
    }

    const registryUrl = this.configService.getOrThrow('transientRegistry.url')
    const registryAdmin = this.configService.getOrThrow('transientRegistry.admin')
    const registryPassword = this.configService.getOrThrow('transientRegistry.password')
    const registryProjectId = this.configService.getOrThrow('transientRegistry.projectId')

    if (!registryUrl || !registryAdmin || !registryPassword || !registryProjectId) {
      this.logger.warn('Registry configuration not found, skipping transient registry setup')
      return
    }

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
    const existingRegistry = await this.dockerRegistryService.getAvailableInternalRegistry()
    if (existingRegistry) {
      return
    }

    const registryUrl = this.configService.getOrThrow('internalRegistry.url')
    const registryAdmin = this.configService.getOrThrow('internalRegistry.admin')
    const registryPassword = this.configService.getOrThrow('internalRegistry.password')
    const registryProjectId = this.configService.getOrThrow('internalRegistry.projectId')

    if (!registryUrl || !registryAdmin || !registryPassword || !registryProjectId) {
      this.logger.warn('Registry configuration not found, skipping internal registry setup')
      return
    }

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

  private async initializeBackupRegistry(): Promise<void> {
    const existingRegistry = await this.dockerRegistryService.getAvailableBackupRegistry(
      this.configService.getOrThrow('defaultRegion.id'),
    )
    if (existingRegistry) {
      return
    }

    const registryUrl = this.configService.getOrThrow('internalRegistry.url')
    const registryAdmin = this.configService.getOrThrow('internalRegistry.admin')
    const registryPassword = this.configService.getOrThrow('internalRegistry.password')
    const registryProjectId = this.configService.getOrThrow('internalRegistry.projectId')

    if (!registryUrl || !registryAdmin || !registryPassword || !registryProjectId) {
      this.logger.warn('Registry configuration not found, skipping backup registry setup')
      return
    }

    this.logger.log('Initializing default backup registry...')

    await this.dockerRegistryService.create(
      {
        name: 'Backup Registry',
        url: registryUrl,
        username: registryAdmin,
        password: registryPassword,
        project: registryProjectId,
        registryType: RegistryType.BACKUP,
        isDefault: true,
      },
      undefined,
      true,
    )

    this.logger.log('Default backup registry initialized successfully')
  }

  private async initializeDefaultSnapshot(): Promise<void> {
    const adminPersonalOrg = await this.organizationService.findPersonal(DAYTONA_ADMIN_USER_ID)

    try {
      const existingSnapshot = await this.snapshotService.getSnapshotByName(
        this.configService.getOrThrow('defaultSnapshot'),
        adminPersonalOrg.id,
      )
      if (existingSnapshot) {
        return
      }
    } catch {
      this.logger.log('Default snapshot not found, creating...')
    }

    const defaultSnapshot = this.configService.getOrThrow('defaultSnapshot')

    await this.snapshotService.createFromPull(
      adminPersonalOrg,
      {
        name: defaultSnapshot,
        imageName: defaultSnapshot,
      },
      true,
    )

    this.logger.log('Default snapshot created successfully')
  }
}
