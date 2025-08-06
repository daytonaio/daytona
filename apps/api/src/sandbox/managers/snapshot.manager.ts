/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, IsNull, LessThan, Not, Or, Raw, Repository } from 'typeorm'
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
import { RegistryType } from '../../docker-registry/enums/registry-type.enum'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { OrganizationService } from '../../organization/services/organization.service'
import { BuildInfo } from '../entities/build-info.entity'
import { fromAxiosError } from '../../common/utils/from-axios-error'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { RunnerService } from '../services/runner.service'
import { RunnerAdapterFactory, RunnerSnapshotInfo } from '../runner-adapter/runnerAdapter'
import { OnEvent } from '@nestjs/event-emitter'
import { SnapshotEvents } from '../constants/snapshot-events'
import { SnapshotCreatedEvent } from '../events/snapshot-created.event'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'

const DEFAULT_INITIAL_RUNNER_REGION = 'us'

const SYNC_AGAIN = 'sync-again'
const DONT_SYNC_AGAIN = 'dont-sync-again'
type SyncState = typeof SYNC_AGAIN | typeof DONT_SYNC_AGAIN

@Injectable()
export class SnapshotManager {
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
  ) {}

  @Cron(CronExpression.EVERY_5_SECONDS)
  async syncRunnerSnapshots() {
    const lockKey = 'sync-runner-snapshots-lock'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
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
      await this.redis.set('sync-runner-snapshots-skip', 0)
      return
    }

    await this.redis.set('sync-runner-snapshots-skip', Number(skip) + snapshots.length)

    await Promise.all(
      snapshots.map((snapshot) => {
        this.propagateSnapshotToRunners(snapshot.ref).catch((err) => {
          this.logger.error(`Error propagating snapshot ${snapshot.id} to runners: ${err}`)
        })
      }),
    )

    await this.redisLockProvider.unlock(lockKey)
  }

  @Cron(CronExpression.EVERY_10_SECONDS)
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

  async propagateSnapshotToRunners(ref: string) {
    //  todo: remove try catch block and implement error handling
    try {
      const runners = await this.runnerRepository.find({
        where: {
          state: RunnerState.READY,
          unschedulable: false,
        },
      })

      //  get all runners that have the snapshot in their base image
      const snapshotRunners = await this.snapshotRunnerRepository.find({
        where: {
          snapshotRef: ref,
          state: In([SnapshotRunnerState.READY, SnapshotRunnerState.PULLING_SNAPSHOT]),
        },
      })
      //  filter duplicate snapshot runner records
      const snapshotRunnersDistinctRunnersIds = [
        ...new Set(snapshotRunners.map((snapshotRunner) => snapshotRunner.runnerId)),
      ]

      const propagateLimit = Math.ceil(runners.length / 3) - snapshotRunnersDistinctRunnersIds.length
      const unallocatedRunners = runners.filter(
        (runner) => !snapshotRunnersDistinctRunnersIds.some((snapshotRunnerId) => snapshotRunnerId === runner.id),
      )
      //  shuffle the runners to propagate to
      unallocatedRunners.sort(() => Math.random() - 0.5)
      //  limit the number of runners to propagate to
      const runnersToPropagateTo = unallocatedRunners.slice(0, propagateLimit)

      let dockerRegistry = await this.dockerRegistryService.findOneBySnapshotImageName(ref)

      // If no registry found by image name, use the default internal registry
      if (!dockerRegistry) {
        dockerRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
        if (!dockerRegistry) {
          throw new Error('No registry found for snapshot and no default internal registry configured')
        }
      }

      const results = await Promise.allSettled(
        runnersToPropagateTo.map(async (runner) => {
          const snapshotRunner = await this.runnerService.getSnapshotRunner(runner.id, ref)

          try {
            if (!snapshotRunner) {
              await this.runnerService.createSnapshotRunnerEntry(runner.id, ref, SnapshotRunnerState.PULLING_SNAPSHOT)
              await this.pullSnapshotRunnerWithRetries(runner, ref, dockerRegistry)
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
    sourceRegistry?: DockerRegistry,
    destinationRegistry?: DockerRegistry,
  ) {
    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    let retries = 0
    while (retries < 10) {
      try {
        await runnerAdapter.pullSnapshot(snapshotRef, sourceRegistry, destinationRegistry)
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
    if (Date.now() - snapshotRunner.createdAt.getTime() > timeoutMs) {
      snapshotRunner.state = SnapshotRunnerState.ERROR
      snapshotRunner.errorReason = 'Timeout while pulling snapshot to runner'
      await this.snapshotRunnerRepository.save(snapshotRunner)
      return
    }

    const retryTimeoutMinutes = 10
    const retryTimeoutMs = retryTimeoutMinutes * 60 * 1000
    if (Date.now() - snapshotRunner.createdAt.getTime() > retryTimeoutMs) {
      const dockerRegistry = await this.dockerRegistryService.findOneBySnapshotImageName(snapshotRunner.snapshotRef)
      await this.pullSnapshotRunnerWithRetries(runner, snapshotRunner.snapshotRef, dockerRegistry)
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

  @Cron(CronExpression.EVERY_10_SECONDS)
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
        await this.snapshotRunnerRepository.update(
          {
            snapshotRef: snapshot.ref,
          },
          {
            state: SnapshotRunnerState.REMOVING,
          },
        )

        await this.snapshotRepository.remove(snapshot)
      }),
    )

    await this.redisLockProvider.unlock(lockKey)
  }

  @Cron(CronExpression.EVERY_10_SECONDS)
  async checkSnapshotState() {
    //  the first time the snapshot is created it needs to be validated and pushed to the internal registry
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
    const lockKey = `check-snapshot-state-lock-${snapshotId}`
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
        case SnapshotState.VALIDATING:
          syncState = await this.handleSnapshotStateValidating(snapshot)
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

  async handleSnapshotStateValidating(snapshot: Snapshot): Promise<SyncState> {
    const timeoutMinutes = 1
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - snapshot.updatedAt.getTime() > timeoutMs) {
      await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, 'Timeout while validating snapshot')
      return DONT_SYNC_AGAIN
    }

    // TODO: add sleep?
    return SYNC_AGAIN
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
        await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, snapshotRunner.errorReason)
      }
      return DONT_SYNC_AGAIN
    }

    const runner = await this.runnerRepository.findOneOrFail({
      where: {
        id: snapshot.initialRunnerId,
      },
    })

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    const initialImageRefOnRunner = snapshot.buildInfo ? snapshot.buildInfo.snapshotRef : snapshot.imageName

    const snapshotInfoResponse = await runnerAdapter.getSnapshotInfo(initialImageRefOnRunner)

    // Process snapshot info in case it had failed or it's a build snapshot
    if (!snapshot.ref) {
      await this.processSnapshotInfo(snapshot, snapshotInfoResponse)
    }

    try {
      await runnerAdapter.removeSnapshot(snapshot.imageName)
    } catch (error) {
      this.logger.error(`Failed to remove snapshot ${snapshot.imageName}: ${fromAxiosError(error)}`)
    }

    if (!snapshot.skipValidation) {
      snapshot.state = SnapshotState.VALIDATING
      await this.snapshotRepository.save(snapshot)

      await this.validateSandboxCreationOnRunner(snapshot, runner)
    }

    snapshot.state = SnapshotState.ACTIVE
    await this.snapshotRepository.save(snapshot)

    await this.runnerService.createSnapshotRunnerEntry(runner.id, snapshot.ref, SnapshotRunnerState.READY)

    // Best effort removal of old snapshot from transient registry
    const registry = await this.dockerRegistryService.findOneBySnapshotImageName(
      snapshot.imageName,
      snapshot.organizationId,
    )
    if (registry && registry.registryType === RegistryType.TRANSIENT) {
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

  async validateSandboxCreationOnRunner(snapshot: Snapshot, runner: Runner) {
    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    const sandbox = new Sandbox()
    sandbox.id = uuidv4()
    sandbox.snapshot = snapshot.ref
    sandbox.osUser = 'root'
    sandbox.disk = snapshot.disk
    sandbox.mem = snapshot.mem
    sandbox.cpu = snapshot.cpu
    sandbox.organizationId = snapshot.organizationId
    sandbox.runnerId = runner.id
    sandbox.labels = {
      DAYTONA_VALIDATING_SNAPSHOT_ID: snapshot.id,
    }

    const registry = await this.dockerRegistryService.getDefaultInternalRegistry()
    await runnerAdapter.createSandbox(sandbox, registry, snapshot.entrypoint)

    // Wait for 5 seconds to ensure the sandbox hasn't exited
    await new Promise((resolve) => setTimeout(resolve, 5000))

    const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
    if (sandboxInfo.state === SandboxState.STARTED) {
      await this.updateSnapshotState(snapshot.id, SnapshotState.ACTIVE)
    } else {
      await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, 'Snapshot validation failed')
    }

    try {
      await runnerAdapter.destroySandbox(sandbox.id)
    } catch (error) {
      this.logger.error(`Failed to destroy sandbox ${sandbox.id}: ${fromAxiosError(error)}`)
    }
  }

  async processPullOnInitialRunner(snapshot: Snapshot, runner: Runner) {
    const sourceRegistry = await this.dockerRegistryService.findOneBySnapshotImageName(
      snapshot.imageName,
      snapshot.organizationId,
    )
    const destinationRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()

    // Using image name for pull instead of the ref
    try {
      await this.pullSnapshotRunnerWithRetries(runner, snapshot.imageName, sourceRegistry, destinationRegistry)
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
      const registry = await this.dockerRegistryService.getDefaultInternalRegistry()

      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      await runnerAdapter.buildSnapshot(snapshot.buildInfo, snapshot.organizationId, registry, true)

      // // save snapshotRunner

      // const snapshotRef = `${registry.url}/${registry.project}/${snapshot.buildInfo.snapshotRef}`

      // snapshot.ref = snapshotRef
      // await this.snapshotRepository.save(snapshot)

      // // Wait for 30 seconds because of Harbor's delay at making newly created snapshots available
      // await new Promise((resolve) => setTimeout(resolve, 30000))
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

    const initialRunner = await this.runnerService.getRandomAvailableRunner({
      region: DEFAULT_INITIAL_RUNNER_REGION,
      excludedRunnerIds: excludedRunnerIds,
    })

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
    const snapshot = await this.snapshotRepository.findOneOrFail({
      where: {
        id: snapshotId,
      },
    })
    snapshot.state = state
    if (errorReason) {
      snapshot.errorReason = errorReason
    }
    await this.snapshotRepository.save(snapshot)
  }

  @Cron(CronExpression.EVERY_HOUR)
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

  @Cron(CronExpression.EVERY_10_MINUTES)
  async deactivateOldSnapshots() {
    const lockKey = 'deactivate-old-snapshots-lock'
    if (!(await this.redisLockProvider.lock(lockKey, 300))) {
      return
    }

    try {
      const twoWeeksAgo = new Date(Date.now() - 14 * 1000 * 60 * 60 * 24)

      // Find all active snapshots that haven't been used in over 14 days or have null lastUsedAt
      const oldSnapshots = await this.snapshotRepository.find({
        where: [
          {
            general: false,
            state: SnapshotState.ACTIVE,
            lastUsedAt: Or(IsNull(), LessThan(twoWeeksAgo)),
            createdAt: LessThan(twoWeeksAgo),
          },
        ],
        take: 100,
      })

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

  @Cron(CronExpression.EVERY_10_MINUTES)
  async cleanupInactiveSnapshotsFromRunners() {
    const lockKey = 'cleanup-inactive-snapshots-from-runners-lock'
    if (!(await this.redisLockProvider.lock(lockKey, 300))) {
      return
    }

    try {
      // Only fetch inactive snapshots that have associated snapshot runner entries
      const queryResult = await this.snapshotRepository
        .createQueryBuilder('snapshot')
        .select('snapshot."internalName"')
        .where('snapshot.state = :snapshotState', { snapshotState: SnapshotState.INACTIVE })
        .andWhere('snapshot."internalName" IS NOT NULL')
        .andWhereExists(
          this.snapshotRunnerRepository
            .createQueryBuilder('snapshot_runner')
            .select('1')
            .where('snapshot_runner."snapshotRef" = snapshot."internalName"')
            .andWhere('snapshot_runner.state != :snapshotRunnerState', {
              snapshotRunnerState: SnapshotRunnerState.REMOVING,
            }),
        )
        .take(100)
        .getRawMany()

      const inactiveSnapshotInternalNames = queryResult.map((result) => result.internalName)

      if (inactiveSnapshotInternalNames.length > 0) {
        // Set associated SnapshotRunner records to REMOVING state
        const result = await this.snapshotRunnerRepository.update(
          { snapshotRef: In(inactiveSnapshotInternalNames) },
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

  private async processSnapshotInfo(snapshot: Snapshot, snapshotInfoResponse: RunnerSnapshotInfo) {
    const defaultInternalRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
    snapshot.ref = `${defaultInternalRegistry.url}/${defaultInternalRegistry.project}/daytona-${snapshotInfoResponse.hash}:daytona`

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

    // Ensure entrypoint is set
    if (!snapshot.entrypoint) {
      if (snapshotInfoResponse.entrypoint) {
        if (Array.isArray(snapshotInfoResponse.entrypoint)) {
          snapshot.entrypoint = snapshotInfoResponse.entrypoint
        } else {
          snapshot.entrypoint = [snapshotInfoResponse.entrypoint]
        }
      } else if (snapshotInfoResponse.cmd) {
        if (Array.isArray(snapshotInfoResponse.cmd)) {
          snapshot.entrypoint = snapshotInfoResponse.cmd
        } else {
          snapshot.entrypoint = [snapshotInfoResponse.cmd]
        }
      } else {
        snapshot.entrypoint = ['sleep', 'infinity']
      }
    }
  }

  @OnEvent(SnapshotEvents.CREATED)
  private async handleSnapshotCreatedEvent(event: SnapshotCreatedEvent) {
    this.syncSnapshotState(event.snapshot.id).catch(this.logger.error)
  }
}
