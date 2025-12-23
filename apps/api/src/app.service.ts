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
import { RunnerState } from './sandbox/enums/runner-state.enum'

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
    await this.initializeAdminUser()
    await this.initializeTransientRegistry()
    await this.initializeBackupRegistry()
    await this.initializeInternalRegistry()
    await this.initializeBackupRegistry()

    // Default runner init is not awaited because v2 runners depend on the API to be ready
    this.initializeDefaultRunner()
      .then(() => this.initializeDefaultSnapshot())
      .catch((error) => {
        this.logger.error('Error initializing default runner', error)
      })
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
    if (!this.configService.get('defaultRunner.name')) {
      return
    }

    const defaultRegionId = this.configService.getOrThrow('defaultRegion.id')

    const existingRunners = await this.runnerService.findAllByRegion(defaultRegionId)
    if (
      existingRunners.length > 0 &&
      existingRunners.some((r) => r.name === this.configService.get('defaultRunner.name'))
    ) {
      return
    }

    this.logger.log(`Creating default runner: ${this.configService.getOrThrow('defaultRunner.name')}`)

    const runnerVersion = this.configService.get('defaultRunner.apiVersion') || '0'

    if (runnerVersion === '0') {
      const { runner } = await this.runnerService.create({
        apiUrl: this.configService.getOrThrow('defaultRunner.apiUrl'),
        proxyUrl: this.configService.getOrThrow('defaultRunner.proxyUrl'),
        apiKey: this.configService.getOrThrow('defaultRunner.apiKey'),
        cpu: this.configService.getOrThrow('defaultRunner.cpu'),
        memoryGiB: this.configService.getOrThrow('defaultRunner.memory'),
        diskGiB: this.configService.getOrThrow('defaultRunner.disk'),
        regionId: this.configService.getOrThrow('defaultRegion.id'),
        domain: this.configService.getOrThrow('defaultRunner.domain'),
        apiVersion: runnerVersion,
        name: this.configService.getOrThrow('defaultRunner.name'),
      })

      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      this.logger.log(`Waiting for runner ${runner.name} to be healthy...`)
      for (let i = 0; i < 30; i++) {
        try {
          await runnerAdapter.healthCheck()
          this.logger.log(`Runner ${runner.name} is healthy`)
          return
        } catch {
          // ignore
        }
        await new Promise((resolve) => setTimeout(resolve, 1000))
      }
    } else if (runnerVersion === '2') {
      const { runner } = await this.runnerService.create({
        apiKey: this.configService.getOrThrow('defaultRunner.apiKey'),
        regionId: this.configService.getOrThrow('defaultRegion.id'),
        apiVersion: runnerVersion,
        name: this.configService.getOrThrow('defaultRunner.name'),
      })

      this.logger.log(`Waiting for runner ${runner.name} to be healthy...`)
      for (let i = 0; i < 30; i++) {
        const { state } = await this.runnerService.findOneFullOrFail(runner.id)
        if (state === RunnerState.READY) {
          this.logger.log(`Runner ${runner.name} is healthy`)
          return
        }
        await new Promise((resolve) => setTimeout(resolve, 1000))
      }
    }

    this.logger.log(
      `Default runner ${this.configService.getOrThrow('defaultRunner.name')} created successfully but didn't pass health check`,
    )
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
    const existingRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
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
