/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException, OnApplicationShutdown } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, LessThan, Not, Repository } from 'typeorm'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { Runner } from '../entities/runner.entity'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { RunnerState } from '../enums/runner-state.enum'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { v4 as uuidv4 } from 'uuid'
import { RunnerNotReadyError } from '../errors/runner-not-ready.error'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { OrganizationService } from '../../organization/services/organization.service'
import { BuildInfo } from '../entities/build-info.entity'
import { fromAxiosError } from '../../common/utils/from-axios-error'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { RunnerService } from '../services/runner.service'
import { TrackableJobExecutions } from '../../common/interfaces/trackable-job-executions'
import { TrackJobExecution } from '../../common/decorators/track-job-execution.decorator'
import { setTimeout as sleep } from 'timers/promises'
import { TypedConfigService } from '../../config/typed-config.service'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'
import { RunnerAdapterFactory, RunnerSnapshotInfo } from '../runner-adapter/runnerAdapter'
import { OnEvent } from '@nestjs/event-emitter'
import { SnapshotEvents } from '../constants/snapshot-events'
import { SnapshotCreatedEvent } from '../events/snapshot-created.event'
import { SnapshotService } from '../services/snapshot.service'

const SYNC_AGAIN = 'sync-again'
const DONT_SYNC_AGAIN = 'dont-sync-again'
type SyncState = typeof SYNC_AGAIN | typeof DONT_SYNC_AGAIN

@Injectable()
export class SnapshotManager implements TrackableJobExecutions, OnApplicationShutdown {
  activeJobs = new Set<string>()

  private readonly logger = new Logger(SnapshotManager.name)
  //  generate a unique instance id used to ensure only one instance of the worker is handing the
  //  snapshot activation
  private readonly instanceId = uuidv4()

  constructor(
    @InjectRedis() private readonly redis: Redis,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    @InjectRepository(SnapshotRunner)
    private readonly snapshotRunnerRepository: Repository<SnapshotRunner>,
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    @InjectRepository(BuildInfo)
    private readonly buildInfoRepository: Repository<BuildInfo>,
    private readonly runnerService: RunnerService,
    private readonly dockerRegistryService: DockerRegistryService,
    private readonly runnerAdapterFactory: RunnerAdapterFactory,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly organizationService: OrganizationService,
    private readonly configService: TypedConfigService,
    private readonly snapshotService: SnapshotService,
  ) {}

  async onApplicationShutdown() {
    //  wait for all active jobs to finish
    while (this.activeJobs.size > 0) {
      this.logger.log(`Waiting for ${this.activeJobs.size} active jobs to finish`)
      await sleep(1000)
    }
  }

