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
import { DockerProvider } from '../docker/docker-provider'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { Runner } from '../entities/runner.entity'
import { RunnerState } from '../enums/runner-state.enum'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { v4 as uuidv4 } from 'uuid'
import { RunnerNotReadyError } from '../errors/runner-not-ready.error'
import { RegistryType } from '../../docker-registry/enums/registry-type.enum'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { OrganizationService } from '../../organization/services/organization.service'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { BuildInfo } from '../entities/build-info.entity'
import { fromAxiosError } from '../../common/utils/from-axios-error'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { RunnerService } from '../services/runner.service'
import { RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'
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
    private readonly dockerProvider: DockerProvider,
    private readonly runnerAdapterFactory: RunnerAdapterFactory,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly organizationService: OrganizationService,
  ) { }

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
        this.propagateSnapshotToRunners(snapshot.internalName).catch((err) => {
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
        await this.handleSnapshotRunnerStatePullingSnapshot(snapshotRunner)
        break
      case SnapshotRunnerState.BUILDING_SNAPSHOT:
        await this.handleSnapshotRunnerStateBuildingSnapshot(snapshotRunner)
        break
      case SnapshotRunnerState.REMOVING:
        await this.handleSnapshotRunnerStateRemoving(snapshotRunner)
        break
    }
  }

  async propagateSnapshotToRunners(internalSnapshotName: string) {
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
          snapshotRef: internalSnapshotName,
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

      const results = await Promise.allSettled(
        runnersToPropagateTo.map(async (runner) => {
          let snapshotRunner = await this.snapshotRunnerRepository.findOne({
            where: {
              snapshotRef: internalSnapshotName,
              runnerId: runner.id,
            },
          })

          try {
            if (snapshotRunner && !snapshotRunner.snapshotRef) {
              //  this should never happen
              this.logger.warn(`Internal snapshot name not found for snapshot runner ${snapshotRunner.id}`)
              return
            }

            if (!snapshotRunner) {
              snapshotRunner = new SnapshotRunner()
              snapshotRunner.snapshotRef = internalSnapshotName
              snapshotRunner.runnerId = runner.id
              snapshotRunner.state = SnapshotRunnerState.PULLING_SNAPSHOT
              await this.snapshotRunnerRepository.save(snapshotRunner)
              await this.propagateSnapshotToRunner(internalSnapshotName, runner)
            } else if (snapshotRunner.state === SnapshotRunnerState.PULLING_SNAPSHOT) {
              await this.handleSnapshotRunnerStatePullingSnapshot(snapshotRunner)
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

  async propagateSnapshotToRunner(internalSnapshotName: string, runner: Runner) {
    let dockerRegistry = await this.dockerRegistryService.findOneBySnapshotImageName(internalSnapshotName)

    // If no registry found by image name, use the default internal registry
    if (!dockerRegistry) {
      dockerRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
      if (!dockerRegistry) {
        throw new Error('No registry found for snapshot and no default internal registry configured')
      }
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    let retries = 0
    while (retries < 10) {
      try {
        await runnerAdapter.pullSnapshot(internalSnapshotName, dockerRegistry)
      } catch (err) {
        if (err.code !== 'ECONNRESET') {
          throw err
        }
      }
      retries++
      await new Promise((resolve) => setTimeout(resolve, retries * 1000))
    }
  }

  async handleSnapshotRunnerStatePullingSnapshot(snapshotRunner: SnapshotRunner) {
    const runner = await this.runnerRepository.findOneOrFail({
      where: {
        id: snapshotRunner.runnerId,
      },
    })

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

  async handleSnapshotRunnerStateBuildingSnapshot(snapshotRunner: SnapshotRunner) {
    const runner = await this.runnerRepository.findOneOrFail({
      where: {
        id: snapshotRunner.runnerId,
      },
    })

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

    //  get all snapshots
    const snapshots = await this.snapshotRepository.find({
      where: {
        state: SnapshotState.REMOVING,
      },
    })

    await Promise.all(
      snapshots.map(async (snapshot) => {
        await this.snapshotRunnerRepository.update(
          {
            snapshotRef: snapshot.internalName,
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
        const lockKey = `check-snapshot-state-lock-${snapshot.id}`
        if (!(await this.redisLockProvider.lock(lockKey, 720))) {
          return
        }

        try {
          switch (snapshot.state) {
            case SnapshotState.BUILD_PENDING:
              await this.handleSnapshotStateBuildPending(snapshot)
              break
            case SnapshotState.BUILDING:
              await this.handleSnapshotStateBuilding(snapshot)
              break
            case SnapshotState.PENDING:
              await this.handleSnapshotStatePending(snapshot)
              break
            case SnapshotState.PULLING:
              await this.handleSnapshotStatePulling(snapshot)
              break
            case SnapshotState.PENDING_VALIDATION:
              //  temp workaround to avoid race condition between api instances
              {
                let imageName = snapshot.imageName
                if (snapshot.buildInfo) {
                  imageName = snapshot.internalName
                }
                if (!(await this.dockerProvider.imageExists(imageName))) {
                  await this.redisLockProvider.unlock(lockKey)
                  return
                }
              }

              await this.handleSnapshotStatePendingValidation(snapshot)
              break
            case SnapshotState.VALIDATING:
              await this.handleSnapshotStateValidating(snapshot)
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

  @Cron(CronExpression.EVERY_30_MINUTES, {
    name: 'cleanup-local-snapshots',
  })
  async cleanupLocalSnapshots() {
    await this.dockerProvider.imagePrune()
  }

  async removeSnapshotFromRunner(runner: Runner, snapshotRunner: SnapshotRunner) {
    if (snapshotRunner && !snapshotRunner.snapshotRef) {
      //  this should never happen
      this.logger.warn(`Internal snapshot name not found for snapshot runner ${snapshotRunner.id}`)
      return
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)
    const exists = await runnerAdapter.snapshotExists(snapshotRunner.snapshotRef)
    if (exists) {
      await runnerAdapter.removeSnapshot(snapshotRunner.snapshotRef)
    }

    snapshotRunner.state = SnapshotRunnerState.REMOVING
    await this.snapshotRunnerRepository.save(snapshotRunner)
  }

  async handleSnapshotRunnerStateRemoving(snapshotRunner: SnapshotRunner) {
    const runner = await this.runnerRepository.findOne({
      where: {
        id: snapshotRunner.runnerId,
      },
    })
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

  async handleSnapshotStateRemoving(snapshot: Snapshot) {
    const snapshotRunnerItems = await this.snapshotRunnerRepository.find({
      where: {
        snapshotRef: snapshot.internalName,
      },
    })

    if (snapshotRunnerItems.length === 0) {
      await this.snapshotRepository.remove(snapshot)
    }
  }

  async handleSnapshotStateBuildPending(snapshot: Snapshot) {
    await this.updateSnapshotState(snapshot.id, SnapshotState.BUILDING)
  }

  async handleSnapshotStateBuilding(snapshot: Snapshot) {
    // Check if build has timed out
    const timeoutMinutes = 30
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - snapshot.createdAt.getTime() > timeoutMs) {
      await this.updateSnapshotState(snapshot.id, SnapshotState.BUILD_FAILED, 'Timeout while building snapshot')
      return
    }

    // Get build info
    if (!snapshot.buildInfo) {
      await this.updateSnapshotState(snapshot.id, SnapshotState.BUILD_FAILED, 'Missing build information')
      return
    }

    try {
      const excludedRunnerIds = await this.runnerService.getRunnersWithMultipleSnapshotsBuilding()

      // Find a runner to build the snapshot on
      const runner = await this.runnerService.getRandomAvailableRunner({
        excludedRunnerIds: excludedRunnerIds,
      })

      // TODO: get only runners where the base snapshot is available (extract from buildInfo)

      if (!runner) {
        // No ready runners available, retry later
        return
      }

      // Assign the runner ID to the snapshot for tracking build progress
      snapshot.buildRunnerId = runner.id
      await this.snapshotRepository.save(snapshot)

      const registry = await this.dockerRegistryService.getDefaultInternalRegistry()

      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      await runnerAdapter.buildSnapshot(snapshot.buildInfo, snapshot.organizationId, registry, true)

      // save snapshotRunner

      const internalSnapshotName = `${registry.url}/${registry.project}/${snapshot.buildInfo.snapshotRef}`

      snapshot.internalName = internalSnapshotName
      await this.snapshotRepository.save(snapshot)

      // Wait for 30 seconds because of Harbor's delay at making newly created snapshots available
      await new Promise((resolve) => setTimeout(resolve, 30000))

      // Move to next state
      await this.updateSnapshotState(snapshot.id, SnapshotState.PENDING)
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
    let dockerRegistry: DockerRegistry

    await this.updateSnapshotState(snapshot.id, SnapshotState.PULLING)

    let localImageName = snapshot.imageName

    if (snapshot.buildInfo) {
      //  get the default internal registry
      dockerRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
      localImageName = snapshot.internalName
    } else {
      //  find docker registry based on snapshot name and organization id
      dockerRegistry = await this.dockerRegistryService.findOneBySnapshotImageName(
        snapshot.imageName,
        snapshot.organizationId,
      )
    }

    // Use the dockerRegistry for pulling the snapshot
    await this.dockerProvider.pullImage(localImageName, dockerRegistry)
  }

  async handleSnapshotStatePulling(snapshot: Snapshot) {
    const localImageName = snapshot.buildInfo ? snapshot.internalName : snapshot.imageName
    // Check timeout first
    const timeoutMinutes = 15
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - snapshot.createdAt.getTime() > timeoutMs) {
      await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, 'Timeout while pulling snapshot')
      return
    }

    const exists = await this.dockerProvider.imageExists(localImageName)
    if (!exists) {
      //  retry until the snapshot exists (or eventually timeout)
      return
    }

    //  sleep for 30 seconds
    //  workaround for docker snapshot not being ready, but exists
    await new Promise((resolve) => setTimeout(resolve, 30000))

    //  get the organization
    const organization = await this.organizationService.findOne(snapshot.organizationId)
    if (!organization) {
      throw new NotFoundException(`Organization with ID ${snapshot.organizationId} not found`)
    }

    // Check snapshot size
    const snapshotInfo = await this.dockerProvider.getImageInfo(localImageName)
    const MAX_SIZE_GB = organization.maxSnapshotSize

    if (snapshotInfo.sizeGB > MAX_SIZE_GB) {
      await this.updateSnapshotState(
        snapshot.id,
        SnapshotState.ERROR,
        `Snapshot size (${snapshotInfo.sizeGB.toFixed(2)}GB) exceeds maximum allowed size of ${MAX_SIZE_GB}GB`,
      )
      return
    }

    snapshot.size = snapshotInfo.sizeGB
    snapshot.state = SnapshotState.PENDING_VALIDATION

    // Ensure entrypoint is set
    if (!snapshot.entrypoint) {
      if (snapshotInfo.entrypoint) {
        if (Array.isArray(snapshotInfo.entrypoint)) {
          snapshot.entrypoint = snapshotInfo.entrypoint
        } else {
          snapshot.entrypoint = [snapshotInfo.entrypoint]
        }
      } else {
        snapshot.entrypoint = ['sleep', 'infinity']
      }
    }

    await this.snapshotRepository.save(snapshot)
  }

  async handleSnapshotStatePendingValidation(snapshot: Snapshot) {
    try {
      await this.updateSnapshotState(snapshot.id, SnapshotState.VALIDATING)

      await this.validateSnapshotRuntime(snapshot.id)

      if (!snapshot.buildInfo) {
        // Snapshots that have gone through the build process are already in the internal registry
        snapshot.internalName = await this.pushSnapshotToInternalRegistry(snapshot.id)
      }
      const runner = await this.runnerRepository.findOne({
        where: {
          state: RunnerState.READY,
          unschedulable: false,
          used: Raw((alias) => `${alias} < capacity`),
        },
      })
      // Propagate snapshot to one runner so it can be used immediately
      if (runner) {
        await this.propagateSnapshotToRunner(snapshot.internalName, runner)
      }
      await this.updateSnapshotState(snapshot.id, SnapshotState.ACTIVE)

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
    } catch (error) {
      // workaround when app runners don't use a single docker host instance
      if (error.statusCode === 404 || error.message?.toLowerCase().includes('no such snapshot')) {
        return
      }
      await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, error.message)
    }
  }

  async handleSnapshotStateValidating(snapshot: Snapshot) {
    //  check the timeout
    const timeoutMinutes = 10
    const timeoutMs = timeoutMinutes * 60 * 1000
    if (Date.now() - snapshot.createdAt.getTime() > timeoutMs) {
      await this.updateSnapshotState(snapshot.id, SnapshotState.ERROR, 'Timeout while validating snapshot')
      return
    }
  }

  async validateSnapshotRuntime(snapshotId: string): Promise<void> {
    const snapshot = await this.snapshotRepository.findOneOrFail({
      where: {
        id: snapshotId,
      },
    })

    let containerId: string | null = null

    try {
      const localImageName = snapshot.buildInfo ? snapshot.internalName : snapshot.imageName

      // Create and start the container
      containerId = await this.dockerProvider.create(localImageName, snapshot.entrypoint)

      // Wait for 1 minute while checking container state
      const startTime = Date.now()
      const checkDuration = 60 * 1000 // 1 minute in milliseconds

      while (Date.now() - startTime < checkDuration) {
        const isRunning = await this.dockerProvider.isRunning(containerId)
        if (!isRunning) {
          throw new Error('Container exited')
        }
        await new Promise((resolve) => setTimeout(resolve, 2000))
      }
    } catch (error) {
      this.logger.debug('Error validating snapshot runtime:', error)
      throw error
    } finally {
      // Cleanup: Destroy the container
      if (containerId) {
        try {
          await this.dockerProvider.remove(containerId)
        } catch (cleanupError) {
          this.logger.error('Error cleaning up container:', fromAxiosError(cleanupError))
        }
      }
    }
  }

  async pushSnapshotToInternalRegistry(snapshotId: string): Promise<string> {
    const snapshot = await this.snapshotRepository.findOneOrFail({
      where: {
        id: snapshotId,
      },
    })

    const registry = await this.dockerRegistryService.getDefaultInternalRegistry()
    if (!registry) {
      throw new Error('No default internal registry configured')
    }

    //  get tag from snapshot name
    const tag = snapshot.imageName.split(':')[1]
    const internalSnapshotName = `${registry.url.replace(/^(https?:\/\/)/, '')}/${registry.project}/${snapshot.id}:${tag}`

    snapshot.internalName = internalSnapshotName
    await this.snapshotRepository.save(snapshot)

    // Tag the snapshot with the internal registry name
    await this.dockerProvider.tagImage(snapshot.imageName, internalSnapshotName)

    // Push the newly tagged snapshot
    await this.dockerProvider.pushImage(internalSnapshotName, registry)

    return internalSnapshotName
  }

  async retrySnapshotRunnerPull(snapshotRunner: SnapshotRunner) {
    const runner = await this.runnerRepository.findOneOrFail({
      where: {
        id: snapshotRunner.runnerId,
      },
    })

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    const dockerRegistry = await this.dockerRegistryService.getDefaultInternalRegistry()
    //  await this.redis.setex(lockKey, 360, this.instanceId)

    await runnerAdapter.pullSnapshot(snapshotRunner.snapshotRef, dockerRegistry)
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
      const internalNames = oldSnapshots.map((snapshot) => snapshot.internalName).filter((name) => name) // Filter out null/undefined values

      if (internalNames.length > 0) {
        // Set associated SnapshotRunner records to REMOVING state
        const result = await this.snapshotRunnerRepository.update(
          { snapshotRef: In(internalNames) },
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
}
