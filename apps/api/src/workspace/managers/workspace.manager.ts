/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, Not, Raw, Repository } from 'typeorm'
import { Workspace } from '../entities/workspace.entity'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { WorkspaceDesiredState } from '../enums/workspace-desired-state.enum'
import { RunnerApiFactory } from '../runner-api/runnerApi'
import { RunnerService } from '../services/runner.service'
import { EnumsSandboxState as RunnerWorkspaceState } from '@daytonaio/runner-api-client'
import { RunnerState } from '../enums/runner-state.enum'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { BackupState } from '../enums/backup-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { ImageService } from '../services/image.service'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/workspace.constants'
import { DockerProvider } from '../docker/docker-provider'
import { ImageRunnerState } from '../enums/image-runner-state.enum'
import { BuildInfo } from '../entities/build-info.entity'
import { CreateSandboxDTO } from '@daytonaio/runner-api-client'
import { fromAxiosError } from '../../common/utils/from-axios-error'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceEvents } from '../constants/workspace-events.constants'
import { WorkspaceStoppedEvent } from '../events/workspace-stopped.event'
import { WorkspaceStartedEvent } from '../events/workspace-started.event'
import { WorkspaceArchivedEvent } from '../events/workspace-archived.event'
import { WorkspaceDestroyedEvent } from '../events/workspace-destroyed.event'
import { WorkspaceCreatedEvent } from '../events/workspace-create.event'
import { ImageRunner } from '../entities/image-runner.entity'

const SYNC_INSTANCE_STATE_LOCK_KEY = 'sync-instance-state-'
const SYNC_AGAIN = true
const DONT_SYNC_AGAIN = false
type ShouldSyncAgain = boolean
type StateSyncHandler = (workspace: Workspace) => Promise<ShouldSyncAgain>

@Injectable()
export class WorkspaceManager {
  private readonly logger = new Logger(WorkspaceManager.name)

