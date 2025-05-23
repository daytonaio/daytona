/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, Not, Repository } from 'typeorm'
import { Workspace } from '../entities/workspace.entity'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { RunnerApiFactory } from '../runner-api/runnerApi'
import { RunnerService } from '../services/runner.service'
import { RunnerState } from '../enums/runner-state.enum'
import { ResourceNotFoundError } from '../../exceptions/not-found.exception'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { BackupState } from '../enums/backup-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/workspace.constants'
import { DockerProvider } from '../docker/docker-provider'
import { fromAxiosError } from '../../common/utils/from-axios-error'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceEvents } from '../constants/workspace-events.constants'
import { WorkspaceDestroyedEvent } from '../events/workspace-destroyed.event'
import { WorkspaceBackupCreatedEvent } from '../events/workspace-backup-created.event'
import { WorkspaceArchivedEvent } from '../events/workspace-archived.event'

@Injectable()
export class BackupManager {
  private readonly logger = new Logger(BackupManager.name)

  constructor(
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    private readonly runnerService: RunnerService,
    private readonly runnerApiFactory: RunnerApiFactory,
    private readonly dockerRegistryService: DockerRegistryService,
    @InjectRedis() private readonly redis: Redis,
    private readonly dockerProvider: DockerProvider,
    private readonly redisLockProvider: RedisLockProvider,
  ) {}

  //  on init
  async onApplicationBootstrap() {
    await this.adHocBackupCheck()
  }

  //  todo: make frequency configurable or more efficient
  @Cron(CronExpression.EVERY_5_MINUTES, { name: 'ad-hoc-backup-check' })
  async adHocBackupCheck(): Promise<void> {
    // Get all ready runners
    const allRunners = await this.runnerService.findAll()
    const readyRunners = allRunners.filter((runner) => runner.state === RunnerState.READY)

    // Process all runners in parallel
    await Promise.all(
      readyRunners.map(async (runner) => {
        const workspaces = await this.workspaceRepository.find({
          where: {
            runnerId: runner.id,
            organizationId: Not(WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION),
            state: In([WorkspaceState.STARTED, WorkspaceState.ARCHIVING]),
            backupState: In([BackupState.NONE, BackupState.COMPLETED]),
          },
          order: {
            lastBackupAt: 'ASC',
          },
          //  todo: increase this number when backup is stable
          take: 10,
        })

        await Promise.all(
          workspaces
            .filter(
              (workspace) =>
                !workspace.lastBackupAt || workspace.lastBackupAt < new Date(Date.now() - 1 * 60 * 60 * 1000),
            )
            .map(async (workspace) => {
              const lockKey = `workspace-backup-${workspace.id}`
              const hasLock = await this.redisLockProvider.lock(lockKey, 60)
              if (!hasLock) {
                return
              }

              try {
                //  todo: remove the catch handler asap
                await this.startBackupCreate(workspace.id).catch((error) => {
                  if (error instanceof BadRequestError && error.message === 'A backup is already in progress') {
                    return
                  }
                  this.logger.error(`Failed to create backup for workspace ${workspace.id}:`, fromAxiosError(error))
                })
              } catch (error) {
                this.logger.error(`Error processing stop state for workspace ${workspace.id}:`, fromAxiosError(error))
              }
            }),
        )
      }),
    )
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-backup-states' }) // Run every 10 seconds
  async syncBackupStates(): Promise<void> {
    //  lock the sync to only run one instance at a time
    const lockKey = 'sync-backup-states'
    const hasLock = await this.redisLockProvider.lock(lockKey, 10)
    if (!hasLock) {
      return
    }

    const workspaces = await this.workspaceRepository.find({
      where: {
        state: In([WorkspaceState.STARTED, WorkspaceState.STOPPED, WorkspaceState.ARCHIVING]),
        backupState: In([BackupState.PENDING, BackupState.IN_PROGRESS]),
      },
    })

    await Promise.all(
      workspaces.map(async (w) => {
        const lockKey = `workspace-backup-${w.id}`
        const hasLock = await this.redisLockProvider.lock(lockKey, 60)
        if (!hasLock) {
          return
        }

        const runner = await this.runnerService.findOne(w.runnerId)
        if (runner.state !== RunnerState.READY) {
          return
        }

        //  get the latest workspace state
        const workspace = await this.workspaceRepository.findOneByOrFail({
          id: w.id,
        })

        try {
          switch (workspace.backupState) {
            case BackupState.PENDING: {
              await this.handlePendingBackup(workspace)
              break
            }
            case BackupState.IN_PROGRESS: {
              await this.checkBackupProgress(workspace)
              break
            }
          }
        } catch (error) {
          //  if error, retry 10 times
          const errorRetryKey = `${lockKey}-error-retry`
          const errorRetryCount = await this.redis.get(errorRetryKey)
          if (!errorRetryCount) {
            await this.redis.setex(errorRetryKey, 300, '1')
          } else if (parseInt(errorRetryCount) > 10) {
            this.logger.error(`Error processing backup for workspace ${workspace.id}:`, fromAxiosError(error))
            await this.updateWorkspacBackupState(workspace.id, BackupState.ERROR)
          } else {
            await this.redis.setex(errorRetryKey, 300, errorRetryCount + 1)
          }
        }
      }),
    ).catch((ex) => {
      this.logger.error(ex)
    })
  }

  async startBackupCreate(workspaceId: string): Promise<void> {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })

