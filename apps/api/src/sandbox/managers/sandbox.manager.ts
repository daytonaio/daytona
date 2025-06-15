/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, Not, Raw, Repository } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { RunnerApiFactory } from '../runner-api/runnerApi'
import { RunnerService } from '../services/runner.service'
import { EnumsSandboxState as RunnerSandboxState } from '@daytonaio/runner-api-client'
import { RunnerState } from '../enums/runner-state.enum'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { BackupState } from '../enums/backup-state.enum'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { SnapshotService } from '../services/snapshot.service'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/sandbox.constants'
import { DockerProvider } from '../docker/docker-provider'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { BuildInfo } from '../entities/build-info.entity'
import { CreateSandboxDTO } from '@daytonaio/runner-api-client'
import { fromAxiosError } from '../../common/utils/from-axios-error'
import { OnEvent } from '@nestjs/event-emitter'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxStoppedEvent } from '../events/sandbox-stopped.event'
import { SandboxStartedEvent } from '../events/sandbox-started.event'
import { SandboxArchivedEvent } from '../events/sandbox-archived.event'
import { SandboxDestroyedEvent } from '../events/sandbox-destroyed.event'
import { SandboxCreatedEvent } from '../events/sandbox-create.event'
import { OtelSpan } from '../../common/decorators/otel.decorator'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'

const SYNC_INSTANCE_STATE_LOCK_KEY = 'sync-instance-state-'
const SYNC_AGAIN = true
const DONT_SYNC_AGAIN = false
type ShouldSyncAgain = boolean
type StateSyncHandler = (sandbox: Sandbox) => Promise<ShouldSyncAgain>

@Injectable()
export class SandboxManager {
  private readonly logger = new Logger(SandboxManager.name)

  constructor(
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    @InjectRepository(SnapshotRunner)
    private readonly snapshotRunnerRepository: Repository<SnapshotRunner>,
    private readonly runnerService: RunnerService,
    private readonly runnerApiFactory: RunnerApiFactory,
    private readonly dockerRegistryService: DockerRegistryService,
    @InjectRedis() private readonly redis: Redis,
    private readonly snapshotService: SnapshotService,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly dockerProvider: DockerProvider,
  ) {}

