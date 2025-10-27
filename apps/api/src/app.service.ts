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
import { CronJob } from 'cron'

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
  ) {}

  async onApplicationShutdown(signal?: string) {
    this.logger.log(`Received shutdown signal: ${signal}. Shutting down gracefully...`)
    this.stopAllCronJobs()
  }

  async onApplicationBootstrap() {
    this.stopUnusedCronJobs()

    await this.initializeAdminUser()
    await this.initializeTransientRegistry()
    await this.initializeBackupRegistry()
    await this.initializeInternalRegistry()
    await this.initializeDefaultSnapshot()
  }

  private stopUnusedCronJobs(): void {
    if (this.configService.get('cron.disableAll') || this.configService.get('maintananceMode')) {
      this.stopAllCronJobs()
      return
    }

    const disabledScopes = this.configService.get('cron.disabledCronScopes') || []
    const onlyEnabledScopes = this.configService.get('cron.onlyEnabledCronScopes') || []

    if (disabledScopes.length > 0 && onlyEnabledScopes.length > 0) {
      throw new Error('Cannot have both disabled and enabled cron scopes set')
    }

    if (disabledScopes.length === 0 && onlyEnabledScopes.length === 0) {
      return
    }

    const cronJobs = this.schedulerRegistry.getCronJobs()

    const scopedJobs = new Map<string, Map<string, CronJob>>()

    cronJobs.forEach((job, name) => {
      const scope = name.split(':')[0]
      if (!scopedJobs.has(scope)) {
        scopedJobs.set(scope, new Map<string, CronJob>())
      }
      scopedJobs.get(scope)?.set(name, job)
    })

    let scopesToDisable = disabledScopes
    if (onlyEnabledScopes.length > 0) {
      scopesToDisable = Array.from(scopedJobs.keys()).filter((scope) => !onlyEnabledScopes.includes(scope))
    }

    scopesToDisable.forEach((scope) => {
      const jobs = scopedJobs.get(scope)
      if (jobs) {
        Array.from(jobs.keys()).forEach((job) => {
          this.logger.error(`Stopping cron job: ${job} due to not being in enabled scopes`)
          this.schedulerRegistry.deleteCronJob(job)
        })
      }
    })
  }

  private stopAllCronJobs() {
    for (const cronName of this.schedulerRegistry.getCronJobs().keys()) {
      this.logger.debug(`Stopping cron job: ${cronName}`)
      this.schedulerRegistry.deleteCronJob(cronName)
    }
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
        maxCpuPerSandbox: 0,
        maxMemoryPerSandbox: 0,
        maxDiskPerSandbox: 0,
        snapshotQuota: 100,
        maxSnapshotSize: 100,
        volumeQuota: 0,
      },
      role: SystemRole.ADMIN,
    })
    const personalOrg = await this.organizationService.findPersonal(user.id)
    const { value } = await this.apiKeyService.createApiKey(personalOrg.id, user.id, DAYTONA_ADMIN_USER_ID, [])
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
    const existingRegistry = await this.dockerRegistryService.getAvailableBackupRegistry('us')
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

    await this.snapshotService.createSnapshot(
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