  @Cron(CronExpression.EVERY_5_SECONDS, { name: 'sync-runner-snapshots', waitForCompletion: true })
  @TrackJobExecution()
  @LogExecution('sync-runner-snapshots')
  @WithInstrumentation()
  async syncRunnerSnapshots() {
    const lockKey = 'sync-runner-snapshots-lock'
    const lockTtl = 10 * 60 // seconds (10 min)
    if (!(await this.redisLockProvider.lock(lockKey, lockTtl))) {
      return
    }

    const skip = (await this.redis.get('sync-runner-snapshots-skip')) || 0

    const snapshots = await this.snapshotRepository
      .createQueryBuilder('snapshot')
      .innerJoin('organization', 'org', 'org.id = snapshot.organizationId')
      .where('snapshot.state = :snapshotState', { snapshotState: SnapshotState.ACTIVE })
      .andWhere('org.suspended = false')
      .orderBy('snapshot.createdAt', 'ASC')
      .take(100)
      .skip(Number(skip))
      .getMany()

    if (snapshots.length === 0) {
      await this.redisLockProvider.unlock(lockKey)
      await this.redis.set('sync-runner-snapshots-skip', 0)
      return
    }

    await this.redis.set('sync-runner-snapshots-skip', Number(skip) + snapshots.length)

    const results = await Promise.allSettled(
      snapshots.map(async (snapshot) => {
        const regions = await this.snapshotService.getSnapshotRegions(snapshot.id)

        const sharedRegionIds = regions.filter((r) => r.organizationId === null).map((r) => r.id)
        const organizationRegionIds = regions
          .filter((r) => r.organizationId === snapshot.organizationId)
          .map((r) => r.id)

        return this.propagateSnapshotToRunners(snapshot, sharedRegionIds, organizationRegionIds)
      }),
    )

    // Log all promise errors
    results.forEach((result) => {
      if (result.status === 'rejected') {
        this.logger.error(`Error propagating snapshot to runners: ${fromAxiosError(result.reason)}`)
      }
    })

    await this.redisLockProvider.unlock(lockKey)
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'sync-runner-snapshot-states', waitForCompletion: true })
  @TrackJobExecution()
  @LogExecution('sync-runner-snapshot-states')
  @WithInstrumentation()
  async syncRunnerSnapshotStates() {
    //  this approach is not ideal, as if the number of runners is large, this will take a long time
    //  also, if some snapshots stuck in a "pulling" state, they will infest the queue
    //  todo: find a better approach

    const lockKey = 'sync-runner-snapshot-states-lock'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    const runnerSnapshots = await this.snapshotRunnerRepository
      .createQueryBuilder('snapshotRunner')
      .where({
        state: In([
          SnapshotRunnerState.PULLING_SNAPSHOT,
          SnapshotRunnerState.BUILDING_SNAPSHOT,
          SnapshotRunnerState.REMOVING,
        ]),
      })
      .orderBy('RANDOM()')
      .take(100)
      .getMany()

    await Promise.allSettled(
      runnerSnapshots.map((snapshotRunner) => {
        return this.syncRunnerSnapshotState(snapshotRunner).catch((err) => {
          if (err.code !== 'ECONNRESET') {
            if (err instanceof RunnerNotReadyError) {
              this.logger.debug(
                `Runner ${snapshotRunner.runnerId} is not ready while trying to sync snapshot runner ${snapshotRunner.id}: ${err}`,
              )
              return
            }
            this.logger.error(`Error syncing runner snapshot state ${snapshotRunner.id}: ${fromAxiosError(err)}`)
            this.snapshotRunnerRepository.update(snapshotRunner.id, {
              state: SnapshotRunnerState.ERROR,
              errorReason: fromAxiosError(err).message,
            })
          }
        })
      }),
    )

    await this.redisLockProvider.unlock(lockKey)
  }

  async syncRunnerSnapshotState(snapshotRunner: SnapshotRunner): Promise<void> {
    const runner = await this.runnerRepository.findOne({
      where: {
        id: snapshotRunner.runnerId,
      },
    })
    if (!runner) {
      //  cleanup the snapshot runner record if the runner is not found
      //  this can happen if the runner is deleted from the database without cleaning up the snapshot runners
      await this.snapshotRunnerRepository.delete(snapshotRunner.id)
      this.logger.warn(
        `Runner ${snapshotRunner.runnerId} not found while trying to process snapshot runner ${snapshotRunner.id}. Snapshot runner has been removed.`,
      )
      return
    }
    if (runner.state !== RunnerState.READY) {
      //  todo: handle timeout policy
      //  for now just remove the snapshot runner record if the runner is not ready
      await this.snapshotRunnerRepository.delete(snapshotRunner.id)

      throw new RunnerNotReadyError(`Runner ${runner.id} is not ready`)
    }

    switch (snapshotRunner.state) {
      case SnapshotRunnerState.PULLING_SNAPSHOT:
        await this.handleSnapshotRunnerStatePullingSnapshot(snapshotRunner, runner)
        break
      case SnapshotRunnerState.BUILDING_SNAPSHOT:
        await this.handleSnapshotRunnerStateBuildingSnapshot(snapshotRunner, runner)
        break
      case SnapshotRunnerState.REMOVING:
        await this.handleSnapshotRunnerStateRemoving(snapshotRunner, runner)
        break
    }
  }

  async propagateSnapshotToRunners(snapshot: Snapshot, sharedRegionIds: string[], organizationRegionIds: string[]) {
    //  todo: remove try catch block and implement error handling
    try {
      //  get all runners in the regions to propagate to
      const runners = await this.runnerRepository.find({
        where: {
          state: RunnerState.READY,
          unschedulable: Not(true),
          region: In([...sharedRegionIds, ...organizationRegionIds]),
        },
      })

      const sharedRunners = runners.filter((runner) => sharedRegionIds.includes(runner.region))
      const sharedRunnerIds = sharedRunners.map((runner) => runner.id)

      const organizationRunners = runners.filter((runner) => organizationRegionIds.includes(runner.region))
      const organizationRunnerIds = organizationRunners.map((runner) => runner.id)

      //  get all runners where the snapshot is already propagated to (or in progress)
      const sharedSnapshotRunners = await this.snapshotRunnerRepository.find({
        where: {
          snapshotRef: snapshot.ref,
          state: In([SnapshotRunnerState.READY, SnapshotRunnerState.PULLING_SNAPSHOT]),
          runnerId: In(sharedRunnerIds),
        },
      })
      const sharedSnapshotRunnersDistinctRunnersIds = new Set(
        sharedSnapshotRunners.map((snapshotRunner) => snapshotRunner.runnerId),
      )

      const organizationSnapshotRunners = await this.snapshotRunnerRepository.find({
        where: {
          snapshotRef: snapshot.ref,
          state: In([SnapshotRunnerState.READY, SnapshotRunnerState.PULLING_SNAPSHOT]),
          runnerId: In(organizationRunnerIds),
        },
      })
      const organizationSnapshotRunnersDistinctRunnersIds = new Set(
        organizationSnapshotRunners.map((snapshotRunner) => snapshotRunner.runnerId),
      )

      //  get all runners where the snapshot is not propagated to
      const unallocatedSharedRunners = sharedRunners.filter(
        (runner) => !sharedSnapshotRunnersDistinctRunnersIds.has(runner.id),
      )
      const unallocatedOrganizationRunners = organizationRunners.filter(
        (runner) => !organizationSnapshotRunnersDistinctRunnersIds.has(runner.id),
      )

      const runnersToPropagateTo: Runner[] = []

      // propagate the snapshot to all organization runners
      runnersToPropagateTo.push(...unallocatedOrganizationRunners)

      // respect the propagation limit for shared runners
      const sharedRunnersPropagateLimit = Math.max(
        0,
        Math.ceil(sharedRunners.length / 3) - sharedSnapshotRunnersDistinctRunnersIds.size,
      )
      runnersToPropagateTo.push(
        ...unallocatedSharedRunners.sort(() => Math.random() - 0.5).slice(0, sharedRunnersPropagateLimit),
      )

      if (runnersToPropagateTo.length === 0) {
        return
      }

      // regionId -> registry
      const internalRegistriesMap = new Map<string, DockerRegistry>()

      for (const regionId of [...sharedRegionIds, ...organizationRegionIds]) {
        const registry = await this.dockerRegistryService.findInternalRegistryBySnapshotRef(snapshot.ref, regionId)
        if (registry) {
          internalRegistriesMap.set(regionId, registry)
        }
      }

      const results = await Promise.allSettled(
        runnersToPropagateTo.map(async (runner) => {
          const registry = internalRegistriesMap.get(runner.region)
          if (!registry) {
            throw new Error(`No internal registry found for snapshot ${snapshot.ref} in region ${runner.region}`)
          }

          const snapshotRunner = await this.runnerService.getSnapshotRunner(runner.id, snapshot.ref)

          try {
            if (!snapshotRunner) {
              await this.runnerService.createSnapshotRunnerEntry(
                runner.id,
                snapshot.ref,
                SnapshotRunnerState.PULLING_SNAPSHOT,
              )
              await this.pullSnapshotRunnerWithRetries(runner, snapshot.ref, registry)
            } else if (snapshotRunner.state === SnapshotRunnerState.PULLING_SNAPSHOT) {
              await this.handleSnapshotRunnerStatePullingSnapshot(snapshotRunner, runner)
            }
          } catch (err) {
            this.logger.error(`Error propagating snapshot to runner ${runner.id}: ${fromAxiosError(err)}`)
            snapshotRunner.state = SnapshotRunnerState.ERROR
            snapshotRunner.errorReason = err.message
            await this.snapshotRunnerRepository.update(snapshotRunner.id, snapshotRunner)
          }
        }),
      )

      results.forEach((result) => {
        if (result.status === 'rejected') {
          this.logger.error(result.reason)
        }
      })
    } catch (err) {
      this.logger.error(err)
    }
  }

  async pullSnapshotRunnerWithRetries(
    runner: Runner,
    snapshotRef: string,
    registry?: DockerRegistry,
    destinationRegistry?: DockerRegistry,
    destinationRef?: string,
  ) {
    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    let retries = 0
    while (retries < 10) {
      try {
        await runnerAdapter.pullSnapshot(snapshotRef, registry, destinationRegistry, destinationRef)
        return
      } catch (err) {
        if (err.code !== 'ECONNRESET') {
          throw err
        }
        if (++retries >= 10) {
          throw err
        }
        await new Promise((resolve) => setTimeout(resolve, retries * 1000))
      }
    }
  }

  async handleSnapshotRunnerStatePullingSnapshot(snapshotRunner: SnapshotRunner, runner: Runner) {
    const runnerAdapter = await this.runnerAdapterFactory.create(runner)
    const exists = await runnerAdapter.snapshotExists(snapshotRunner.snapshotRef)
    if (exists) {
      snapshotRunner.state = SnapshotRunnerState.READY
      await this.snapshotRunnerRepository.save(snapshotRunner)
      return
    }

    const timeoutMinutes = 60
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - snapshotRunner.updatedAt.getTime() > timeoutMs) {
      snapshotRunner.state = SnapshotRunnerState.ERROR
      snapshotRunner.errorReason = 'Timeout while pulling snapshot to runner'
      await this.snapshotRunnerRepository.save(snapshotRunner)
      return
    }

    const retryTimeoutMinutes = 10
    const retryTimeoutMs = retryTimeoutMinutes * 60 * 1000
    if (Date.now() - snapshotRunner.createdAt.getTime() > retryTimeoutMs) {
      const registry = await this.dockerRegistryService.findInternalRegistryBySnapshotRef(
        snapshotRunner.snapshotRef,
        runner.region,
      )
      if (!registry) {
        throw new Error(
          `No internal registry found for snapshot ${snapshotRunner.snapshotRef} in region ${runner.region}`,
        )
      }
      await this.pullSnapshotRunnerWithRetries(runner, snapshotRunner.snapshotRef, registry)
      return
    }
  }

  async handleSnapshotRunnerStateBuildingSnapshot(snapshotRunner: SnapshotRunner, runner: Runner) {
    const runnerAdapter = await this.runnerAdapterFactory.create(runner)
    const exists = await runnerAdapter.snapshotExists(snapshotRunner.snapshotRef)
    if (exists) {
      snapshotRunner.state = SnapshotRunnerState.READY
      await this.snapshotRunnerRepository.save(snapshotRunner)
      return
    }
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'check-snapshot-cleanup' })
  @TrackJobExecution()
  @LogExecution('check-snapshot-cleanup')
  @WithInstrumentation()
  async checkSnapshotCleanup() {
    const lockKey = 'check-snapshot-cleanup-lock'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    const snapshots = await this.snapshotRepository.find({
      where: {
        state: SnapshotState.REMOVING,
      },
    })

    await Promise.all(
      snapshots.map(async (snapshot) => {
        const countActiveSnapshots = await this.snapshotRepository.count({
          where: {
            state: SnapshotState.ACTIVE,
            ref: snapshot.ref,
          },
        })

        // Only remove snapshot runners if no other snapshots depend on them
        if (countActiveSnapshots === 0) {
          await this.snapshotRunnerRepository.update(
            {
              snapshotRef: snapshot.ref,
            },
            {
              state: SnapshotRunnerState.REMOVING,
            },
          )
        }

        await this.snapshotRepository.remove(snapshot)
      }),
    )

    await this.redisLockProvider.unlock(lockKey)
  }

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'check-snapshot-state' })
  @TrackJobExecution()
  @LogExecution('check-snapshot-state')
  @WithInstrumentation()
  async checkSnapshotState() {
    //  the first time the snapshot is created it needs to be pushed to the internal registry
    //  before propagating to the runners
    //  this cron job will process the snapshot states until the snapshot is active (or error)

    //  get all snapshots
    const snapshots = await this.snapshotRepository.find({
      where: {
        state: Not(In([SnapshotState.ACTIVE, SnapshotState.ERROR, SnapshotState.BUILD_FAILED, SnapshotState.INACTIVE])),
      },
    })

    await Promise.all(
      snapshots.map(async (snapshot) => {
        this.syncSnapshotState(snapshot.id)
      }),
    )
  }

  async syncSnapshotState(snapshotId: string): Promise<void> {
    const lockKey = `sync-snapshot-state-${snapshotId}`
    if (!(await this.redisLockProvider.lock(lockKey, 720))) {
      return
    }

    const snapshot = await this.snapshotRepository.findOne({
      where: { id: snapshotId },
    })

    if (
      !snapshot ||
      [SnapshotState.ACTIVE, SnapshotState.ERROR, SnapshotState.BUILD_FAILED, SnapshotState.INACTIVE].includes(
        snapshot.state,
      )
    ) {
      await this.redisLockProvider.unlock(lockKey)
      return
    }

    let syncState = DONT_SYNC_AGAIN

    try {
      switch (snapshot.state) {
        case SnapshotState.PENDING:
          syncState = await this.handleSnapshotStatePending(snapshot)
          break
        case SnapshotState.PULLING:
        case SnapshotState.BUILDING:
          syncState = await this.handleCheckInitialRunnerSnapshot(snapshot)
          break
        case SnapshotState.REMOVING:
          syncState = await this.handleSnapshotStateRemoving(snapshot)
          break
      }
    } catch (error) {
      if (error.code === 'ECONNRESET') {
        syncState = SYNC_AGAIN
      } else {
        const message = error.message || String(error)
        await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, message)
      }
    }

    await this.redisLockProvider.unlock(lockKey)
    if (syncState === SYNC_AGAIN) {
      this.syncSnapshotState(snapshotId)
    }
  }

  async handleSnapshotRunnerStateRemoving(snapshotRunner: SnapshotRunner, runner: Runner) {
    if (!runner) {
      //  generally this should not happen
      //  in case the runner has been deleted from the database, delete the snapshot runner record
      const errorMessage = `Runner not found while trying to remove snapshot ${snapshotRunner.snapshotRef} from runner ${snapshotRunner.runnerId}`
      this.logger.warn(errorMessage)

      this.snapshotRunnerRepository.delete(snapshotRunner.id).catch((err) => {
        this.logger.error(fromAxiosError(err))
      })
      return
    }
    if (!snapshotRunner.snapshotRef) {
      //  this should never happen
      //  remove the snapshot runner record (it will be recreated again by the snapshot propagation job)
      this.logger.warn(`Internal snapshot name not found for snapshot runner ${snapshotRunner.id}`)
      this.snapshotRunnerRepository.delete(snapshotRunner.id).catch((err) => {
        this.logger.error(fromAxiosError(err))
      })
      return
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)
    const exists = await runnerAdapter.snapshotExists(snapshotRunner.snapshotRef)
    if (!exists) {
      await this.snapshotRunnerRepository.delete(snapshotRunner.id)
    } else {
      //  just in case the snapshot is still there
      runnerAdapter.removeSnapshot(snapshotRunner.snapshotRef).catch((err) => {
        //  this should not happen, and is not critical
        //  if the runner can not remote the snapshot, just delete the runner record
        this.snapshotRunnerRepository.delete(snapshotRunner.id).catch((err) => {
          this.logger.error(fromAxiosError(err))
        })
        //  and log the error for tracking
        const errorMessage = `Failed to do just in case remove snapshot ${snapshotRunner.snapshotRef} from runner ${runner.id}: ${fromAxiosError(err)}`
        this.logger.warn(errorMessage)
      })
    }
  }

  async handleSnapshotStateRemoving(snapshot: Snapshot): Promise<SyncState> {
    const snapshotRunnerItems = await this.snapshotRunnerRepository.find({
      where: {
        snapshotRef: snapshot.ref,
      },
    })

    if (snapshotRunnerItems.length === 0) {
      await this.snapshotRepository.remove(snapshot)
    }

    return DONT_SYNC_AGAIN
  }

  async handleCheckInitialRunnerSnapshot(snapshot: Snapshot): Promise<SyncState> {
    // Check for timeout - allow up to 30 minutes
    const timeoutMinutes = 30
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - snapshot.createdAt.getTime() > timeoutMs) {
      await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, 'Timeout processing snapshot on initial runner')
      return DONT_SYNC_AGAIN
    }

    // Check if the snapshot ref is already set and it is already on the runner
    const snapshotRunner = await this.snapshotRunnerRepository.findOne({
      where: {
        snapshotRef: snapshot.ref,
        runnerId: snapshot.initialRunnerId,
      },
    })

    if (snapshot.ref && snapshotRunner) {
      if (snapshotRunner.state === SnapshotRunnerState.READY) {
        await this.updateSnapshotState(snapshot.id, SnapshotState.ACTIVE)
      } else if (snapshotRunner.state === SnapshotRunnerState.ERROR) {
        await this.snapshotRunnerRepository.delete(snapshotRunner.id)
      }
      return DONT_SYNC_AGAIN
    }

    const runner = await this.runnerRepository.findOneOrFail({
      where: {
        id: snapshot.initialRunnerId,
      },
    })

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    const initialImageRefOnRunner = snapshot.buildInfo
      ? snapshot.buildInfo.snapshotRef
      : this.getInitialRunnerSnapshotTag(snapshot)

    const exists = await runnerAdapter.snapshotExists(initialImageRefOnRunner)
    if (!exists) {
      return DONT_SYNC_AGAIN
    }

    const snapshotInfoResponse = await runnerAdapter.getSnapshotInfo(initialImageRefOnRunner)

    const internalRegistry = await this.dockerRegistryService.getAvailableInternalRegistry(runner.region)
    if (!internalRegistry) {
      throw new Error('No internal registry found for snapshot')
    }

    // Process snapshot info in case it had failed or it's a build snapshot
    if (!snapshot.ref) {
      await this.processSnapshotInfo(snapshot, snapshotInfoResponse, internalRegistry)
    }

    try {
      await runnerAdapter.removeSnapshot(initialImageRefOnRunner)
    } catch (error) {
      this.logger.error(`Failed to remove snapshot ${snapshot.imageName}: ${fromAxiosError(error)}`)
    }

    await this.runnerService.createSnapshotRunnerEntry(runner.id, snapshot.ref, SnapshotRunnerState.READY)
    await this.updateSnapshotState(snapshot.id, SnapshotState.ACTIVE)

    // Best effort removal of old snapshot from transient registry
    const registry = await this.dockerRegistryService.findTransientRegistryBySnapshotImageName(snapshot.imageName)
    if (registry) {
      try {
        await this.dockerRegistryService.removeImage(snapshot.imageName, registry.id)
      } catch (error) {
        if (error.statusCode === 404) {
          //  image not found, just return
          return DONT_SYNC_AGAIN
        }
        this.logger.error('Failed to remove transient image:', fromAxiosError(error))
      }
    }

    return DONT_SYNC_AGAIN
  }

  async processPullOnInitialRunner(snapshot: Snapshot, runner: Runner) {
    // Check for timeout - allow up to 30 minutes
    const timeoutMinutes = 30
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - snapshot.updatedAt.getTime() > timeoutMs) {
      await this.updateSnapshotState(
        snapshot.id,
        SnapshotState.ERROR,
        'Timeout processing snapshot pull on initial runner',
      )
      return DONT_SYNC_AGAIN
    }

    let sourceRegistry = await this.dockerRegistryService.findSourceRegistryBySnapshotImageName(
      snapshot.imageName,
      runner.region,
    )
    if (!sourceRegistry) {
      sourceRegistry = await this.dockerRegistryService.getDefaultDockerHubRegistry()
    }
    const destinationRegistry = await this.dockerRegistryService.getAvailableInternalRegistry(runner.region)

    // Using image name for pull instead of the ref
    try {
      // If snapshot already has the designated ref from the manifest, pass it, otherwise let the runner build it
      await this.pullSnapshotRunnerWithRetries(
        runner,
        snapshot.imageName,
        sourceRegistry,
        destinationRegistry,
        snapshot.ref ? snapshot.ref : undefined,
      )

      // Tag image to org and creation timestamp for future use
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      const exists = await runnerAdapter.snapshotExists(snapshot.imageName)
      if (!exists) {
        return DONT_SYNC_AGAIN
      }

      await runnerAdapter.tagImage(snapshot.imageName, this.getInitialRunnerSnapshotTag(snapshot))

      // Best-effort cleanup of the original tag
      // Only if there is no other snapshot in a processing state that uses the same image
      try {
        const anotherSnapshot = await this.snapshotRepository.findOne({
          where: {
            name: snapshot.imageName,
            state: Not(In([SnapshotState.ACTIVE, SnapshotState.INACTIVE])),
          },
        })
        if (!anotherSnapshot) {
          await runnerAdapter.removeSnapshot(snapshot.imageName)
        }
      } catch (err) {
        this.logger.error(`Failed to cleanup original tag ${snapshot.imageName}: ${fromAxiosError(err)}`)
      }
    } catch (err) {
      if (err.code !== 'ECONNRESET') {
        await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, err.message)
        throw err
      }
      // TODO: check if retry
      return
    }
  }

  async processBuildOnRunner(snapshot: Snapshot, runner: Runner) {
    // todo: split dockerfile by FROM's and pass all docker registry creds to the building process

    try {
      const sourceRegistry = await this.dockerRegistryService.getDefaultDockerHubRegistry()
      const registry = await this.dockerRegistryService.getAvailableInternalRegistry(runner.region)

      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      registry.url = registry.url.replace(/^(https?:\/\/)/, '')
      await runnerAdapter.buildSnapshot(
        snapshot.buildInfo,
        snapshot.organizationId,
        sourceRegistry ? [sourceRegistry] : undefined,
        registry,
        true,
      )
    } catch (err) {
      if (err.code === 'ECONNRESET') {
        // Connection reset, retry later
        return
      }

      this.logger.error(`Error building snapshot ${snapshot.name}: ${fromAxiosError(err)}`)
      await this.updateSnapshotState(snapshot.id, SnapshotState.BUILD_FAILED, fromAxiosError(err).message)
    }
  }

  async handleSnapshotStatePending(snapshot: Snapshot): Promise<SyncState> {
    // TODO: get only runners where the base snapshot is available (extract from buildInfo)
    const excludedRunnerIds = snapshot.buildInfo
      ? await this.runnerService.getRunnersWithMultipleSnapshotsBuilding()
      : await this.runnerService.getRunnersWithMultipleSnapshotsPulling()

    let initialRunner: Runner | null = null
    try {
      const regions = await this.snapshotService.getSnapshotRegions(snapshot.id)
      if (!regions.length) {
        throw new Error('No regions found for snapshot')
      }

      initialRunner = await this.runnerService.getRandomAvailableRunner({
        regions: regions.map((region) => region.id),
        excludedRunnerIds: excludedRunnerIds,
      })
    } catch (error) {
      this.logger.warn(`Failed to get initial runner: ${fromAxiosError(error)}`)
    }

    if (!initialRunner) {
      // No runners available, retry later
      return DONT_SYNC_AGAIN
    }

    snapshot.initialRunnerId = initialRunner.id
    await this.snapshotRepository.save(snapshot)

    if (snapshot.buildInfo) {
      await this.updateSnapshotState(snapshot.id, SnapshotState.BUILDING)
      await this.processBuildOnRunner(snapshot, initialRunner)
    } else {
      await this.updateSnapshotState(snapshot.id, SnapshotState.PULLING)
      await this.processPullOnInitialRunner(snapshot, initialRunner)
    }

    return SYNC_AGAIN
  }

  private async updateSnapshotState(snapshotId: string, state: SnapshotState, errorReason?: string) {
    const partialUpdate: Partial<Snapshot> = {
      state,
    }

    if (errorReason !== undefined) {
      partialUpdate.errorReason = errorReason
    }

    const result = await this.snapshotRepository.update(
      {
        id: snapshotId,
      },
      partialUpdate,
    )

    if (!result.affected) {
      throw new NotFoundException(`Snapshot with ID ${snapshotId} not found`)
    }
  }

  @Cron(CronExpression.EVERY_HOUR, { name: 'cleanup-old-buildinfo-snapshot-runners' })
  @TrackJobExecution()
  @LogExecution('cleanup-old-buildinfo-snapshot-runners')
  @WithInstrumentation()
  async cleanupOldBuildInfoSnapshotRunners() {
    const lockKey = 'cleanup-old-buildinfo-snapshots-lock'
    if (!(await this.redisLockProvider.lock(lockKey, 300))) {
      return
    }

    try {
      const oneDayAgo = new Date()
      oneDayAgo.setDate(oneDayAgo.getDate() - 1)

      // Find all BuildInfo entities that haven't been used in over a day
      const oldBuildInfos = await this.buildInfoRepository.find({
        where: {
          lastUsedAt: LessThan(oneDayAgo),
        },
      })

      if (oldBuildInfos.length === 0) {
        return
      }

      const snapshotRefs = oldBuildInfos.map((buildInfo) => buildInfo.snapshotRef)

      const result = await this.snapshotRunnerRepository.update(
        { snapshotRef: In(snapshotRefs) },
        { state: SnapshotRunnerState.REMOVING },
      )

      if (result.affected > 0) {
        this.logger.debug(`Marked ${result.affected} SnapshotRunners for removal due to unused BuildInfo`)
      }
    } catch (error) {
      this.logger.error(`Failed to mark old BuildInfo SnapshotRunners for removal: ${fromAxiosError(error)}`)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_10_MINUTES, { name: 'deactivate-old-snapshots' })
  @TrackJobExecution()
  @LogExecution('deactivate-old-snapshots')
  @WithInstrumentation()
  async deactivateOldSnapshots() {
    const lockKey = 'deactivate-old-snapshots-lock'
    if (!(await this.redisLockProvider.lock(lockKey, 300))) {
      return
    }

    try {
      const twoWeeksAgo = new Date(Date.now() - 14 * 1000 * 60 * 60 * 24)

      const oldSnapshots = await this.snapshotRepository
        .createQueryBuilder('snapshot')
        .where('snapshot.general = false')
        .andWhere('snapshot.state = :snapshotState', { snapshotState: SnapshotState.ACTIVE })
        .andWhere('(snapshot."lastUsedAt" IS NULL OR snapshot."lastUsedAt" < :twoWeeksAgo)', { twoWeeksAgo })
        .andWhere('snapshot."createdAt" < :twoWeeksAgo', { twoWeeksAgo })
        .andWhere(
          () => {
            const query = this.snapshotRepository
              .createQueryBuilder('s')
              .select('1')
              .where('s."ref" = snapshot."ref"')
              .andWhere('s.state = :activeState')
              .andWhere('(s."lastUsedAt" >= :twoWeeksAgo OR s."createdAt" >= :twoWeeksAgo)')

            return `NOT EXISTS (${query.getQuery()})`
          },
          {
            activeState: SnapshotState.ACTIVE,
            twoWeeksAgo,
          },
        )
        .take(100)
        .getMany()

      if (oldSnapshots.length === 0) {
        return
      }

      // Deactivate the snapshots
      const snapshotIds = oldSnapshots.map((snapshot) => snapshot.id)
      await this.snapshotRepository.update({ id: In(snapshotIds) }, { state: SnapshotState.INACTIVE })

      // Get internal names of deactivated snapshots
      const refs = oldSnapshots.map((snapshot) => snapshot.ref).filter((name) => name) // Filter out null/undefined values

      if (refs.length > 0) {
        // Set associated SnapshotRunner records to REMOVING state
        const result = await this.snapshotRunnerRepository.update(
          { snapshotRef: In(refs) },
          { state: SnapshotRunnerState.REMOVING },
        )

        this.logger.debug(
          `Deactivated ${oldSnapshots.length} snapshots and marked ${result.affected} SnapshotRunners for removal`,
        )
      }
    } catch (error) {
      this.logger.error(`Failed to deactivate old snapshots: ${fromAxiosError(error)}`)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  @Cron(CronExpression.EVERY_10_MINUTES, { name: 'cleanup-inactive-snapshots-from-runners' })
  @TrackJobExecution()
  @LogExecution('cleanup-inactive-snapshots-from-runners')
  @WithInstrumentation()
  async cleanupInactiveSnapshotsFromRunners() {
    const lockKey = 'cleanup-inactive-snapshots-from-runners-lock'
    if (!(await this.redisLockProvider.lock(lockKey, 300))) {
      return
    }

    try {
      // Only fetch inactive snapshots that have associated snapshot runner entries
      const queryResult = await this.snapshotRepository
        .createQueryBuilder('snapshot')
        .select('snapshot."ref"')
        .where('snapshot.state = :snapshotState', { snapshotState: SnapshotState.INACTIVE })
        .andWhere('snapshot."ref" IS NOT NULL')
        .andWhereExists(
          this.snapshotRunnerRepository
            .createQueryBuilder('snapshot_runner')
            .select('1')
            .where('snapshot_runner."snapshotRef" = snapshot."ref"')
            .andWhere('snapshot_runner.state != :snapshotRunnerState', {
              snapshotRunnerState: SnapshotRunnerState.REMOVING,
            }),
        )
        .andWhere(
          () => {
            const query = this.snapshotRepository
              .createQueryBuilder('s')
              .select('1')
              .where('s."ref" = snapshot."ref"')
              .andWhere('s.state = :snapshotState')
            return `NOT EXISTS (${query.getQuery()})`
          },
          {
            snapshotState: SnapshotState.ACTIVE,
          },
        )
        .take(100)
        .getRawMany()

      const inactiveSnapshotRefs = queryResult.map((result) => result.ref)

      if (inactiveSnapshotRefs.length > 0) {
        // Set associated SnapshotRunner records to REMOVING state
        const result = await this.snapshotRunnerRepository.update(
          { snapshotRef: In(inactiveSnapshotRefs) },
          { state: SnapshotRunnerState.REMOVING },
        )

        this.logger.debug(`Marked ${result.affected} SnapshotRunners for removal`)
      }
    } catch (error) {
      this.logger.error(`Failed to cleanup inactive snapshots from runners: ${fromAxiosError(error)}`)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  private async processSnapshotInfo(
    snapshot: Snapshot,
    snapshotInfoResponse: RunnerSnapshotInfo,
    internalRegistry: DockerRegistry,
  ) {
    const sanitizedUrl = internalRegistry.url.replace(/^https?:\/\//, '')
    snapshot.ref = `${sanitizedUrl}/${internalRegistry.project || 'daytona'}/daytona-${snapshotInfoResponse.hash}:daytona`

    const organization = await this.organizationService.findOne(snapshot.organizationId)
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${snapshot.organizationId} not found`)
    }

    const MAX_SIZE_GB = organization.maxSnapshotSize

    if (snapshotInfoResponse.sizeGB > MAX_SIZE_GB) {
      await this.updateSnapshotState(
        snapshot.id,
        SnapshotState.ERROR,
        `Snapshot size (${snapshotInfoResponse.sizeGB.toFixed(2)}GB) exceeds maximum allowed size of ${MAX_SIZE_GB}GB`,
      )
      return DONT_SYNC_AGAIN
    }

    snapshot.size = snapshotInfoResponse.sizeGB

    // If entrypoint is not explicitly set, set it from snapshotInfoResponse
    if (!snapshot.entrypoint) {
      if (snapshotInfoResponse.entrypoint && snapshotInfoResponse.entrypoint.length > 0) {
        if (Array.isArray(snapshotInfoResponse.entrypoint)) {
          snapshot.entrypoint = snapshotInfoResponse.entrypoint
        } else {
          snapshot.entrypoint = [snapshotInfoResponse.entrypoint]
        }
      } else {
        snapshot.entrypoint = ['sleep', 'infinity']
      }
    }
  }

  private getInitialRunnerSnapshotTag(snapshot: Snapshot) {
    // Extract the base image name without any tag or digest
    let baseImageName = snapshot.imageName
    const colonIndex = baseImageName.indexOf(':')
    if (colonIndex !== -1) {
      baseImageName = baseImageName.substring(0, colonIndex)
    }
    const atIndex = baseImageName.indexOf('@')
    if (atIndex !== -1) {
      baseImageName = baseImageName.substring(0, atIndex)
    }
    return `${baseImageName}-${snapshot.id}-${snapshot.createdAt.getTime()}:daytona`
  }

  @OnEvent(SnapshotEvents.CREATED)
  private async handleSnapshotCreatedEvent(event: SnapshotCreatedEvent) {
    this.syncSnapshotState(event.snapshot.id).catch(this.logger.error)
  }
}
