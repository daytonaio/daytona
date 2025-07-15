/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, IsNull, LessThan, Not, Or, Repository } from 'typeorm'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { Runner } from '../entities/runner.entity'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { RunnerState } from '../enums/runner-state.enum'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { RunnerApiFactory } from '../runner-api/runnerApi'
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
import { RunnerRegion } from '../enums/runner-region.enum'

const SYNC_WARM_RUNNER_SNAPSHOTS_LOCK_KEY = 'sync-warm-runner-snapshots-lock'
const SYNC_WARM_RUNNER_SNAPSHOTS_SKIP_KEY = 'sync-warm-runner-snapshots-skip'
const SYNC_ACTIVE_RUNNER_SNAPSHOTS_LOCK_KEY = 'sync-active-runner-snapshots-lock'
const RUNNER_USAGE_THRESHOLD = 0.75

const MINIMUM_RUNNER_PROPAGATION_COUNT = 3

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
    private readonly runnerApiFactory: RunnerApiFactory,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly organizationService: OrganizationService,
  ) {}

  // @Cron(CronExpression.EVERY_5_SECONDS)
  async syncRunnerSnapshots() {
    if (!(await this.redisLockProvider.lock(SYNC_WARM_RUNNER_SNAPSHOTS_LOCK_KEY, 30))) {
      return
    }

    const skip = (await this.redis.get(SYNC_WARM_RUNNER_SNAPSHOTS_SKIP_KEY)) || 0

    // Use the optimized query to find snapshots that need propagation
    const snapshotsNeedingPropagation = await this.snapshotRepository
      .createQueryBuilder('s')
      .select('s.ref', 'snapshotRef')
      .addSelect('(s.desiredPropagation - COUNT(sr.id))', 'remainingPropagation')
      .innerJoin('organization', 'o', 'o.id = s.organizationId')
      .leftJoin('snapshot_runner', 'sr', 'sr.snapshotRef = s.ref AND sr.state = :readyState', {
        readyState: SnapshotRunnerState.READY,
      })
      .where('s.state IN (:...snapshotStates)', {
        snapshotStates: [SnapshotState.WARMING_UP, SnapshotState.ACTIVE],
      })
      .andWhere('o.suspended = false')
      .groupBy('s.id')
      .having('COUNT(sr.id) < s.desiredPropagation')
      .orderBy('s.createdAt', 'ASC')
      .take(100)
      .skip(Number(skip))
      .getRawMany()

    if (snapshotsNeedingPropagation.length === 0) {
      await this.redis.set(SYNC_WARM_RUNNER_SNAPSHOTS_SKIP_KEY, 0)
      return
    }

    await this.redis.set(SYNC_WARM_RUNNER_SNAPSHOTS_SKIP_KEY, Number(skip) + snapshotsNeedingPropagation.length)

    await Promise.all(
      snapshotsNeedingPropagation.map((item) => {
        this.propagateSnapshotToRunners(item.snapshotRef).catch((err) => {
          this.logger.error(`Error propagating snapshot with ref ${item.snapshotRef} to runners: ${err}`)
        })
      }),
    )

    await this.redisLockProvider.unlock(SYNC_WARM_RUNNER_SNAPSHOTS_LOCK_KEY)
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

  async propagateSnapshotToRunners(snapshotRef: string) {
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
          snapshotRef: snapshotRef,
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

      let dockerRegistry = await this.dockerRegistryService.findOneBySnapshotImageName(snapshotRef)

      // If no registry found by image name, use the default internal registry
      if (!dockerRegistry) {
        dockerRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
        if (!dockerRegistry) {
          throw new Error('No registry found for snapshot and no default internal registry configured')
        }
      }

      const results = await Promise.allSettled(
        runnersToPropagateTo.map(async (runner) => {
          const snapshotRunner = await this.runnerService.getSnapshotRunner(runner.id, snapshotRef)

          try {
            if (snapshotRunner && !snapshotRunner.snapshotRef) {
              //  this should never happen
              this.logger.warn(`Internal snapshot name not found for snapshot runner ${snapshotRunner.id}`)
              return
            }

            if (!snapshotRunner) {
              await this.runnerService.createSnapshotRunner(
                runner.id,
                snapshotRef,
                SnapshotRunnerState.PULLING_SNAPSHOT,
              )
              await this.pullSnapshotRunnerWithRetries(runner, snapshotRef, dockerRegistry)
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
    const snapshotApi = this.runnerApiFactory.createSnapshotApi(runner)

    let retries = 0
    while (retries < 10) {
      try {
        await snapshotApi.pullSnapshot({
          snapshot: snapshotRef,
          sourceRegistry: sourceRegistry
            ? {
                url: sourceRegistry.url,
                username: sourceRegistry.username,
                password: sourceRegistry.password,
                project: sourceRegistry.project,
              }
            : undefined,
          destinationRegistry: destinationRegistry
            ? {
                url: destinationRegistry.url,
                username: destinationRegistry.username,
                password: destinationRegistry.password,
                project: destinationRegistry.project,
              }
            : undefined,
        })
        break
      } catch (err) {
        if (err.code !== 'ECONNRESET') {
          throw err
        }
      }
      retries++
      await new Promise((resolve) => setTimeout(resolve, retries * 1000))
    }
  }

  async handleSnapshotRunnerStatePullingSnapshot(snapshotRunner: SnapshotRunner, runner: Runner) {
    const snapshotApi = this.runnerApiFactory.createSnapshotApi(runner)
    const response = (await snapshotApi.snapshotExists(snapshotRunner.snapshotRef)).data
    if (response.exists) {
      snapshotRunner.state = SnapshotRunnerState.READY
      await this.snapshotRunnerRepository.save(snapshotRunner)
      return
    }

    const timeoutMinutes = 60
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - snapshotRunner.createdAt.getTime() > timeoutMs) {
      snapshotRunner.state = SnapshotRunnerState.ERROR
      snapshotRunner.errorReason = 'Timeout while pulling snapshot'
      await this.snapshotRunnerRepository.save(snapshotRunner)
      return
    }

    const retryTimeoutMinutes = 10
    const retryTimeoutMs = retryTimeoutMinutes * 60 * 1000
    if (Date.now() - snapshotRunner.createdAt.getTime() > retryTimeoutMs) {
      await this.retrySnapshotRunnerPull(snapshotRunner)
      return
    }
  }

  async handleSnapshotRunnerStateBuildingSnapshot(snapshotRunner: SnapshotRunner, runner: Runner) {
    const runnerSandboxApi = this.runnerApiFactory.createSnapshotApi(runner)
    const response = (await runnerSandboxApi.snapshotExists(snapshotRunner.snapshotRef)).data
    if (response && response.exists) {
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
        // Check if there are other snapshots with the same ref not in REMOVING state
        const otherSnapshotsWithSameRef = await this.snapshotRepository.count({
          where: {
            ref: snapshot.ref,
            state: Not(SnapshotState.REMOVING),
          },
        })

        // Only update snapshot runners to REMOVING if no other snapshots use this ref
        if (otherSnapshotsWithSameRef === 0) {
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

  @Cron(CronExpression.EVERY_10_SECONDS)
  async checkSnapshotState() {
    const snapshots = await this.snapshotRepository.find({
      where: {
        state: Not(In([SnapshotState.ACTIVE, SnapshotState.ERROR, SnapshotState.BUILD_FAILED, SnapshotState.INACTIVE])),
      },
    })

    await Promise.all(
      snapshots.map(async (snapshot) => {
        const lockKey = `check-snapshot-state-lock-${snapshot.id}`
        if (!(await this.redisLockProvider.lock(lockKey, 720))) {
          return
        }

        try {
          switch (snapshot.state) {
            case SnapshotState.PENDING:
              await this.handleSnapshotStatePending(snapshot)
              break
            case SnapshotState.PULLING || SnapshotState.BUILDING:
              await this.handleCheckInitialRunnerSnapshot(snapshot)
              break
            case SnapshotState.REMOVING:
              await this.handleSnapshotStateRemoving(snapshot)
              break
          }
        } catch (error) {
          if (error.code === 'ECONNRESET') {
            await this.redisLockProvider.unlock(lockKey)
            this.checkSnapshotState()
            return
          }

          const message = error.message || String(error)
          await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, message)
        }

        await this.redisLockProvider.unlock(lockKey)
      }),
    )
  }

  async handleSnapshotRunnerStateRemoving(snapshotRunner: SnapshotRunner, runner: Runner) {
    // TODO: check - if snapshot runner was updated to this state less than one minute ago, skip it in case a Sandbox creation is taking place
    if (Date.now() - snapshotRunner.updatedAt.getTime() < 60000) {
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

    const snapshotApi = this.runnerApiFactory.createSnapshotApi(runner)
    const response = await snapshotApi.snapshotExists(snapshotRunner.snapshotRef)
    if (response.data && !response.data.exists) {
      await this.snapshotRunnerRepository.delete(snapshotRunner.id)
    } else {
      //  just in case the snapshot is still there
      snapshotApi.removeSnapshot(snapshotRunner.snapshotRef).catch((err) => {
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

  async handleSnapshotStateRemoving(snapshot: Snapshot) {
    const snapshotRunnerItems = await this.snapshotRunnerRepository.find({
      where: {
        snapshotRef: snapshot.ref,
      },
    })

    if (snapshotRunnerItems.length === 0) {
      await this.snapshotRepository.remove(snapshot)
    }
  }

  async handleCheckInitialRunnerSnapshot(snapshot: Snapshot) {
    // Check for timeout - allow up to 30 minutes
    const timeoutMinutes = 30
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - snapshot.createdAt.getTime() > timeoutMs) {
      await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, 'Timeout processing snapshot on initial runner')
      return
    }

    // This is the only case where we search for id instead of the ref
    const snapshotRunner = await this.snapshotRunnerRepository.findOne({
      where: {
        snapshotRef: snapshot.id,
        runnerId: snapshot.initialRunnerId,
      },
    })

    // If no snapshot runner found or it's not in READY/ERROR state, just return
    if (!snapshotRunner) {
      return
    }

    if (snapshotRunner.state === SnapshotRunnerState.ERROR) {
      await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, snapshotRunner.errorReason)
      return
    }

    if (snapshotRunner.state === SnapshotRunnerState.READY) {
      const runner = await this.runnerRepository.findOneOrFail({
        where: {
          id: snapshot.initialRunnerId,
        },
      })

      const snapshotApi = this.runnerApiFactory.createSnapshotApi(runner)

      const snapshotInfoResponse = (await snapshotApi.getSnapshotInfo(snapshot.imageName)).data

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
        return
      }

      snapshot.size = snapshotInfoResponse.sizeGB
      snapshot.state = SnapshotState.ACTIVE

      // Ensure entrypoint is set
      if (!snapshot.entrypoint) {
        if (snapshotInfoResponse.entrypoint) {
          if (Array.isArray(snapshotInfoResponse.entrypoint)) {
            snapshot.entrypoint = snapshotInfoResponse.entrypoint
          } else {
            snapshot.entrypoint = [snapshotInfoResponse.entrypoint]
          }
        } else {
          snapshot.entrypoint = ['sleep', 'infinity']
        }
      }

      // Update snapshot ref
      snapshotRunner.snapshotRef = snapshot.ref

      // // Update all the snapshotIds of the targetPropagations from the snapshot to now be the ref
      // await this.snapshotTargetPropagationRepository.update(
      //   { snapshotId: snapshot.id },
      //   { snapshotId: snapshot.ref },
      // )

      await this.snapshotRunnerRepository.save(snapshotRunner)

      await this.snapshotRepository.save(snapshot)
    }

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
          //  snapshot not found, just return
          return
        }
        this.logger.error('Failed to remove old snapshot:', fromAxiosError(error))
      }
    }
  }

  async processPullOnInitialRunner(snapshot: Snapshot) {
    const runner = await this.runnerRepository.findOneOrFail({
      where: {
        id: snapshot.initialRunnerId,
      },
    })

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

  async processBuildOnRunner(snapshot: Snapshot) {
    // todo: split dockerfile by FROM's and pass all docker registry creds to the building process

    try {
      const registry = await this.dockerRegistryService.getDefaultInternalRegistry()
      const runner = await this.runnerService.findOne(snapshot.initialRunnerId)

      const runnerSnapshotApi = this.runnerApiFactory.createSnapshotApi(runner)

      await runnerSnapshotApi.buildSnapshot({
        snapshot: snapshot.buildInfo.snapshotRef, // Name doesn't matter for runner, it uses the snapshot ID when pushing to internal registry
        registry: {
          url: registry.url,
          project: registry.project,
          username: registry.username,
          password: registry.password,
        },
        organizationId: snapshot.organizationId,
        dockerfile: snapshot.buildInfo.dockerfileContent,
        context: snapshot.buildInfo.contextHashes,
        pushToInternalRegistry: true,
      })

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

  async handleSnapshotStatePending(snapshot: Snapshot) {
    let excludedRunnerIds = []

    if (snapshot.buildInfo) {
      // TODO: get only runners where the base snapshot is available (extract from buildInfo)
      excludedRunnerIds = await this.runnerService.getRunnersWithMultipleSnapshotsBuilding()
    } else {
      excludedRunnerIds = await this.runnerService.getRunnersWithMultipleSnapshotsPulling()
    }

    const initialRunner = await this.runnerService.getRandomAvailableRunner({
      region: RunnerRegion.US,
      excludedRunnerIds: excludedRunnerIds,
    })

    if (!initialRunner) {
      // No runners available, retry later
      return
    }

    snapshot.initialRunnerId = initialRunner.id

    if (snapshot.buildInfo) {
      await this.updateSnapshotState(snapshot.id, SnapshotState.BUILDING)
      await this.processBuildOnRunner(snapshot)
    } else {
      await this.updateSnapshotState(snapshot.id, SnapshotState.PULLING)
      await this.processPullOnInitialRunner(snapshot)
    }

    // Check if should just create without the update check
    const snapshotRunner = await this.snapshotRunnerRepository.findOne({
      where: {
        runnerId: initialRunner.id,
        snapshotRef: snapshot.id,
      },
    })

    if (!snapshotRunner) {
      await this.runnerService.createSnapshotRunner(initialRunner.id, snapshot.id, SnapshotRunnerState.READY)
      return
    }

    if (snapshotRunner.state !== SnapshotRunnerState.READY) {
      await this.snapshotRunnerRepository.update(snapshotRunner.id, {
        state: SnapshotRunnerState.READY,
      })
    }
  }

  async retrySnapshotRunnerPull(snapshotRunner: SnapshotRunner) {
    const runner = await this.runnerRepository.findOneOrFail({
      where: {
        id: snapshotRunner.runnerId,
      },
    })

    const snapshotApi = this.runnerApiFactory.createSnapshotApi(runner)

    const dockerRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
    //  await this.redis.setex(lockKey, 360, this.instanceId)

    await snapshotApi.pullSnapshot({
      snapshot: snapshotRunner.snapshotRef,
      sourceRegistry: {
        url: dockerRegistry.url,
        username: dockerRegistry.username,
        password: dockerRegistry.password,
      },
    })
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

  @Cron(CronExpression.EVERY_30_SECONDS)
  async scaleOverusedSnapshots() {
    const lockKey = 'scale-overused-snapshots-lock'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    try {
      // First check the average usage of all schedulable runners
      const avgUsageResult = await this.runnerRepository.query(`
        SELECT AVG(used * 1.0 / capacity) as avg_usage
        FROM runner
        WHERE state = 'ready' AND capacity > 0 AND unschedulable = false
      `)

      const avgUsage = parseFloat(avgUsageResult[0]?.avg_usage || '0')

      // Only proceed if average usage is less than 75%
      if (avgUsage >= RUNNER_USAGE_THRESHOLD) {
        this.logger.debug(
          `Skipping scaling overused snapshots as overall runner usage is high (≥${RUNNER_USAGE_THRESHOLD * 100}%)`,
        )
        return
      }

      const MINIMUM_RUNNER_PROPAGATION_COUNT = 3
      const AVAILABLE_VCPU_PER_RUNNER = 32
      const AVAILABLE_MEMORY_GB_PER_RUNNER = 128

      const syncRunnersQuery = `
      WITH snapshot_data AS (
          SELECT
              stp.snapshotRef,
              s.id AS snapshot_id,
              GREATEST(
                  COALESCE(stp.userOverride, stp.desiredConcurrentSandboxes),
                  ${MINIMUM_RUNNER_PROPAGATION_COUNT}
              ) AS desired_count,
              s.cpu,
              s.mem
          FROM SnapshotTargetPropagation stp
          JOIN Snapshot s ON s.ref = stp.snapshotRef
      ),
      current_propagations AS (
          SELECT
              sr.snapshotRef,
              COUNT(*) AS active_count
          FROM SnapshotRunner sr
          WHERE sr.state IN ('active', 'warming_up')
            AND sr.snapshotRef IS NOT NULL
          GROUP BY sr.snapshotRef
      ),
      needed_propagation AS (
          SELECT
              sd.snapshot_id,
              sd.snapshotRef,
              sd.cpu,
              sd.mem,
              sd.desired_count,
              COALESCE(cp.active_count, 0) AS current_count,
              GREATEST(0, sd.desired_count - COALESCE(cp.active_count, 0)) AS needed_count
          FROM snapshot_data sd
          LEFT JOIN current_propagations cp ON cp.snapshotRef = sd.snapshotRef
      ),
      runners_with_capacity AS (
          SELECT
              sr.runnerId,
              ${AVAILABLE_VCPU_PER_RUNNER} - COALESCE(SUM(s.cpu), 0) AS remaining_cpu,
              ${AVAILABLE_MEMORY_GB_PER_RUNNER} - COALESCE(SUM(s.mem), 0) AS remaining_mem
          FROM SnapshotRunner sr
          LEFT JOIN Snapshot s ON s.ref = sr.snapshotRef
          WHERE sr.state IN ('active', 'warming_up')
          GROUP BY sr.runnerId
      ),
      eligible_runners AS (
          SELECT
              rwc.runnerId
          FROM runners_with_capacity rwc
          JOIN needed_propagation np ON
              rwc.remaining_cpu >= np.cpu AND
              rwc.remaining_mem >= np.mem
      )
      SELECT
          er.runnerId,
          np.snapshot_id,
          FALSE AS "shouldPropagate"
      FROM needed_propagation np
      JOIN eligible_runners er ON true
      WHERE np.needed_count > 0
      LIMIT (
          SELECT COUNT(*) FROM needed_propagation WHERE needed_count > 0
      )

      UNION ALL

      SELECT
          NULL AS runnerId,
          np.snapshot_id,
          TRUE AS "shouldPropagate"
      FROM needed_propagation np
      WHERE np.needed_count = 0;
      `

      // TODO: update the SnapshotTargetPropagationState to PROPAGATING if doing more, or READY if satisfied or scaling down

      // Find overused snapshots and increase their desired propagation
      const result = await this.snapshotRepository.query(syncRunnersQuery)

      // Iterate through the results and call different functions based on shouldPropagate
      for (const item of result) {
        if (item.shouldPropagate === true) {
          // propagate to each runner
          await this.propagateSnapshotToRunners(item.snapshot_id)
        } else {
          // For items where shouldPropagate is false and runnerId is not null
          //     await this.pullSnapshotRunnerWithRetries(runner, snapshotRef, dockerRegistry)
        }
      }
    } catch (error) {
      this.logger.error(`Error syncing snapshot runners: ${fromAxiosError(error)}`)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }
}