  @Cron(CronExpression.EVERY_MINUTE, { name: 'auto-stop-check' })
  @OtelSpan()
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
        const sandboxs = await this.sandboxRepository.find({
          where: {
            runnerId: runner.id,
            organizationId: Not(SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION),
            state: SandboxState.STARTED,
            desiredState: SandboxDesiredState.STARTED,
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
          sandboxs.map(async (sandbox) => {
            const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + sandbox.id
            const acquired = await this.redisLockProvider.lock(lockKey, 30)
            if (!acquired) {
              return
            }

            try {
              sandbox.desiredState = SandboxDesiredState.STOPPED
              await this.sandboxRepository.save(sandbox)
              await this.redisLockProvider.unlock(lockKey)
              this.syncInstanceState(sandbox.id)
            } catch (error) {
              this.logger.error(`Error processing auto-stop state for sandbox ${sandbox.id}:`, fromAxiosError(error))
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
        const sandboxs = await this.sandboxRepository.find({
          where: {
            runnerId: runner.id,
            organizationId: Not(SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION),
            state: SandboxState.STOPPED,
            desiredState: SandboxDesiredState.STOPPED,
            pending: false,
            lastActivityAt: Raw((alias) => `${alias} < NOW() - INTERVAL '1 minute' * "autoArchiveInterval"`),
          },
          order: {
            lastBackupAt: 'ASC',
          },
          //  max 3 sandboxs can be archived at the same time on the same runner
          //  this is to prevent the runner from being overloaded
          take: 3,
        })

        await Promise.all(
          sandboxs.map(async (sandbox) => {
            const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + sandbox.id
            const acquired = await this.redisLockProvider.lock(lockKey, 30)
            if (!acquired) {
              return
            }

            try {
              sandbox.desiredState = SandboxDesiredState.ARCHIVED
              await this.sandboxRepository.save(sandbox)
              await this.redisLockProvider.unlock(lockKey)
              this.syncInstanceState(sandbox.id)
            } catch (error) {
              this.logger.error(`Error processing auto-archive state for sandbox ${sandbox.id}:`, fromAxiosError(error))
            }
          }),
        )
      }),
    )
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-states' })
  @OtelSpan()
  async syncStates(): Promise<void> {
    const lockKey = 'sync-states'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    const sandboxs = await this.sandboxRepository.find({
      where: {
        state: Not(In([SandboxState.DESTROYED, SandboxState.ERROR, SandboxState.BUILD_FAILED])),
        desiredState: Raw(
          () =>
            `"Sandbox"."desiredState"::text != "Sandbox"."state"::text AND "Sandbox"."desiredState"::text != 'archived'`,
        ),
      },
      take: 100,
      order: {
        lastActivityAt: 'DESC',
      },
    })

    await Promise.all(
      sandboxs.map(async (sandbox) => {
        this.syncInstanceState(sandbox.id)
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

    const runnersWith3InProgress = await this.sandboxRepository
      .createQueryBuilder('sandbox')
      .select('"runnerId"')
      .where('"sandbox"."state" = :state', { state: SandboxState.ARCHIVING })
      .groupBy('"runnerId"')
      .having('COUNT(*) >= 3')
      .getRawMany()

    const sandboxs = await this.sandboxRepository.find({
      where: [
        {
          state: SandboxState.ARCHIVING,
          desiredState: SandboxDesiredState.ARCHIVED,
        },
        {
          state: Not(
            In([SandboxState.ARCHIVED, SandboxState.DESTROYED, SandboxState.ERROR, SandboxState.BUILD_FAILED]),
          ),
          desiredState: SandboxDesiredState.ARCHIVED,
          runnerId: Not(In(runnersWith3InProgress.map((runner) => runner.runnerId))),
        },
      ],
      take: 100,
      order: {
        lastActivityAt: 'DESC',
      },
    })

    await Promise.all(
      sandboxs.map(async (sandbox) => {
        this.syncInstanceState(sandbox.id)
      }),
    )
    await this.redisLockProvider.unlock(lockKey)
  }

  async syncInstanceState(sandboxId: string): Promise<void> {
    //  prevent syncState cron from running multiple instances of the same sandbox
    const lockKey = SYNC_INSTANCE_STATE_LOCK_KEY + sandboxId
    const acquired = await this.redisLockProvider.lock(lockKey, 360)
    if (!acquired) {
      return
    }

    const sandbox = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })

    if (sandbox.state === SandboxState.ERROR || sandbox.state === SandboxState.BUILD_FAILED) {
      await this.redisLockProvider.unlock(lockKey)
      return
    }

    let shouldSyncAgain = DONT_SYNC_AGAIN

    try {
      switch (sandbox.desiredState) {
        case SandboxDesiredState.STARTED: {
          shouldSyncAgain = await this.handleSandboxDesiredStateStarted(sandbox)
          break
        }
        case SandboxDesiredState.STOPPED: {
          shouldSyncAgain = await this.handleSandboxDesiredStateStopped(sandbox)
          break
        }
        case SandboxDesiredState.DESTROYED: {
          shouldSyncAgain = await this.handleSandboxDesiredStateDestroyed(sandbox)
          break
        }
        case SandboxDesiredState.ARCHIVED: {
          shouldSyncAgain = await this.handleSandboxDesiredStateArchived(sandbox)
          break
        }
      }
    } catch (error) {
      if (error.code === 'ECONNRESET') {
        shouldSyncAgain = SYNC_AGAIN
      } else {
        this.logger.error(`Error processing desired state for sandbox ${sandboxId}:`, fromAxiosError(error))

        const sandbox = await this.sandboxRepository.findOneBy({
          id: sandboxId,
        })
        if (!sandbox) {
          //  edge case where sandbox is deleted while desired state is being processed
          return
        }
        await this.updateSandboxState(sandbox.id, SandboxState.ERROR, undefined, error.message || String(error))
      }
    }

    await this.redisLockProvider.unlock(lockKey)
    if (shouldSyncAgain) {
      this.syncInstanceState(sandboxId)
    }
  }

  private handleUnassignedBuildSandbox: StateSyncHandler = async (sandbox: Sandbox): Promise<ShouldSyncAgain> => {
    // Try to assign an available runner with the snapshot build
    let runnerId: string
    try {
      const runner = await this.runnerService.getRandomAvailableRunner({
        region: sandbox.region,
        sandboxClass: sandbox.class,
        snapshotRef: sandbox.buildInfo.snapshotRef,
      })
      runnerId = runner.id
    } catch (error) {
      // Continue to next assignment method
    }

    if (runnerId) {
      await this.updateSandboxState(sandbox.id, SandboxState.UNKNOWN, runnerId)
      return SYNC_AGAIN
    }

    // Try to assign an available runner that is currently building the snapshot
    const snapshotRunners = await this.runnerService.getSnapshotRunners(sandbox.buildInfo.snapshotRef)

    for (const snapshotRunner of snapshotRunners) {
      const runner = await this.runnerService.findOne(snapshotRunner.runnerId)
      if (runner.used < runner.capacity) {
        if (snapshotRunner.state === SnapshotRunnerState.BUILDING_SNAPSHOT) {
          await this.updateSandboxState(sandbox.id, SandboxState.BUILDING_SNAPSHOT, runner.id)
          return SYNC_AGAIN
        } else if (snapshotRunner.state === SnapshotRunnerState.ERROR) {
          await this.updateSandboxState(sandbox.id, SandboxState.BUILD_FAILED, undefined, snapshotRunner.errorReason)
          return DONT_SYNC_AGAIN
        }
      }
    }

    const excludedRunnerIds = await this.runnerService.getRunnersWithMultipleSnapshotsBuilding()

    // Try to assign a new available runner
    const runner = await this.runnerService.getRandomAvailableRunner({
      region: sandbox.region,
      sandboxClass: sandbox.class,
      excludedRunnerIds: excludedRunnerIds,
    })
    runnerId = runner.id

    this.buildOnRunner(sandbox.buildInfo, runnerId, sandbox.organizationId)

    await this.updateSandboxState(sandbox.id, SandboxState.BUILDING_SNAPSHOT, runnerId)
    await this.runnerService.recalculateRunnerUsage(runnerId)
    return SYNC_AGAIN
  }

  // Initiates the snapshot build on the runner and creates an SnapshotRunner depending on the result
  async buildOnRunner(buildInfo: BuildInfo, runnerId: string, organizationId: string) {
    const runner = await this.runnerService.findOne(runnerId)
    const runnerSnapshotApi = this.runnerApiFactory.createSnapshotApi(runner)

    let retries = 0

    while (retries < 10) {
      try {
        await runnerSnapshotApi.buildSnapshot({
          snapshot: buildInfo.snapshotRef,
          organizationId: organizationId,
          dockerfile: buildInfo.dockerfileContent,
          context: buildInfo.contextHashes,
        })
        break
      } catch (err) {
        if (err.code !== 'ECONNRESET') {
          await this.runnerService.createSnapshotRunner(
            runnerId,
            buildInfo.snapshotRef,
            SnapshotRunnerState.ERROR,
            err.message,
          )
          return
        }
      }
      retries++
      await new Promise((resolve) => setTimeout(resolve, retries * 1000))
    }

    if (retries === 10) {
      await this.runnerService.createSnapshotRunner(
        runnerId,
        buildInfo.snapshotRef,
        SnapshotRunnerState.ERROR,
        'Timeout while building',
      )
      return
    }

    const response = (await runnerSnapshotApi.snapshotExists(buildInfo.snapshotRef)).data
    let state = SnapshotRunnerState.BUILDING_SNAPSHOT
    if (response && response.exists) {
      state = SnapshotRunnerState.READY
    }

    await this.runnerService.createSnapshotRunner(runnerId, buildInfo.snapshotRef, state)
  }

  private handleSandboxDesiredStateArchived: StateSyncHandler = async (sandbox: Sandbox): Promise<ShouldSyncAgain> => {
    const lockKey = 'archive-lock-' + sandbox.runnerId
    if (!(await this.redisLockProvider.lock(lockKey, 10))) {
      return DONT_SYNC_AGAIN
    }

    const inProgressOnRunner = await this.sandboxRepository.find({
      where: {
        runnerId: sandbox.runnerId,
        state: In([SandboxState.ARCHIVING]),
      },
      order: {
        lastActivityAt: 'DESC',
      },
      take: 100,
    })

    //  if the sandbox is already in progress, continue
    if (!inProgressOnRunner.find((s) => s.id === sandbox.id)) {
      //  max 3 sandboxs can be archived at the same time on the same runner
      //  this is to prevent the runner from being overloaded
      if (inProgressOnRunner.length > 2) {
        await this.redisLockProvider.unlock(lockKey)
        return
      }
    }

    switch (sandbox.state) {
      case SandboxState.STOPPED: {
        await this.updateSandboxState(sandbox.id, SandboxState.ARCHIVING)
        //  fallthrough to archiving state
      }
      case SandboxState.ARCHIVING: {
        await this.redisLockProvider.unlock(lockKey)

        //  if the backup state is error, we need to retry the backup
        if (sandbox.backupState === BackupState.ERROR) {
          const archiveErrorRetryKey = 'archive-error-retry-' + sandbox.id
          const archiveErrorRetryCountRaw = await this.redis.get(archiveErrorRetryKey)
          const archiveErrorRetryCount = archiveErrorRetryCountRaw ? parseInt(archiveErrorRetryCountRaw) : 0
          //  if the archive error retry count is greater than 3, we need to mark the sandbox as error
          if (archiveErrorRetryCount > 3) {
            await this.updateSandboxState(sandbox.id, SandboxState.ERROR, undefined, 'Failed to archive sandbox')
            await this.redis.del(archiveErrorRetryKey)
            return DONT_SYNC_AGAIN
          }
          await this.redis.setex('archive-error-retry-' + sandbox.id, 720, String(archiveErrorRetryCount + 1))

          //  reset the backup state to pending to retry the backup
          await this.sandboxRepository.update(sandbox.id, {
            backupState: BackupState.PENDING,
          })

          return DONT_SYNC_AGAIN
        }

        // Check for timeout - if more than 30 minutes since last activity
        const thirtyMinutesAgo = new Date(Date.now() - 30 * 60 * 1000)
        if (sandbox.lastActivityAt < thirtyMinutesAgo) {
          await this.updateSandboxState(sandbox.id, SandboxState.ERROR, undefined, 'Archiving operation timed out')
          return DONT_SYNC_AGAIN
        }

        if (sandbox.backupState !== BackupState.COMPLETED) {
          return DONT_SYNC_AGAIN
        }

        //  when the backup is completed, destroy the sandbox on the runner
        //  and deassociate the sandbox from the runner
        const runner = await this.runnerService.findOne(sandbox.runnerId)
        const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)

        try {
          const sandboxInfoResponse = await runnerSandboxApi.info(sandbox.id)
          const sandboxInfo = sandboxInfoResponse.data
          switch (sandboxInfo.state) {
            case RunnerSandboxState.SandboxStateDestroying:
              //  wait until sandbox is destroyed on runner
              return SYNC_AGAIN
            case RunnerSandboxState.SandboxStateDestroyed:
              await this.updateSandboxState(sandbox.id, SandboxState.ARCHIVED, null)
              return DONT_SYNC_AGAIN
            default:
              await runnerSandboxApi.destroy(sandbox.id)
              return SYNC_AGAIN
          }
        } catch (error) {
          //  fail for errors other than sandbox not found or sandbox already destroyed
          if (
            !(
              (error.response?.data?.statusCode === 400 &&
                error.response?.data?.message.includes('Sandbox already destroyed')) ||
              error.response?.status === 404
            )
          ) {
            throw error
          }
          //  if the sandbox is already destroyed, do nothing
          await this.updateSandboxState(sandbox.id, SandboxState.ARCHIVED, null)
          return DONT_SYNC_AGAIN
        }
      }
    }

    return DONT_SYNC_AGAIN
  }

  private handleSandboxDesiredStateDestroyed: StateSyncHandler = async (sandbox: Sandbox): Promise<ShouldSyncAgain> => {
    if (sandbox.state === SandboxState.ARCHIVED) {
      await this.updateSandboxState(sandbox.id, SandboxState.DESTROYED)
      return DONT_SYNC_AGAIN
    }

    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      //  console.debug(`Runner ${runner.id} is not ready`);
      return DONT_SYNC_AGAIN
    }

    switch (sandbox.state) {
      case SandboxState.DESTROYED:
        return DONT_SYNC_AGAIN
      case SandboxState.DESTROYING: {
        // check if sandbox is destroyed
        const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)

        try {
          const sandboxInfoResponse = await runnerSandboxApi.info(sandbox.id)
          const sandboxInfo = sandboxInfoResponse.data
          if (
            sandboxInfo.state === RunnerSandboxState.SandboxStateDestroyed ||
            sandboxInfo.state === RunnerSandboxState.SandboxStateError
          ) {
            await runnerSandboxApi.removeDestroyed(sandbox.id)
          }
        } catch (e) {
          //  if the sandbox is not found on runner, it is already destroyed
          if (!e.response || e.response.status !== 404) {
            throw e
          }
        }

        await this.updateSandboxState(sandbox.id, SandboxState.DESTROYED)
        return SYNC_AGAIN
      }
      default: {
        // destroy sandbox
        try {
          const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)
          const sandboxInfoResponse = await runnerSandboxApi.info(sandbox.id)
          const sandboxInfo = sandboxInfoResponse.data
          if (sandboxInfo?.state === RunnerSandboxState.SandboxStateDestroyed) {
            return DONT_SYNC_AGAIN
          }
          await runnerSandboxApi.destroy(sandbox.id)
        } catch (e) {
          //  if the sandbox is not found on runner, it is already destroyed
          if (e.response.status !== 404) {
            throw e
          }
        }
        await this.updateSandboxState(sandbox.id, SandboxState.DESTROYING)
        return SYNC_AGAIN
      }
    }
  }