    if (!workspace) {
      throw new ResourceNotFoundError('Workspace not found')
    }

    if (workspace.backupState === BackupState.COMPLETED) {
      return
    }

    // Allow backups for STARTED workspaces or STOPPED workspaces with runnerId
    if (
      !(
        workspace.state === WorkspaceState.STARTED ||
        workspace.state === WorkspaceState.ARCHIVING ||
        (workspace.state === WorkspaceState.STOPPED && workspace.runnerId)
      )
    ) {
      throw new BadRequestError('Workspace must be started or stopped with assigned runner to create a backup')
    }

    if (workspace.backupState === BackupState.IN_PROGRESS || workspace.backupState === BackupState.PENDING) {
      return
    }

    // Get default registry
    const registry = await this.dockerRegistryService.getDefaultInternalRegistry()
    if (!registry) {
      throw new BadRequestError('No default registry configured')
    }

    // Generate backup image name
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
    const backupImage = `${registry.url}/${registry.project}/backup-${workspace.id}:${timestamp}`

    //  if workspace has a backup image, add it to the existingBackupImages array
    if (
      workspace.lastBackupAt &&
      workspace.backupImage &&
      [BackupState.NONE, BackupState.COMPLETED].includes(workspace.backupState)
    ) {
      workspace.existingBackupImages.push({
        imageName: workspace.backupImage,
        createdAt: workspace.lastBackupAt,
      })
    }
    const existingBackupImages = workspace.existingBackupImages
    existingBackupImages.push({
      imageName: backupImage,
      createdAt: new Date(),
    })