  constructor(
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    @InjectRepository(ImageRunner)
    private readonly imageRunnerRepository: Repository<ImageRunner>,
    private readonly runnerService: RunnerService,
    private readonly runnerApiFactory: RunnerApiFactory,
    private readonly dockerRegistryService: DockerRegistryService,
    @InjectRedis() private readonly redis: Redis,
    private readonly imageService: ImageService,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly dockerProvider: DockerProvider,
  ) { }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'auto-stop-check' })
  async autostopCheck(): Promise<void> {
    //  lock the sync to only run one instance at a time
    //  keep the worker selected for 1 minute

    if (!(await this.redisLockProvider.lock('auto-stop-check-worker-selected', 60))) {
      return
    }

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
            state: WorkspaceState.STARTED,
            desiredState: WorkspaceDesiredState.STARTED,
            pending: false,
            autoStopInterval: Not(0),
            lastActivityAt: Raw((alias) => `${alias} < NOW() - INTERVAL '1 minute' * "autoStopInterval"`),
          },
          order: {
            lastBackupAt: 'ASC',
          },
          //  todo: increase this number when auto-stop is stable
          take: 10,
        })

        await Promise.all(
          workspaces.map(async (workspace) => {
            const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + workspace.id
            const acquired = await this.redisLockProvider.lock(lockKey, 30)
            if (!acquired) {
              return
            }

            try {
              workspace.desiredState = WorkspaceDesiredState.STOPPED
              await this.workspaceRepository.save(workspace)
              await this.redisLockProvider.unlock(lockKey)
              this.syncInstanceState(workspace.id)
            } catch (error) {
              this.logger.error(
                `Error processing auto-stop state for workspace ${workspace.id}:`,
                fromAxiosError(error),
              )
            }
          }),
        )
      }),
    )
  }

  @Cron(CronExpression.EVERY_MINUTE, { name: 'auto-archive-check' })
  async autoArchiveCheck(): Promise<void> {
    //  lock the sync to only run one instance at a time
    const autoArchiveCheckWorkerSelected = await this.redis.get('auto-archive-check-worker-selected')
    if (autoArchiveCheckWorkerSelected) {
      return
    }
    //  keep the worker selected for 1 minute
    await this.redis.setex('auto-archive-check-worker-selected', 60, '1')

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
            state: WorkspaceState.STOPPED,
            desiredState: WorkspaceDesiredState.STOPPED,
            pending: false,
            lastActivityAt: Raw((alias) => `${alias} < NOW() - INTERVAL '1 minute' * "autoArchiveInterval"`),
          },
          order: {
            lastBackupAt: 'ASC',
          },
          //  max 3 workspaces can be archived at the same time on the same runner
          //  this is to prevent the runner from being overloaded
          take: 3,
        })

        await Promise.all(
          workspaces.map(async (workspace) => {
            const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + workspace.id
            const acquired = await this.redisLockProvider.lock(lockKey, 30)
            if (!acquired) {
              return
            }

            try {
              workspace.desiredState = WorkspaceDesiredState.ARCHIVED
              await this.workspaceRepository.save(workspace)
              await this.redisLockProvider.unlock(lockKey)
              this.syncInstanceState(workspace.id)
            } catch (error) {
              this.logger.error(
                `Error processing auto-archive state for workspace ${workspace.id}:`,
                fromAxiosError(error),
              )
            }
          }),
        )
      }),
    )
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-states' })
  async syncStates(): Promise<void> {
    const lockKey = 'sync-states'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    const workspaces = await this.workspaceRepository.find({
      where: {
        state: Not(In([WorkspaceState.DESTROYED, WorkspaceState.ERROR])),
        desiredState: Raw(
          () =>
            `"Workspace"."desiredState"::text != "Workspace"."state"::text AND "Workspace"."desiredState"::text != 'archived'`,
        ),
      },
      take: 100,
      order: {
        lastActivityAt: 'DESC',
      },
    })

    await Promise.all(
      workspaces.map(async (workspace) => {
        this.syncInstanceState(workspace.id)
      }),
    )
    await this.redisLockProvider.unlock(lockKey)
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-archived-desired-states' })
  async syncArchivedDesiredStates(): Promise<void> {
    const lockKey = 'sync-archived-desired-states'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    const runnersWith3InProgress = await this.workspaceRepository
      .createQueryBuilder('workspace')
      .select('"runnerId"')
      .where('"workspace"."state" = :state', { state: WorkspaceState.ARCHIVING })
      .groupBy('"runnerId"')
      .having('COUNT(*) >= 3')
      .getRawMany()

    const workspaces = await this.workspaceRepository.find({
      where: [
        {
          state: WorkspaceState.ARCHIVING,
          desiredState: WorkspaceDesiredState.ARCHIVED,
        },
        {
          state: Not(In([WorkspaceState.ARCHIVED, WorkspaceState.DESTROYED, WorkspaceState.ERROR])),
          desiredState: WorkspaceDesiredState.ARCHIVED,
          runnerId: Not(In(runnersWith3InProgress.map((runner) => runner.runnerId))),
        },
      ],
      take: 100,
      order: {
        lastActivityAt: 'DESC',
      },
    })

    await Promise.all(
      workspaces.map(async (workspace) => {
        this.syncInstanceState(workspace.id)
      }),
    )
    await this.redisLockProvider.unlock(lockKey)
  }

  async syncInstanceState(workspaceId: string): Promise<void> {
    //  prevent syncState cron from running multiple instances of the same workspace
    const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + workspaceId
    const acquired = await this.redisLockProvider.lock(lockKey, 360)
    if (!acquired) {
      return
    }

    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })

    if (workspace.state === WorkspaceState.ERROR) {
      await this.redisLockProvider.unlock(lockKey)
      return
    }

    let shouldSyncAgain = DONT_SYNC_AGAIN

    try {
      switch (workspace.desiredState) {
        case WorkspaceDesiredState.STARTED: {
          shouldSyncAgain = await this.handleWorkspaceDesiredStateStarted(workspace)
          break
        }
        case WorkspaceDesiredState.STOPPED: {
          shouldSyncAgain = await this.handleWorkspaceDesiredStateStopped(workspace)
          break
        }
        case WorkspaceDesiredState.DESTROYED: {
          shouldSyncAgain = await this.handleWorkspaceDesiredStateDestroyed(workspace)
          break
        }
        case WorkspaceDesiredState.ARCHIVED: {
          shouldSyncAgain = await this.handleWorkspaceDesiredStateArchived(workspace)
          break
        }
      }
    } catch (error) {
      if (error.code === 'ECONNRESET') {
        shouldSyncAgain = SYNC_AGAIN
      } else {
        this.logger.error(`Error processing desired state for workspace ${workspaceId}:`, fromAxiosError(error))

        const workspace = await this.workspaceRepository.findOneBy({
          id: workspaceId,
        })
        if (!workspace) {
          //  edge case where workspace is deleted while desired state is being processed
          return
        }
        await this.updateWorkspaceErrorState(workspace.id, error.message || String(error))
      }
    }

    await this.redisLockProvider.unlock(lockKey)
    if (shouldSyncAgain) {
      this.syncInstanceState(workspaceId)
    }
  }

  private handleUnassignedBuildWorkspace: StateSyncHandler = async (workspace: Workspace): Promise<ShouldSyncAgain> => {
    // Try to assign an available runner with the image build
    let runnerId: string
    try {
      runnerId = await this.runnerService.getRandomAvailableRunner({
        region: workspace.region,
        workspaceClass: workspace.class,
        imageRef: workspace.buildInfo.imageRef,
      })
    } catch (error) {
      // Continue to next assignment method
    }

    if (runnerId) {
      await this.updateWorkspaceState(workspace.id, WorkspaceState.UNKNOWN, runnerId)
      return SYNC_AGAIN
    }

    // Try to assign an available runner that is currently building the image
    const imageRunners = await this.runnerService.getImageRunners(workspace.buildInfo.imageRef)

    for (const imageRunner of imageRunners) {
      const runner = await this.runnerService.findOne(imageRunner.runnerId)
      if (runner.used < runner.capacity) {
        if (imageRunner.state === ImageRunnerState.BUILDING_IMAGE) {
          await this.updateWorkspaceState(workspace.id, WorkspaceState.BUILDING_IMAGE, runner.id)
          return SYNC_AGAIN
        } else if (imageRunner.state === ImageRunnerState.ERROR) {
          await this.updateWorkspaceErrorState(workspace.id, imageRunner.errorReason)
          return DONT_SYNC_AGAIN
        }
      }
    }

    const excludedRunnerIds = await this.runnerService.getRunnersWithMultipleImagesBuilding()

    // Try to assign a new available runner
    runnerId = await this.runnerService.getRandomAvailableRunner({
      region: workspace.region,
      workspaceClass: workspace.class,
      excludedRunnerIds: excludedRunnerIds,
    })

    this.buildOnRunner(workspace.buildInfo, runnerId, workspace.organizationId)

    await this.updateWorkspaceState(workspace.id, WorkspaceState.BUILDING_IMAGE, runnerId)
    await this.runnerService.recalculateRunnerUsage(runnerId)
    return SYNC_AGAIN
  }

  // Initiates the image build on the runner and creates an ImageRunner depending on the result
  async buildOnRunner(buildInfo: BuildInfo, runnerId: string, organizationId: string) {
    const runner = await this.runnerService.findOne(runnerId)
    const runnerImageApi = this.runnerApiFactory.createImageApi(runner)

    let retries = 0

    while (retries < 10) {
      try {
        await runnerImageApi.buildImage({
          image: buildInfo.imageRef,
          organizationId: organizationId,
          dockerfile: buildInfo.dockerfileContent,
          context: buildInfo.contextHashes,
        })
        break
      } catch (err) {
        if (err.code !== 'ECONNRESET') {
          await this.runnerService.createImageRunner(runnerId, buildInfo.imageRef, ImageRunnerState.ERROR, err.message)
          return
        }
      }
      retries++
      await new Promise((resolve) => setTimeout(resolve, retries * 1000))
    }

    if (retries === 10) {
      await this.runnerService.createImageRunner(
        runnerId,
        buildInfo.imageRef,
        ImageRunnerState.ERROR,
        'Timeout while building',
      )
      return
    }

    const response = (await runnerImageApi.imageExists(buildInfo.imageRef)).data
    let state = ImageRunnerState.BUILDING_IMAGE
    if (response && response.exists) {
      state = ImageRunnerState.READY
    }

    await this.runnerService.createImageRunner(runnerId, buildInfo.imageRef, state)
  }

  private handleWorkspaceDesiredStateArchived: StateSyncHandler = async (
    workspace: Workspace,
  ): Promise<ShouldSyncAgain> => {
    const lockKey = 'archive-lock-' + workspace.runnerId
    if (!(await this.redisLockProvider.lock(lockKey, 10))) {
      return DONT_SYNC_AGAIN
    }

    const inProgressOnRunner = await this.workspaceRepository.find({
      where: {
        runnerId: workspace.runnerId,
        state: In([WorkspaceState.ARCHIVING]),
      },
      order: {
        lastActivityAt: 'DESC',
      },
      take: 100,
    })

    //  if the workspace is already in progress, continue
    if (!inProgressOnRunner.find((w) => w.id === workspace.id)) {
      //  max 3 workspaces can be archived at the same time on the same runner
      //  this is to prevent the runner from being overloaded
      if (inProgressOnRunner.length > 2) {
        await this.redisLockProvider.unlock(lockKey)
        return
      }
    }

    switch (workspace.state) {
      case WorkspaceState.STOPPED: {
        await this.updateWorkspaceState(workspace.id, WorkspaceState.ARCHIVING)
        //  fallthrough to archiving state
      }
      case WorkspaceState.ARCHIVING: {
        await this.redisLockProvider.unlock(lockKey)

        //  if the backup state is error, we need to retry the backup
        if (workspace.backupState === BackupState.ERROR) {
          const archiveErrorRetryKey = 'archive-error-retry-' + workspace.id
          const archiveErrorRetryCountRaw = await this.redis.get(archiveErrorRetryKey)
          const archiveErrorRetryCount = archiveErrorRetryCountRaw ? parseInt(archiveErrorRetryCountRaw) : 0
          //  if the archive error retry count is greater than 3, we need to mark the workspace as error
          if (archiveErrorRetryCount > 3) {
            await this.updateWorkspaceErrorState(workspace.id, 'Failed to archive workspace')
            await this.redis.del(archiveErrorRetryKey)
            return DONT_SYNC_AGAIN
          }
          await this.redis.setex('archive-error-retry-' + workspace.id, 720, String(archiveErrorRetryCount + 1))

          //  reset the backup state to pending to retry the backup
          await this.workspaceRepository.update(workspace.id, {
            backupState: BackupState.PENDING,
          })

          return DONT_SYNC_AGAIN
        }

        // Check for timeout - if more than 30 minutes since last activity
        const thirtyMinutesAgo = new Date(Date.now() - 30 * 60 * 1000)
        if (workspace.lastActivityAt < thirtyMinutesAgo) {
          await this.updateWorkspaceErrorState(workspace.id, 'Archiving operation timed out')
          return DONT_SYNC_AGAIN
        }

        if (workspace.backupState !== BackupState.COMPLETED) {
          return DONT_SYNC_AGAIN
        }

        //  when the backup is completed, destroy the workspace on the runner
        //  and deassociate the workspace from the runner
        const runner = await this.runnerService.findOne(workspace.runnerId)
        const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)

        try {
          const workspaceInfoResponse = await runnerWorkspaceApi.info(workspace.id)
          const workspaceInfo = workspaceInfoResponse.data
          switch (workspaceInfo.state) {
            case RunnerWorkspaceState.SandboxStateDestroying:
              //  wait until workspace is destroyed on runner
              return SYNC_AGAIN
            case RunnerWorkspaceState.SandboxStateDestroyed:
              await this.updateWorkspaceState(workspace.id, WorkspaceState.ARCHIVED, null)
              return DONT_SYNC_AGAIN
            default:
              await runnerWorkspaceApi.destroy(workspace.id)
              return SYNC_AGAIN
          }
        } catch (error) {
          //  fail for errors other than workspace not found or workspace already destroyed
          if (
            !(
              (error.response?.data?.statusCode === 400 &&
                error.response?.data?.message.includes('Workspace already destroyed')) ||
              error.response?.status === 404
            )
          ) {
            throw error
          }
          //  if the workspace is already destroyed, do nothing
          await this.updateWorkspaceState(workspace.id, WorkspaceState.ARCHIVED, null)
          return DONT_SYNC_AGAIN
        }
      }
    }

    return DONT_SYNC_AGAIN
  }

  private handleWorkspaceDesiredStateDestroyed: StateSyncHandler = async (
    workspace: Workspace,
  ): Promise<ShouldSyncAgain> => {
    if (workspace.state === WorkspaceState.ARCHIVED) {
      await this.updateWorkspaceState(workspace.id, WorkspaceState.DESTROYED)
      return DONT_SYNC_AGAIN
    }

    const runner = await this.runnerService.findOne(workspace.runnerId)
    if (runner.state !== RunnerState.READY) {
      //  console.debug(`Runner ${runner.id} is not ready`);
      return DONT_SYNC_AGAIN
    }

    switch (workspace.state) {
      case WorkspaceState.DESTROYED:
        return DONT_SYNC_AGAIN
      case WorkspaceState.DESTROYING: {
        // check if workspace is destroyed
        const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)

        try {
          const workspaceInfoResponse = await runnerWorkspaceApi.info(workspace.id)
          const workspaceInfo = workspaceInfoResponse.data
          if (
            workspaceInfo.state === RunnerWorkspaceState.SandboxStateDestroyed ||
            workspaceInfo.state === RunnerWorkspaceState.SandboxStateError
          ) {
            await runnerWorkspaceApi.removeDestroyed(workspace.id)
          }
        } catch (e) {
          //  if the workspace is not found on runner, it is already destroyed
          if (!e.response || e.response.status !== 404) {
            throw e
          }
        }

        await this.updateWorkspaceState(workspace.id, WorkspaceState.DESTROYED)
        return SYNC_AGAIN
      }
      default: {
        // destroy workspace
        try {
          const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)
          const workspaceInfoResponse = await runnerWorkspaceApi.info(workspace.id)
          const workspaceInfo = workspaceInfoResponse.data
          if (workspaceInfo?.state === RunnerWorkspaceState.SandboxStateDestroyed) {
            return DONT_SYNC_AGAIN
          }
          await runnerWorkspaceApi.destroy(workspace.id)
        } catch (e) {
          //  if the workspace is not found on runner, it is already destroyed
          if (e.response.status !== 404) {
            throw e
          }
        }
        await this.updateWorkspaceState(workspace.id, WorkspaceState.DESTROYING)
        return SYNC_AGAIN
      }
    }
  }

  private handleWorkspaceDesiredStateStarted: StateSyncHandler = async (
    workspace: Workspace,
  ): Promise<ShouldSyncAgain> => {
    switch (workspace.state) {
      case WorkspaceState.PENDING_BUILD: {
        return this.handleUnassignedBuildWorkspace(workspace)
      }
      case WorkspaceState.BUILDING_IMAGE: {
        return this.handleRunnerWorkspaceBuildingImageStateOnDesiredStateStart(workspace)
      }
      case WorkspaceState.UNKNOWN: {
        return this.handleRunnerWorkspaceUnknownStateOnDesiredStateStart(workspace)
      }
      case WorkspaceState.ARCHIVED:
      case WorkspaceState.STOPPED: {
        return this.handleRunnerWorkspaceStoppedOrArchivedStateOnDesiredStateStart(workspace)
      }
      case WorkspaceState.RESTORING:
      case WorkspaceState.CREATING: {
        return this.handleRunnerWorkspacePullingImageStateCheck(workspace)
      }
      case WorkspaceState.PULLING_IMAGE:
      case WorkspaceState.STARTING: {
        return this.handleRunnerWorkspaceStartedStateCheck(workspace)
      }
      //  TODO: remove this case
      case WorkspaceState.ERROR: {
        //  TODO: remove this asap
        //  this was a temporary solution to recover from the false positive error state
        if (workspace.id.startsWith('err_')) {
          return DONT_SYNC_AGAIN
        }
        const runner = await this.runnerService.findOne(workspace.runnerId)
        const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)
        const workspaceInfoResponse = await runnerWorkspaceApi.info(workspace.id)
        const workspaceInfo = workspaceInfoResponse.data
        if (workspaceInfo.state === RunnerWorkspaceState.SandboxStateStarted) {
          const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
            id: workspace.id,
          })
          workspaceToUpdate.state = WorkspaceState.STARTED
          workspaceToUpdate.backupState = BackupState.NONE
          await this.workspaceRepository.save(workspaceToUpdate)
        }
      }
    }

    return DONT_SYNC_AGAIN
  }

  private handleWorkspaceDesiredStateStopped: StateSyncHandler = async (
    workspace: Workspace,
  ): Promise<ShouldSyncAgain> => {
    const runner = await this.runnerService.findOne(workspace.runnerId)
    if (runner.state !== RunnerState.READY) {
      //  console.debug(`Runner ${runner.id} is not ready`);
      return DONT_SYNC_AGAIN
    }

    switch (workspace.state) {
      case WorkspaceState.STARTED: {
        // stop workspace
        const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)
        await runnerWorkspaceApi.stop(workspace.id)
        await this.updateWorkspaceState(workspace.id, WorkspaceState.STOPPING)
        //  sync states again immediately for workspace
        return SYNC_AGAIN
      }
      case WorkspaceState.STOPPING: {
        // check if workspace is stopped
        const runner = await this.runnerService.findOne(workspace.runnerId)
        const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)
        const workspaceInfoResponse = await runnerWorkspaceApi.info(workspace.id)
        const workspaceInfo = workspaceInfoResponse.data
        switch (workspaceInfo.state) {
          case RunnerWorkspaceState.SandboxStateStopped: {
            const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
              id: workspace.id,
            })
            workspaceToUpdate.state = WorkspaceState.STOPPED
            workspaceToUpdate.backupState = BackupState.NONE
            await this.workspaceRepository.save(workspaceToUpdate)
            return SYNC_AGAIN
          }
          case RunnerWorkspaceState.SandboxStateError: {
            await this.updateWorkspaceErrorState(workspace.id, 'Sandbox is in error state on runner')
            return DONT_SYNC_AGAIN
          }
        }
        return SYNC_AGAIN
      }
      case WorkspaceState.ERROR: {
        if (workspace.id.startsWith('err_')) {
          return DONT_SYNC_AGAIN
        }
        const runner = await this.runnerService.findOne(workspace.runnerId)
        const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)
        const workspaceInfoResponse = await runnerWorkspaceApi.info(workspace.id)
        const workspaceInfo = workspaceInfoResponse.data
        if (workspaceInfo.state === RunnerWorkspaceState.SandboxStateStopped) {
          await this.updateWorkspaceState(workspace.id, WorkspaceState.STOPPED)
        }
      }
    }

    return DONT_SYNC_AGAIN
  }

  private handleRunnerWorkspaceBuildingImageStateOnDesiredStateStart: StateSyncHandler = async (
    workspace: Workspace,
  ): Promise<ShouldSyncAgain> => {
    const imageRunner = await this.runnerService.getImageRunner(workspace.runnerId, workspace.buildInfo.imageRef)
    if (imageRunner) {
      switch (imageRunner.state) {
        case ImageRunnerState.READY: {
          // TODO: "UNKNOWN" should probably be changed to something else
          await this.updateWorkspaceState(workspace.id, WorkspaceState.UNKNOWN)
          return SYNC_AGAIN
        }
        case ImageRunnerState.ERROR: {
          await this.updateWorkspaceErrorState(workspace.id, imageRunner.errorReason)
          return DONT_SYNC_AGAIN
        }
      }
    }
    if (!imageRunner || imageRunner.state === ImageRunnerState.BUILDING_IMAGE) {
      // Sleep for a second and go back to syncing instance state
      await new Promise((resolve) => setTimeout(resolve, 1000))
      return SYNC_AGAIN
    }

    return DONT_SYNC_AGAIN
  }

  private handleRunnerWorkspaceUnknownStateOnDesiredStateStart: StateSyncHandler = async (
    workspace: Workspace,
  ): Promise<ShouldSyncAgain> => {
    const runner = await this.runnerService.findOne(workspace.runnerId)
    if (runner.state !== RunnerState.READY) {
      //  console.debug(`Runner ${runner.id} is not ready`);
      return DONT_SYNC_AGAIN
    }

    let createWorkspaceDto: CreateSandboxDTO = {
      id: workspace.id,
      osUser: workspace.osUser,
      image: '',
      // TODO: organizationId: workspace.organizationId,
      userId: workspace.organizationId,
      storageQuota: workspace.disk,
      memoryQuota: workspace.mem,
      cpuQuota: workspace.cpu,
      // gpuQuota: workspace.gpu,
      env: workspace.env,
      // public: workspace.public,
      volumes: workspace.volumes,
    }

    if (!workspace.buildInfo) {
      //  get internal image name
      const image = await this.imageService.getImageByName(workspace.image, workspace.organizationId)
      const internalImageName = image.internalName

      const registry = await this.dockerRegistryService.findOneByImageName(internalImageName, workspace.organizationId)
      if (!registry) {
        throw new Error('No registry found for image')
      }

      createWorkspaceDto = {
        ...createWorkspaceDto,
        image: internalImageName,
        entrypoint: image.entrypoint,
        registry: {
          url: registry.url,
          username: registry.username,
          password: registry.password,
        },
      }
    } else {
      createWorkspaceDto = {
        ...createWorkspaceDto,
        image: workspace.buildInfo.imageRef,
        entrypoint: this.getEntrypointFromDockerfile(workspace.buildInfo.dockerfileContent),
      }
    }

    const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)
    await runnerWorkspaceApi.create(createWorkspaceDto)
    await this.updateWorkspaceState(workspace.id, WorkspaceState.CREATING)
    //  sync states again immediately for workspace
    return SYNC_AGAIN
  }

  // TODO: revise/cleanup
  private getEntrypointFromDockerfile(dockerfileContent: string): string[] {
    // Match ENTRYPOINT with either a string or JSON array
    const entrypointMatch = dockerfileContent.match(/ENTRYPOINT\s+(.*)/)
    if (entrypointMatch) {
      const rawEntrypoint = entrypointMatch[1].trim()
      try {
        // Try parsing as JSON array
        const parsed = JSON.parse(rawEntrypoint)
        if (Array.isArray(parsed)) {
          return parsed
        }
      } catch {
        // Fallback: it's probably a plain string
        return [rawEntrypoint.replace(/["']/g, '')]
      }
    }

    // Match CMD with either a string or JSON array
    const cmdMatch = dockerfileContent.match(/CMD\s+(.*)/)
    if (cmdMatch) {
      const rawCmd = cmdMatch[1].trim()
      try {
        const parsed = JSON.parse(rawCmd)
        if (Array.isArray(parsed)) {
          return parsed
        }
      } catch {
        return [rawCmd.replace(/["']/g, '')]
      }
    }

    return ['sleep', 'infinity']
  }

  private handleRunnerWorkspaceStoppedOrArchivedStateOnDesiredStateStart: StateSyncHandler = async (
    workspace: Workspace,
  ): Promise<ShouldSyncAgain> => {
    //  check if workspace is assigned to a runner and if that runner is unschedulable
    //  if it is, move workspace to prevRunnerId, and set runnerId to null
    //  this will assign a new runner to the workspace and restore the workspace from the latest backup
    if (workspace.runnerId) {
      const runner = await this.runnerService.findOne(workspace.runnerId)
      if (runner.unschedulable) {
        //  check if workspace has a valid backup
        if (workspace.backupState !== BackupState.COMPLETED) {
          //  if not, keep workspace on the same runner
        } else {
          workspace.prevRunnerId = workspace.runnerId
          workspace.runnerId = null

          const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
            id: workspace.id,
          })
          workspaceToUpdate.prevRunnerId = workspace.runnerId
          workspaceToUpdate.runnerId = null
          await this.workspaceRepository.save(workspaceToUpdate)
        }
      }

      if (workspace.backupState === BackupState.COMPLETED) {
        const usageThreshold = 35
        const runningWorkspacesCount = await this.workspaceRepository.count({
          where: {
            runnerId: workspace.runnerId,
            state: WorkspaceState.STARTED,
          },
        })
        if (runningWorkspacesCount > usageThreshold) {
          //  TODO: usage should be based on compute usage

          const image = await this.imageService.getImageByName(workspace.image, workspace.organizationId)
          const availableRunners = await this.runnerService.findAvailableRunners({
            region: workspace.region,
            workspaceClass: workspace.class,
            imageRef: image.internalName,
          })
          const lessUsedRunners = availableRunners.filter((runner) => runner.id !== workspace.runnerId)

          //  temp workaround to move workspaces to less used runner
          if (lessUsedRunners.length > 0) {
            await this.workspaceRepository.update(workspace.id, {
              runnerId: null,
              prevRunnerId: workspace.runnerId,
            })
            try {
              const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)
              await runnerWorkspaceApi.removeDestroyed(workspace.id)
            } catch (e) {
              this.logger.error(
                `Failed to cleanup workspace ${workspace.id} on previous runner ${runner.id}:`,
                fromAxiosError(e),
              )
            }
            workspace.prevRunnerId = workspace.runnerId
            workspace.runnerId = null
          }
        }
      }
    }

    if (workspace.runnerId === null) {
      //  if workspace has no runner, check if backup is completed
      //  if not, set workspace to error
      //  if backup is completed, get random available runner and start workspace
      //  use the backup image to start the workspace

      if (workspace.backupState !== BackupState.COMPLETED) {
        await this.updateWorkspaceErrorState(workspace.id, 'Workspace has no runner and backup is not completed')
        return true
      }

      const registry = await this.dockerRegistryService.findOne(workspace.backupRegistryId)
      if (!registry) {
        throw new Error('No registry found for image')
      }

      const existingImages = workspace.existingBackupImages.map((existingImage) => existingImage.imageName)
      let validBackupImage
      let exists = false

      while (existingImages.length > 0) {
        try {
          if (!validBackupImage) {
            //  last image is the current image, so we don't need to check it
            //  just in case, we'll use the value from the backupImage property
            validBackupImage = workspace.backupImage
            existingImages.pop()
          } else {
            validBackupImage = existingImages.pop()
          }
          if (await this.dockerProvider.checkImageExistsInRegistry(validBackupImage, registry)) {
            exists = true
            break
          }
        } catch (error) {
          this.logger.error(
            `Failed to check if backup image ${workspace.backupImage} exists in registry ${registry.id}:`,
            fromAxiosError(error),
          )
        }
      }

      if (!exists) {
        await this.updateWorkspaceErrorState(workspace.id, 'No valid backup image found')
        return SYNC_AGAIN
      }

      const image = await this.imageService.getImageByName(workspace.image, workspace.organizationId)

      //  exclude the runner that the last runner workspace was on
      const availableRunners = (
        await this.runnerService.findAvailableRunners({
          region: workspace.region,
          workspaceClass: workspace.class,
          imageRef: image.internalName,
        })
      ).filter((runner) => runner.id != workspace.prevRunnerId)

      //  get random runner from available runners
      const randomRunnerIndex = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1) + min)
      const runnerId = availableRunners[randomRunnerIndex(0, availableRunners.length - 1)].id

      const runner = await this.runnerService.findOne(runnerId)

      const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)

      await runnerWorkspaceApi.create({
        id: workspace.id,
        image: validBackupImage,
        osUser: workspace.osUser,
        // TODO: organizationId: workspace.organizationId,
        userId: workspace.organizationId,
        storageQuota: workspace.disk,
        memoryQuota: workspace.mem,
        cpuQuota: workspace.cpu,
        // gpuQuota: workspace.gpu,
        env: workspace.env,
        // public: workspace.public,
        registry: {
          url: registry.url,
          username: registry.username,
          password: registry.password,
        },
      })

      await this.updateWorkspaceState(workspace.id, WorkspaceState.RESTORING, runnerId)
    } else {
      // if workspace has runner, start workspace
      const runner = await this.runnerService.findOne(workspace.runnerId)

      const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)

      await runnerWorkspaceApi.start(workspace.id)

      await this.updateWorkspaceState(workspace.id, WorkspaceState.STARTING)
      return SYNC_AGAIN
    }

    return SYNC_AGAIN
  }

  //  used to check if workspace is pulling image on runner and update workspace state accordingly
  private handleRunnerWorkspacePullingImageStateCheck: StateSyncHandler = async (
    workspace: Workspace,
  ): Promise<ShouldSyncAgain> => {
    //  edge case when workspace is being transferred to a new runner
    if (!workspace.runnerId) {
      return SYNC_AGAIN
    }

    const runner = await this.runnerService.findOne(workspace.runnerId)
    const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)
    const workspaceInfoResponse = await runnerWorkspaceApi.info(workspace.id)
    const workspaceInfo = workspaceInfoResponse.data

    if (workspaceInfo.state === RunnerWorkspaceState.SandboxStatePullingImage) {
      await this.updateWorkspaceState(workspace.id, WorkspaceState.PULLING_IMAGE)
    } else if (workspaceInfo.state === RunnerWorkspaceState.SandboxStateError) {
      await this.updateWorkspaceErrorState(workspace.id)
    } else {
      await this.updateWorkspaceState(workspace.id, WorkspaceState.STARTING)
    }

    return SYNC_AGAIN
  }

  //  used to check if workspace is started on runner and update workspace state accordingly
  //  also used to handle the case where a workspace is started on a runner and then transferred to a new runner
  private handleRunnerWorkspaceStartedStateCheck: StateSyncHandler = async (
    workspace: Workspace,
  ): Promise<ShouldSyncAgain> => {
    const runner = await this.runnerService.findOne(workspace.runnerId)
    const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)
    const workspaceInfoResponse = await runnerWorkspaceApi.info(workspace.id)
    const workspaceInfo = workspaceInfoResponse.data

    switch (workspaceInfo.state) {
      case RunnerWorkspaceState.SandboxStateStarted: {
        //  if previous backup state is error or completed, set backup state to none
        if ([BackupState.ERROR, BackupState.COMPLETED].includes(workspace.backupState)) {
          workspace.backupState = BackupState.NONE

          const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
            id: workspace.id,
          })
          workspaceToUpdate.state = WorkspaceState.STARTED
          workspaceToUpdate.backupState = BackupState.NONE
          await this.workspaceRepository.save(workspaceToUpdate)
        } else {
          await this.updateWorkspaceState(workspace.id, WorkspaceState.STARTED)
        }

        //  if workspace was transferred to a new runner, remove it from the old runner
        if (workspace.prevRunnerId) {
          const runner = await this.runnerService.findOne(workspace.prevRunnerId)
          if (!runner) {
            this.logger.warn(
              `Previously assigned runner ${workspace.prevRunnerId} for workspace ${workspace.id} not found`,
            )
            //  clear prevRunnerId to avoid trying to cleanup on a non-existent runner
            workspace.prevRunnerId = null

            const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
              id: workspace.id,
            })
            workspaceToUpdate.prevRunnerId = null
            await this.workspaceRepository.save(workspaceToUpdate)
            break
          }
          const runnerWorkspaceApi = this.runnerApiFactory.createWorkspaceApi(runner)
          try {
            // First try to destroy the workspace
            await runnerWorkspaceApi.destroy(workspace.id)

            // Wait for workspace to be destroyed before removing
            let retries = 0
            while (retries < 10) {
              try {
                const workspaceInfo = await runnerWorkspaceApi.info(workspace.id)
                if (workspaceInfo.data.state === RunnerWorkspaceState.SandboxStateDestroyed) {
                  break
                }
              } catch (e) {
                if (e.response?.status === 404) {
                  break // Workspace already gone
                }
                throw e
              }
              await new Promise((resolve) => setTimeout(resolve, 1000 * retries))
              retries++
            }

            // Finally remove the destroyed workspace
            await runnerWorkspaceApi.removeDestroyed(workspace.id)
            workspace.prevRunnerId = null

            const workspaceToUpdate = await this.workspaceRepository.findOneByOrFail({
              id: workspace.id,
            })
            workspaceToUpdate.prevRunnerId = null
            await this.workspaceRepository.save(workspaceToUpdate)
          } catch (e) {
            this.logger.error(
              `Failed to cleanup workspace ${workspace.id} on previous runner ${runner.id}:`,
              fromAxiosError(e),
            )
          }
        }
        break
      }
      case RunnerWorkspaceState.SandboxStateError: {
        await this.updateWorkspaceErrorState(workspace.id)
        break
      }
    }

    return SYNC_AGAIN
  }

  private async updateWorkspaceState(workspaceId: string, state: WorkspaceState, runnerId?: string | null | undefined) {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })
    if (workspace.state === state) {
      return
    }
    workspace.state = state
    if (runnerId !== undefined) {
      workspace.runnerId = runnerId
    }

    await this.workspaceRepository.save(workspace)
  }

  private async updateWorkspaceErrorState(workspaceId: string, errorReason?: string) {
    const workspace = await this.workspaceRepository.findOneByOrFail({
      id: workspaceId,
    })
    workspace.state = WorkspaceState.ERROR
    if (errorReason !== undefined) {
      workspace.errorReason = errorReason
    }
    await this.workspaceRepository.save(workspace)
  }

  @OnEvent(WorkspaceEvents.ARCHIVED)
  private async handleWorkspaceArchivedEvent(event: WorkspaceArchivedEvent) {
    this.syncInstanceState(event.workspace.id).catch(this.logger.error)
  }

  @OnEvent(WorkspaceEvents.DESTROYED)
  private async handleWorkspaceDestroyedEvent(event: WorkspaceDestroyedEvent) {
    this.syncInstanceState(event.workspace.id).catch(this.logger.error)
  }

  @OnEvent(WorkspaceEvents.STARTED)
  private async handleWorkspaceStartedEvent(event: WorkspaceStartedEvent) {
    this.syncInstanceState(event.workspace.id).catch(this.logger.error)
  }

  @OnEvent(WorkspaceEvents.STOPPED)
  private async handleWorkspaceStoppedEvent(event: WorkspaceStoppedEvent) {
    this.syncInstanceState(event.workspace.id).catch(this.logger.error)
  }

  @OnEvent(WorkspaceEvents.CREATED)
  private async handleWorkspaceCreatedEvent(event: WorkspaceCreatedEvent) {
    this.syncInstanceState(event.workspace.id).catch(this.logger.error)
  }
}