  private handleSandboxDesiredStateStarted: StateSyncHandler = async (sandbox: Sandbox): Promise<ShouldSyncAgain> => {
    switch (sandbox.state) {
      case SandboxState.PENDING_BUILD: {
        return this.handleUnassignedBuildSandbox(sandbox)
      }
      case SandboxState.BUILDING_SNAPSHOT: {
        return this.handleRunnerSandboxBuildingSnapshotStateOnDesiredStateStart(sandbox)
      }
      case SandboxState.UNKNOWN: {
        return this.handleRunnerSandboxUnknownStateOnDesiredStateStart(sandbox)
      }
      case SandboxState.ARCHIVED:
      case SandboxState.STOPPED: {
        return this.handleRunnerSandboxStoppedOrArchivedStateOnDesiredStateStart(sandbox)
      }
      case SandboxState.RESTORING:
      case SandboxState.CREATING: {
        return this.handleRunnerSandboxPullingSnapshotStateCheck(sandbox)
      }
      case SandboxState.PULLING_SNAPSHOT:
      case SandboxState.STARTING: {
        return this.handleRunnerSandboxStartedStateCheck(sandbox)
      }
      //  TODO: remove this case
      case SandboxState.ERROR: {
        //  TODO: remove this asap
        //  this was a temporary solution to recover from the false positive error state
        if (sandbox.id.startsWith('err_')) {
          return DONT_SYNC_AGAIN
        }
        const runner = await this.runnerService.findOne(sandbox.runnerId)
        const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)
        const sandboxInfoResponse = await runnerSandboxApi.info(sandbox.id)
        const sandboxInfo = sandboxInfoResponse.data
        if (sandboxInfo.state === RunnerSandboxState.SandboxStateStarted) {
          const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
            id: sandbox.id,
          })
          sandboxToUpdate.state = SandboxState.STARTED
          sandboxToUpdate.backupState = BackupState.NONE
          await this.sandboxRepository.save(sandboxToUpdate)
        }
      }
    }

    return DONT_SYNC_AGAIN
  }

  private handleSandboxDesiredStateStopped: StateSyncHandler = async (sandbox: Sandbox): Promise<ShouldSyncAgain> => {
    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      //  console.debug(`Runner ${runner.id} is not ready`);
      return DONT_SYNC_AGAIN
    }

    switch (sandbox.state) {
      case SandboxState.STARTED: {
        // stop sandbox
        const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)
        await runnerSandboxApi.stop(sandbox.id)
        await this.updateSandboxState(sandbox.id, SandboxState.STOPPING)
        //  sync states again immediately for sandbox
        return SYNC_AGAIN
      }
      case SandboxState.STOPPING: {
        // check if sandbox is stopped
        const runner = await this.runnerService.findOne(sandbox.runnerId)
        const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)
        const sandboxInfoResponse = await runnerSandboxApi.info(sandbox.id)
        const sandboxInfo = sandboxInfoResponse.data
        switch (sandboxInfo.state) {
          case RunnerSandboxState.SandboxStateStopped: {
            const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
              id: sandbox.id,
            })
            sandboxToUpdate.state = SandboxState.STOPPED
            sandboxToUpdate.backupState = BackupState.NONE
            await this.sandboxRepository.save(sandboxToUpdate)
            return SYNC_AGAIN
          }
          case RunnerSandboxState.SandboxStateError: {
            await this.updateSandboxState(
              sandbox.id,
              SandboxState.ERROR,
              undefined,
              'Sandbox is in error state on runner',
            )
            return DONT_SYNC_AGAIN
          }
        }
        return SYNC_AGAIN
      }
      case SandboxState.ERROR: {
        if (sandbox.id.startsWith('err_')) {
          return DONT_SYNC_AGAIN
        }
        const runner = await this.runnerService.findOne(sandbox.runnerId)
        const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)
        const sandboxInfoResponse = await runnerSandboxApi.info(sandbox.id)
        const sandboxInfo = sandboxInfoResponse.data
        if (sandboxInfo.state === RunnerSandboxState.SandboxStateStopped) {
          await this.updateSandboxState(sandbox.id, SandboxState.STOPPED)
        }
      }
    }

    return DONT_SYNC_AGAIN
  }

  private handleRunnerSandboxBuildingSnapshotStateOnDesiredStateStart: StateSyncHandler = async (
    sandbox: Sandbox,
  ): Promise<ShouldSyncAgain> => {
    const snapshotRunner = await this.runnerService.getSnapshotRunner(sandbox.runnerId, sandbox.buildInfo.snapshotRef)
    if (snapshotRunner) {
      switch (snapshotRunner.state) {
        case SnapshotRunnerState.READY: {
          // TODO: "UNKNOWN" should probably be changed to something else
          await this.updateSandboxState(sandbox.id, SandboxState.UNKNOWN)
          return SYNC_AGAIN
        }
        case SnapshotRunnerState.ERROR: {
          await this.updateSandboxState(sandbox.id, SandboxState.BUILD_FAILED, undefined, snapshotRunner.errorReason)
          return DONT_SYNC_AGAIN
        }
      }
    }
    if (!snapshotRunner || snapshotRunner.state === SnapshotRunnerState.BUILDING_SNAPSHOT) {
      // Sleep for a second and go back to syncing instance state
      await new Promise((resolve) => setTimeout(resolve, 1000))
      return SYNC_AGAIN
    }

    return DONT_SYNC_AGAIN
  }

  private handleRunnerSandboxUnknownStateOnDesiredStateStart: StateSyncHandler = async (
    sandbox: Sandbox,
  ): Promise<ShouldSyncAgain> => {
    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      //  console.debug(`Runner ${runner.id} is not ready`);
      return DONT_SYNC_AGAIN
    }

    let createSandboxDto: CreateSandboxDTO = {
      id: sandbox.id,
      osUser: sandbox.osUser,
      snapshot: '',
      // TODO: organizationId: sandbox.organizationId,
      userId: sandbox.organizationId,
      storageQuota: sandbox.disk,
      memoryQuota: sandbox.mem,
      cpuQuota: sandbox.cpu,
      // gpuQuota: sandbox.gpu,
      env: sandbox.env,
      // public: sandbox.public,
      volumes: sandbox.volumes,
    }

    if (!sandbox.buildInfo) {
      //  get internal snapshot name
      const snapshot = await this.snapshotService.getSnapshotByName(sandbox.snapshot, sandbox.organizationId)
      const internalSnapshotName = snapshot.internalName

      const registry = await this.dockerRegistryService.findOneBySnapshotImageName(
        internalSnapshotName,
        sandbox.organizationId,
      )
      if (!registry) {
        throw new Error('No registry found for snapshot')
      }

      createSandboxDto = {
        ...createSandboxDto,
        snapshot: internalSnapshotName,
        entrypoint: snapshot.entrypoint,
        registry: {
          url: registry.url,
          username: registry.username,
          password: registry.password,
        },
      }
    } else {
      createSandboxDto = {
        ...createSandboxDto,
        snapshot: sandbox.buildInfo.snapshotRef,
        entrypoint: this.getEntrypointFromDockerfile(sandbox.buildInfo.dockerfileContent),
      }
    }

    const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)
    await runnerSandboxApi.create(createSandboxDto)
    await this.updateSandboxState(sandbox.id, SandboxState.CREATING)
    //  sync states again immediately for sandbox
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

  private handleRunnerSandboxStoppedOrArchivedStateOnDesiredStateStart: StateSyncHandler = async (
    sandbox: Sandbox,
  ): Promise<ShouldSyncAgain> => {
    //  check if sandbox is assigned to a runner and if that runner is unschedulable
    //  if it is, move sandbox to prevRunnerId, and set runnerId to null
    //  this will assign a new runner to the sandbox and restore the sandbox from the latest backup
    if (sandbox.runnerId) {
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      if (runner.unschedulable) {
        //  check if sandbox has a valid backup
        if (sandbox.backupState !== BackupState.COMPLETED) {
          //  if not, keep sandbox on the same runner
        } else {
          sandbox.prevRunnerId = sandbox.runnerId
          sandbox.runnerId = null

          const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
            id: sandbox.id,
          })
          sandboxToUpdate.prevRunnerId = sandbox.runnerId
          sandboxToUpdate.runnerId = null
          await this.sandboxRepository.save(sandboxToUpdate)
        }
      }

      if (sandbox.backupState === BackupState.COMPLETED) {
        const usageThreshold = 35
        const runningSandboxsCount = await this.sandboxRepository.count({
          where: {
            runnerId: sandbox.runnerId,
            state: SandboxState.STARTED,
          },
        })
        if (runningSandboxsCount > usageThreshold) {
          //  TODO: usage should be based on compute usage

          const snapshot = await this.snapshotService.getSnapshotByName(sandbox.snapshot, sandbox.organizationId)
          const availableRunners = await this.runnerService.findAvailableRunners({
            region: sandbox.region,
            sandboxClass: sandbox.class,
            snapshotRef: snapshot.internalName,
          })
          const lessUsedRunners = availableRunners.filter((runner) => runner.id !== sandbox.runnerId)

          //  temp workaround to move sandboxs to less used runner
          if (lessUsedRunners.length > 0) {
            await this.sandboxRepository.update(sandbox.id, {
              runnerId: null,
              prevRunnerId: sandbox.runnerId,
            })
            try {
              const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)
              await runnerSandboxApi.removeDestroyed(sandbox.id)
            } catch (e) {
              this.logger.error(
                `Failed to cleanup sandbox ${sandbox.id} on previous runner ${runner.id}:`,
                fromAxiosError(e),
              )
            }
            sandbox.prevRunnerId = sandbox.runnerId
            sandbox.runnerId = null
          }
        }
      }
    }

    if (sandbox.runnerId === null) {
      //  if sandbox has no runner, check if backup is completed
      //  if not, set sandbox to error
      //  if backup is completed, get random available runner and start sandbox
      //  use the backup snapshot to start the sandbox

      if (sandbox.backupState !== BackupState.COMPLETED) {
        await this.updateSandboxState(
          sandbox.id,
          SandboxState.ERROR,
          undefined,
          'Sandbox has no runner and backup is not completed',
        )
        return true
      }

      const registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)
      if (!registry) {
        throw new Error('No registry found for snapshot')
      }

      const existingSnapshots = sandbox.existingBackupSnapshots.map((existingSnapshot) => existingSnapshot.snapshotName)
      let validBackupSnapshot
      let exists = false

      while (existingSnapshots.length > 0) {
        try {
          if (!validBackupSnapshot) {
            //  last snapshot is the current snapshot, so we don't need to check it
            //  just in case, we'll use the value from the backupSnapshot property
            validBackupSnapshot = sandbox.backupSnapshot
            existingSnapshots.pop()
          } else {
            validBackupSnapshot = existingSnapshots.pop()
          }
          if (await this.dockerProvider.checkImageExistsInRegistry(validBackupSnapshot, registry)) {
            exists = true
            break
          }
        } catch (error) {
          this.logger.error(
            `Failed to check if backup snapshot ${sandbox.backupSnapshot} exists in registry ${registry.id}:`,
            fromAxiosError(error),
          )
        }
      }

      if (!exists) {
        await this.updateSandboxState(sandbox.id, SandboxState.ERROR, undefined, 'No valid backup snapshot found')
        return SYNC_AGAIN
      }

      const snapshot = await this.snapshotService.getSnapshotByName(sandbox.snapshot, sandbox.organizationId)

      //  exclude the runner that the last runner sandbox was on
      const availableRunners = (
        await this.runnerService.findAvailableRunners({
          region: sandbox.region,
          sandboxClass: sandbox.class,
          snapshotRef: snapshot.internalName,
        })
      ).filter((runner) => runner.id != sandbox.prevRunnerId)

      //  get random runner from available runners
      const randomRunnerIndex = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1) + min)
      const runnerId = availableRunners[randomRunnerIndex(0, availableRunners.length - 1)].id

      const runner = await this.runnerService.findOne(runnerId)

      const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)

      await runnerSandboxApi.create({
        id: sandbox.id,
        snapshot: validBackupSnapshot,
        osUser: sandbox.osUser,
        // TODO: organizationId: sandbox.organizationId,
        userId: sandbox.organizationId,
        storageQuota: sandbox.disk,
        memoryQuota: sandbox.mem,
        cpuQuota: sandbox.cpu,
        // gpuQuota: sandbox.gpu,
        env: sandbox.env,
        // public: sandbox.public,
        registry: {
          url: registry.url,
          username: registry.username,
          password: registry.password,
        },
      })

      await this.updateSandboxState(sandbox.id, SandboxState.RESTORING, runnerId)
    } else {
      // if sandbox has runner, start sandbox
      const runner = await this.runnerService.findOne(sandbox.runnerId)

      const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)

      await runnerSandboxApi.start(sandbox.id)

      await this.updateSandboxState(sandbox.id, SandboxState.STARTING)
      return SYNC_AGAIN
    }

    return SYNC_AGAIN
  }

  //  used to check if sandbox is pulling snapshot on runner and update sandbox state accordingly
  private handleRunnerSandboxPullingSnapshotStateCheck: StateSyncHandler = async (
    sandbox: Sandbox,
  ): Promise<ShouldSyncAgain> => {
    //  edge case when sandbox is being transferred to a new runner
    if (!sandbox.runnerId) {
      return SYNC_AGAIN
    }

    const runner = await this.runnerService.findOne(sandbox.runnerId)
    const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)
    const sandboxInfoResponse = await runnerSandboxApi.info(sandbox.id)
    const sandboxInfo = sandboxInfoResponse.data

    if (sandboxInfo.state === RunnerSandboxState.SandboxStatePullingSnapshot) {
      await this.updateSandboxState(sandbox.id, SandboxState.PULLING_SNAPSHOT)
    } else if (sandboxInfo.state === RunnerSandboxState.SandboxStateError) {
      await this.updateSandboxState(sandbox.id, SandboxState.ERROR)
    } else {
      await this.updateSandboxState(sandbox.id, SandboxState.STARTING)
    }

    return SYNC_AGAIN
  }

  //  used to check if sandbox is started on runner and update sandbox state accordingly
  //  also used to handle the case where a sandbox is started on a runner and then transferred to a new runner
  private handleRunnerSandboxStartedStateCheck: StateSyncHandler = async (
    sandbox: Sandbox,
  ): Promise<ShouldSyncAgain> => {
    const runner = await this.runnerService.findOne(sandbox.runnerId)
    const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)
    const sandboxInfoResponse = await runnerSandboxApi.info(sandbox.id)
    const sandboxInfo = sandboxInfoResponse.data

    switch (sandboxInfo.state) {
      case RunnerSandboxState.SandboxStateStarted: {
        //  if previous backup state is error or completed, set backup state to none
        if ([BackupState.ERROR, BackupState.COMPLETED].includes(sandbox.backupState)) {
          sandbox.backupState = BackupState.NONE

          const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
            id: sandbox.id,
          })
          sandboxToUpdate.state = SandboxState.STARTED
          sandboxToUpdate.backupState = BackupState.NONE
          await this.sandboxRepository.save(sandboxToUpdate)
        } else {
          await this.updateSandboxState(sandbox.id, SandboxState.STARTED)
        }

        //  if sandbox was transferred to a new runner, remove it from the old runner
        if (sandbox.prevRunnerId) {
          const runner = await this.runnerService.findOne(sandbox.prevRunnerId)
          if (!runner) {
            this.logger.warn(`Previously assigned runner ${sandbox.prevRunnerId} for sandbox ${sandbox.id} not found`)
            //  clear prevRunnerId to avoid trying to cleanup on a non-existent runner
            sandbox.prevRunnerId = null

            const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
              id: sandbox.id,
            })
            sandboxToUpdate.prevRunnerId = null
            await this.sandboxRepository.save(sandboxToUpdate)
            break
          }
          const runnerSandboxApi = this.runnerApiFactory.createSandboxApi(runner)
          try {
            // First try to destroy the sandbox
            await runnerSandboxApi.destroy(sandbox.id)

            // Wait for sandbox to be destroyed before removing
            let retries = 0
            while (retries < 10) {
              try {
                const sandboxInfo = await runnerSandboxApi.info(sandbox.id)
                if (sandboxInfo.data.state === RunnerSandboxState.SandboxStateDestroyed) {
                  break
                }
              } catch (e) {
                if (e.response?.status === 404) {
                  break // Sandbox already gone
                }
                throw e
              }
              await new Promise((resolve) => setTimeout(resolve, 1000 * retries))
              retries++
            }

            // Finally remove the destroyed sandbox
            await runnerSandboxApi.removeDestroyed(sandbox.id)
            sandbox.prevRunnerId = null

            const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
              id: sandbox.id,
            })
            sandboxToUpdate.prevRunnerId = null
            await this.sandboxRepository.save(sandboxToUpdate)
          } catch (e) {
            this.logger.error(
              `Failed to cleanup sandbox ${sandbox.id} on previous runner ${runner.id}:`,
              fromAxiosError(e),
            )
          }
        }
        break
      }
      case RunnerSandboxState.SandboxStateError: {
        await this.updateSandboxState(sandbox.id, SandboxState.ERROR)
        break
      }
    }

    return SYNC_AGAIN
  }

  private async updateSandboxState(
    sandboxId: string,
    state: SandboxState,
    runnerId?: string | null | undefined,
    errorReason?: string,
  ) {
    const sandbox = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })
    if (sandbox.state === state && sandbox.runnerId === runnerId && sandbox.errorReason === errorReason) {
      return
    }
    sandbox.state = state
    if (runnerId !== undefined) {
      sandbox.runnerId = runnerId
    }
    if (errorReason !== undefined) {
      sandbox.errorReason = errorReason
    }

    await this.sandboxRepository.save(sandbox)
  }

  @OnEvent(SandboxEvents.ARCHIVED)
  private async handleSandboxArchivedEvent(event: SandboxArchivedEvent) {
    this.syncInstanceState(event.sandbox.id).catch(this.logger.error)
  }

  @OnEvent(SandboxEvents.DESTROYED)
  private async handleSandboxDestroyedEvent(event: SandboxDestroyedEvent) {
    this.syncInstanceState(event.sandbox.id).catch(this.logger.error)
  }

  @OnEvent(SandboxEvents.STARTED)
  private async handleSandboxStartedEvent(event: SandboxStartedEvent) {
    this.syncInstanceState(event.sandbox.id).catch(this.logger.error)
  }

  @OnEvent(SandboxEvents.STOPPED)
  private async handleSandboxStoppedEvent(event: SandboxStoppedEvent) {
    this.syncInstanceState(event.sandbox.id).catch(this.logger.error)
  }

  @OnEvent(SandboxEvents.CREATED)
  private async handleSandboxCreatedEvent(event: SandboxCreatedEvent) {
    this.syncInstanceState(event.sandbox.id).catch(this.logger.error)
  }
}