    const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
      id: workspace.id,
    })
    workspaceToUpdate.existingBackupImages = existingBackupImages
    workspaceToUpdate.backupState = BackupState.PENDING
    workspaceToUpdate.backupRegistryId = registry.id
    workspaceToUpdate.backupImage = backupImage
    await this.workspaceRepository.save(workspaceToUpdate)
  }

  private async checkBackupProgress(workspace: Workspace): Promise<void> {
    try {
      const runner = await this.runnerService.findOne(workspace.runnerId)
      const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)

      // Get workspace info from runner
      const workspaceInfo = await runnerWorkspaceApi.info(workspace.id)

      switch (workspaceInfo.data.backupState?.toUpperCase()) {
        case 'COMPLETED': {
          workspace.backupState = BackupState.COMPLETED
          workspace.lastBackupAt = new Date()
          const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
            id: workspace.id,
          })
          workspaceToUpdate.backupState = BackupState.COMPLETED
          workspaceToUpdate.lastBackupAt = new Date()
          await this.workspaceRepository.save(workspaceToUpdate)
          break
        }
        case 'FAILED':
        case 'ERROR': {
          await this.updateWorkspacBackupState(workspace.id, BackupState.ERROR)
          break
        }

        // If still in progress or any other state, do nothing and wait for next sync
      }
    } catch (error) {
      await this.updateWorkspacBackupState(workspace.id, BackupState.ERROR)
      throw error
    }
  }

  private async deleteSandboxBackupRepositoryFromRegistry(workspace: Workspace): Promise<void> {
    const registry = await this.dockerRegistryService.findOne(workspace.backupRegistryId)

    try {
      await this.dockerProvider.deleteSandboxRepository(workspace.id, registry)
    } catch (error) {
      this.logger.error(
        `Failed to delete backup repository ${workspace.id} from registry ${registry.id}:`,
        fromAxiosError(error),
      )
    }
  }

  private async handlePendingBackup(workspace: Workspace): Promise<void> {
    try {
      const registry = await this.dockerRegistryService.findOne(workspace.backupRegistryId)
      if (!registry) {
        throw new Error('Registry not found')
      }

      const runner = await this.runnerService.findOne(workspace.runnerId)
      const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)

      //  check if backup is already in progress on the runner
      const runnerWorkspaceResponse = await runnerWorkspaceApi.info(workspace.id)
      const runnerWorkspace = runnerWorkspaceResponse.data
      if (runnerWorkspace.backupState?.toUpperCase() === 'IN_PROGRESS') {
        return
      }

      // Initiate backup on runner
      await runnerWorkspaceApi.createBackup(workspace.id, {
        registry: {
          url: registry.url,
          username: registry.username,
          password: registry.password,
        },
        image: workspace.backupImage,
      })

      await this.updateWorkspacBackupState(workspace.id, BackupState.IN_PROGRESS)
    } catch (error) {
      if (error.response?.status === 400 && error.response?.data?.message.includes('A backup is already in progress')) {
        await this.updateWorkspacBackupState(workspace.id, BackupState.IN_PROGRESS)
        return
      }
      await this.updateWorkspacBackupState(workspace.id, BackupState.ERROR)
      throw error
    }
  }

  @Cron(CronExpression.EVERY_30_SECONDS, { name: 'sync-stop-state-create-backups' }) // Run every 30 seconds
  async syncStopStateCreateBackups(): Promise<void> {
    const lockKey = 'sync-stop-state-create-backups'
    const hasLock = await this.redisLockProvider.lock(lockKey, 30)
    if (!hasLock) {
      return
    }

    const workspaces = await this.workspaceRepository.find({
      where: {
        state: In([WorkspaceState.STOPPED, WorkspaceState.ARCHIVING]),
        backupState: In([BackupState.NONE]),
      },
      //  todo: increase this number when auto-stop is stable
      take: 10,
    })

    await Promise.all(
      workspaces
        .filter((workspace) => workspace.runnerId !== null)
        .map(async (workspace) => {
          const lockKey = `workspace-backup-${workspace.id}`
          const hasLock = await this.redisLockProvider.lock(lockKey, 30)
          if (!hasLock) {
            return
          }

          const runner = await this.runnerService.findOne(workspace.runnerId)
          if (runner.state !== RunnerState.READY) {
            return
          }

          //  TODO: this should be revisited
          //  an error should be handled better and not just logged
          try {
            //  todo: remove the catch handler asap
            await this.startBackupCreate(workspace.id).catch((error) => {
              if (error instanceof BadRequestError && error.message === 'A backup is already in progress') {
                return
              }
              this.logger.error(`Failed to create backup for workspace ${workspace.id}:`, fromAxiosError(error))
            })
          } catch (error) {
            this.logger.error(`Failed to create backup for workspace ${workspace.id}:`, fromAxiosError(error))
          }
        }),
    )
  }

  private async updateWorkspacBackupState(workspaceId: string, backupState: BackupState): Promise<void> {
    const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })
    workspaceToUpdate.backupState = backupState
    await this.workspaceRepository.save(workspaceToUpdate)
  }

  @OnEvent(WorkspaceEvents.ARCHIVED)
  private async handleWorkspaceArchivedEvent(event: WorkspaceArchivedEvent) {
    this.startBackupCreate(event.workspace.id)
  }

  @OnEvent(WorkspaceEvents.DESTROYED)
  private async handleWorkspaceDestroyedEvent(event: WorkspaceDestroyedEvent) {
    this.deleteSandboxBackupRepositoryFromRegistry(event.workspace)
  }

  @OnEvent(WorkspaceEvents.BACKUP_CREATED)
  private async handleWorkspaceBackupCreatedEvent(event: WorkspaceBackupCreatedEvent) {
    this.handlePendingBackup(event.workspace)
  }
}
