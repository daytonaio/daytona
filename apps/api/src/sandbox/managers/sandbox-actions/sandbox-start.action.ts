/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { Repository } from 'typeorm'
import { RECOVERY_ERROR_SUBSTRINGS } from '../../constants/errors-for-recovery'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SYNC_AGAIN, SyncState } from './sandbox.action'
import { SnapshotRunnerState } from '../../enums/snapshot-runner-state.enum'
import { BackupState } from '../../enums/backup-state.enum'
import { RunnerState } from '../../enums/runner-state.enum'
import { BuildInfo } from '../../entities/build-info.entity'
import { SnapshotService } from '../../services/snapshot.service'
import { DockerRegistryService } from '../../../docker-registry/services/docker-registry.service'
import { DockerRegistry } from '../../../docker-registry/entities/docker-registry.entity'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { InjectRepository } from '@nestjs/typeorm'
import { Snapshot } from '../../entities/snapshot.entity'
import { OrganizationService } from '../../../organization/services/organization.service'
import { TypedConfigService } from '../../../config/typed-config.service'
import { Runner } from '../../entities/runner.entity'
import { Organization } from '../../../organization/entities/organization.entity'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'

@Injectable()
export class SandboxStartAction extends SandboxAction {
  protected readonly logger = new Logger(SandboxStartAction.name)
  constructor(
    protected runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    protected sandboxRepository: Repository<Sandbox>,
    protected readonly snapshotService: SnapshotService,
    protected readonly dockerRegistryService: DockerRegistryService,
    protected readonly organizationService: OrganizationService,
    protected readonly configService: TypedConfigService,
    protected readonly redisLockProvider: RedisLockProvider,
  ) {
    super(runnerService, runnerAdapterFactory, sandboxRepository, redisLockProvider)
  }

  async run(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState> {
    switch (sandbox.state) {
      case SandboxState.PULLING_SNAPSHOT: {
        if (!sandbox.runnerId) {
          // Using the PULLING_SNAPSHOT state for the case where the runner isn't assigned yet as well
          return this.handleUnassignedRunnerSandbox(sandbox, lockCode)
        } else {
          return this.handleRunnerSandboxStartedStateCheck(sandbox, lockCode)
        }
      }
      case SandboxState.PENDING_BUILD: {
        return this.handleUnassignedRunnerSandbox(sandbox, lockCode, true)
      }
      case SandboxState.BUILDING_SNAPSHOT: {
        return this.handleRunnerSandboxBuildingSnapshotStateOnDesiredStateStart(sandbox, lockCode)
      }
      case SandboxState.UNKNOWN: {
        return this.handleRunnerSandboxUnknownStateOnDesiredStateStart(sandbox, lockCode)
      }
      case SandboxState.ARCHIVED:
      case SandboxState.ARCHIVING:
      case SandboxState.STOPPED: {
        return this.handleRunnerSandboxStoppedOrArchivedStateOnDesiredStateStart(sandbox, lockCode)
      }
      case SandboxState.RESTORING:
      case SandboxState.CREATING: {
        return this.handleRunnerSandboxPullingSnapshotStateCheck(sandbox, lockCode)
      }
      case SandboxState.STARTING: {
        return this.handleRunnerSandboxStartedStateCheck(sandbox, lockCode)
      }
      case SandboxState.ERROR: {
        const runner = await this.runnerService.findOne(sandbox.runnerId)
        const runnerAdapter = await this.runnerAdapterFactory.create(runner)

        const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
        if (sandboxInfo.state === SandboxState.STARTED) {
          let daemonVersion: string | undefined
          try {
            daemonVersion = await runnerAdapter.getSandboxDaemonVersion(sandbox.id)
          } catch (error) {
            this.logger.error(`Failed to get sandbox daemon version for sandbox ${sandbox.id}:`, error)
          }

          await this.updateSandboxState(
            sandbox.id,
            SandboxState.STARTED,
            lockCode,
            undefined,
            undefined,
            daemonVersion,
            BackupState.NONE,
          )
          return DONT_SYNC_AGAIN
        }
      }
    }

    return DONT_SYNC_AGAIN
  }

  private async handleRunnerSandboxBuildingSnapshotStateOnDesiredStateStart(
    sandbox: Sandbox,
    lockCode: LockCode,
  ): Promise<SyncState> {
    // Check for timeout - allow up to 60 minutes since the last sandbox update
    const timeoutMinutes = 60
    const timeoutMs = timeoutMinutes * 60 * 1000

    if (sandbox.updatedAt && Date.now() - sandbox.updatedAt.getTime() > timeoutMs) {
      await this.updateSandboxState(
        sandbox.id,
        SandboxState.ERROR,
        lockCode,
        undefined,
        'Timeout while building snapshot on runner',
      )
      return DONT_SYNC_AGAIN
    }

    const snapshotRunner = await this.runnerService.getSnapshotRunner(sandbox.runnerId, sandbox.buildInfo.snapshotRef)
    if (snapshotRunner) {
      switch (snapshotRunner.state) {
        case SnapshotRunnerState.READY: {
          // TODO: "UNKNOWN" should probably be changed to something else
          await this.updateSandboxState(sandbox.id, SandboxState.UNKNOWN, lockCode)
          return SYNC_AGAIN
        }
        case SnapshotRunnerState.ERROR: {
          await this.updateSandboxState(
            sandbox.id,
            SandboxState.BUILD_FAILED,
            lockCode,
            undefined,
            snapshotRunner.errorReason,
          )
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

  private async handleUnassignedRunnerSandbox(
    sandbox: Sandbox,
    lockCode: LockCode,
    isBuild = false,
  ): Promise<SyncState> {
    // Get snapshot reference based on whether it's a pull or build operation
    let snapshotRef: string

    if (isBuild) {
      snapshotRef = sandbox.buildInfo.snapshotRef
    } else {
      const snapshot = await this.snapshotService.getSnapshotByName(sandbox.snapshot, sandbox.organizationId)
      snapshotRef = snapshot.ref
    }

    // Try to assign an available runner with the snapshot already available
    try {
      const runner = await this.runnerService.getRandomAvailableRunner({
        regions: [sandbox.region],
        sandboxClass: sandbox.class,
        snapshotRef: snapshotRef,
      })
      if (runner) {
        await this.updateSandboxState(sandbox.id, SandboxState.UNKNOWN, lockCode, runner.id)
        return SYNC_AGAIN
      }
    } catch {
      // Continue to next assignment method
    }

    // Try to assign an available runner that is currently processing the snapshot
    const snapshotRunners = await this.runnerService.getSnapshotRunners(snapshotRef)
    const targetState = isBuild ? SnapshotRunnerState.BUILDING_SNAPSHOT : SnapshotRunnerState.PULLING_SNAPSHOT
    const targetSandboxState = isBuild ? SandboxState.BUILDING_SNAPSHOT : SandboxState.PULLING_SNAPSHOT
    const errorSandboxState = isBuild ? SandboxState.BUILD_FAILED : SandboxState.ERROR

    for (const snapshotRunner of snapshotRunners) {
      // Consider removing the runner usage rate check or improving it
      const runner = await this.runnerService.findOne(snapshotRunner.runnerId)
      if (runner.availabilityScore >= this.configService.getOrThrow('runnerUsage.declarativeBuildScoreThreshold')) {
        if (snapshotRunner.state === targetState) {
          await this.updateSandboxState(sandbox.id, targetSandboxState, lockCode, runner.id)
          return SYNC_AGAIN
        } else if (snapshotRunner.state === SnapshotRunnerState.ERROR) {
          await this.updateSandboxState(sandbox.id, errorSandboxState, lockCode, runner.id, snapshotRunner.errorReason)
          return DONT_SYNC_AGAIN
        }
      }
    }

    // Get excluded runner IDs based on operation type
    const excludedRunnerIds = await (isBuild
      ? this.runnerService.getRunnersWithMultipleSnapshotsBuilding()
      : this.runnerService.getRunnersWithMultipleSnapshotsPulling())

    // Try to assign an available runner to start processing the snapshot
    let runner: Runner

    try {
      runner = await this.runnerService.getRandomAvailableRunner({
        regions: [sandbox.region],
        sandboxClass: sandbox.class,
        excludedRunnerIds: excludedRunnerIds,
      })
    } catch {
      // TODO: reconsider the timeout here
      // No runners available, wait for 3 seconds and retry
      await new Promise((resolve) => setTimeout(resolve, 3000))
      return SYNC_AGAIN
    }

    if (isBuild) {
      this.buildOnRunner(sandbox.buildInfo, runner, sandbox.organizationId)
      await this.updateSandboxState(sandbox.id, SandboxState.BUILDING_SNAPSHOT, lockCode, runner.id)
    } else {
      const snapshot = await this.snapshotService.getSnapshotByName(sandbox.snapshot, sandbox.organizationId)
      await this.runnerService.createSnapshotRunnerEntry(runner.id, snapshot.ref, SnapshotRunnerState.PULLING_SNAPSHOT)
      this.pullSnapshotToRunner(snapshot, runner)
      await this.updateSandboxState(sandbox.id, SandboxState.PULLING_SNAPSHOT, lockCode, runner.id)
    }

    return SYNC_AGAIN
  }

  async pullSnapshotToRunner(snapshot: Snapshot, runner: Runner) {
    const registry = await this.dockerRegistryService.findInternalRegistryBySnapshotRef(snapshot.ref, runner.region)
    if (!registry) {
      throw new Error('No internal registry found for sandbox snapshot')
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    let retries = 0
    while (retries < 10) {
      try {
        await runnerAdapter.pullSnapshot(snapshot.ref, registry)
        break
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

  // Initiates the snapshot build on the runner and creates an SnapshotRunner depending on the result
  async buildOnRunner(buildInfo: BuildInfo, runner: Runner, organizationId: string) {
    const runnerAdapter = await this.runnerAdapterFactory.create(runner)
    const sourceRegistry = await this.dockerRegistryService.getDefaultDockerHubRegistry()

    let retries = 0

    while (retries < 10) {
      try {
        await runnerAdapter.buildSnapshot(buildInfo, organizationId, sourceRegistry ? [sourceRegistry] : undefined)
        break
      } catch (err) {
        if (err.code !== 'ECONNRESET') {
          await this.runnerService.createSnapshotRunnerEntry(
            runner.id,
            buildInfo.snapshotRef,
            SnapshotRunnerState.ERROR,
            err.message,
          )
          return
        }
        if (++retries >= 10) {
          throw err
        }
        await new Promise((resolve) => setTimeout(resolve, retries * 1000))
      }
    }

    if (retries === 10) {
      await this.runnerService.createSnapshotRunnerEntry(
        runner.id,
        buildInfo.snapshotRef,
        SnapshotRunnerState.ERROR,
        'Timeout while building',
      )
      return
    }

    const exists = await runnerAdapter.snapshotExists(buildInfo.snapshotRef)
    let state = SnapshotRunnerState.BUILDING_SNAPSHOT
    if (exists) {
      state = SnapshotRunnerState.READY
    }

    await this.runnerService.createSnapshotRunnerEntry(runner.id, buildInfo.snapshotRef, state)
  }

  private async handleRunnerSandboxUnknownStateOnDesiredStateStart(
    sandbox: Sandbox,
    lockCode: LockCode,
  ): Promise<SyncState> {
    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      return DONT_SYNC_AGAIN
    }

    const organization = await this.organizationService.findOne(sandbox.organizationId)

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    let registry: DockerRegistry
    let entrypoint: string[]
    if (!sandbox.buildInfo) {
      //  get internal snapshot name
      const snapshot = await this.snapshotService.getSnapshotByName(sandbox.snapshot, sandbox.organizationId)
      const snapshotRef = snapshot.ref

      registry = await this.dockerRegistryService.findInternalRegistryBySnapshotRef(snapshotRef, sandbox.region)
      if (!registry) {
        throw new Error('No registry found for snapshot')
      }

      sandbox.snapshot = snapshotRef
      entrypoint = snapshot.entrypoint
    } else {
      sandbox.snapshot = sandbox.buildInfo.snapshotRef
      entrypoint = this.snapshotService.getEntrypointFromDockerfile(sandbox.buildInfo.dockerfileContent)
    }

    let metadata: { [key: string]: string } | undefined = undefined
    if (organization) {
      metadata = {
        limitNetworkEgress: String(organization.sandboxLimitedNetworkEgress),
        organizationId: organization.id,
        organizationName: organization.name,
        sandboxName: sandbox.name,
      }
    }

    await runnerAdapter.createSandbox(sandbox, registry, entrypoint, metadata)

    await this.updateSandboxState(sandbox.id, SandboxState.CREATING, lockCode)
    //  sync states again immediately for sandbox
    return SYNC_AGAIN
  }

  private async handleRunnerSandboxStoppedOrArchivedStateOnDesiredStateStart(
    sandbox: Sandbox,
    lockCode: LockCode,
  ): Promise<SyncState> {
    const organization = await this.organizationService.findOne(sandbox.organizationId)

    //  check if sandbox is assigned to a runner and if that runner is unschedulable
    //  if it is, move sandbox to prevRunnerId, and set runnerId to null
    //  this will assign a new runner to the sandbox and restore the sandbox from the latest backup
    if (sandbox.runnerId) {
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      const originalRunnerId = sandbox.runnerId // Store original value

      const startScoreThreshold = this.configService.get('runnerUsage.startScoreThreshold') || 0

      const shouldMoveToNewRunner =
        (runner.unschedulable || runner.state != RunnerState.READY || runner.availabilityScore < startScoreThreshold) &&
        sandbox.backupState === BackupState.COMPLETED

      // if the runner is unschedulable/not ready and sandbox has a valid backup, move sandbox to a new runner
      if (shouldMoveToNewRunner) {
        sandbox.prevRunnerId = originalRunnerId
        sandbox.runnerId = null

        await this.sandboxRepository.update(sandbox.id, {
          prevRunnerId: originalRunnerId,
          runnerId: null,
        })
      }

      // If the sandbox is on a runner and its backupState is COMPLETED
      // but there are too many running sandboxes on that runner, move it to a less used runner
      if (sandbox.backupState === BackupState.COMPLETED) {
        if (runner.availabilityScore < this.configService.getOrThrow('runnerUsage.availabilityScoreThreshold')) {
          const availableRunners = await this.runnerService.findAvailableRunners({
            regions: [sandbox.region],
            sandboxClass: sandbox.class,
          })
          const lessUsedRunners = availableRunners.filter((runner) => runner.id !== originalRunnerId)

          //  temp workaround to move sandboxes to less used runner
          if (lessUsedRunners.length > 0) {
            await this.sandboxRepository.update(sandbox.id, {
              runnerId: null,
              prevRunnerId: originalRunnerId,
            })
            try {
              const runnerAdapter = await this.runnerAdapterFactory.create(runner)
              await runnerAdapter.removeDestroyedSandbox(sandbox.id)
            } catch (e) {
              this.logger.error(`Failed to cleanup sandbox ${sandbox.id} on previous runner ${runner.id}:`, e)
            }
            sandbox.prevRunnerId = originalRunnerId
            sandbox.runnerId = null
          }
        }
      }
    }

    if (sandbox.runnerId === null) {
      //  if sandbox has no runner, check if backup is completed
      //  if not, set sandbox to error
      //  if backup is completed, get random available runner and start sandbox
      //  use the backup to start the sandbox

      if (sandbox.backupState !== BackupState.COMPLETED) {
        await this.updateSandboxState(
          sandbox.id,
          SandboxState.ERROR,
          lockCode,
          undefined,
          'Sandbox has no runner and backup is not completed',
        )
        return DONT_SYNC_AGAIN
      }

      const syncCheck = await this.restoreSandboxOnNewRunner(sandbox, lockCode, organization, sandbox.prevRunnerId)
      if (syncCheck !== null) {
        return syncCheck
      }
    } else {
      // if sandbox has runner, start sandbox
      const runner = await this.runnerService.findOne(sandbox.runnerId)

      if (runner.state !== RunnerState.READY) {
        return DONT_SYNC_AGAIN
      }

      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      let metadata: { [key: string]: string } | undefined = undefined
      if (organization) {
        metadata = {
          limitNetworkEgress: String(organization.sandboxLimitedNetworkEgress),
        }
      }

      try {
        await runnerAdapter.startSandbox(sandbox.id, metadata)
      } catch (error) {
        // Check against a list of substrings that should trigger an automatic recovery
        if (error?.message) {
          const matchesRecovery = RECOVERY_ERROR_SUBSTRINGS.some((substring) =>
            error.message.toLowerCase().includes(substring.toLowerCase()),
          )
          if (matchesRecovery) {
            try {
              await this.restoreSandboxOnNewRunner(sandbox, lockCode, organization, sandbox.runnerId, true)
              this.logger.warn(`Sandbox ${sandbox.id} transferred to a new runner`)
              return SYNC_AGAIN
            } catch (restoreError) {
              this.logger.warn(`Sandbox ${sandbox.id} recovery attempt failed:`, restoreError.message)
            }
          }
        }
        throw error
      }

      await this.updateSandboxState(sandbox.id, SandboxState.STARTING, lockCode)
      return SYNC_AGAIN
    }

    return SYNC_AGAIN
  }

  //  used to check if sandbox is pulling snapshot on runner and update sandbox state accordingly
  private async handleRunnerSandboxPullingSnapshotStateCheck(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState> {
    //  edge case when sandbox is being transferred to a new runner
    if (!sandbox.runnerId) {
      return SYNC_AGAIN
    }

    const runner = await this.runnerService.findOne(sandbox.runnerId)
    const runnerAdapter = await this.runnerAdapterFactory.create(runner)
    const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)

    if (sandboxInfo.state === SandboxState.PULLING_SNAPSHOT) {
      await this.updateSandboxState(sandbox.id, SandboxState.PULLING_SNAPSHOT, lockCode)
    } else if (sandboxInfo.state === SandboxState.ERROR) {
      await this.updateSandboxState(
        sandbox.id,
        SandboxState.ERROR,
        lockCode,
        undefined,
        'Sandbox is in error state on runner',
      )
    } else if (sandboxInfo.state === SandboxState.UNKNOWN) {
      await this.updateSandboxState(sandbox.id, SandboxState.UNKNOWN, lockCode)
    } else {
      await this.updateSandboxState(sandbox.id, SandboxState.STARTING, lockCode)
    }

    return SYNC_AGAIN
  }

  //  used to check if sandbox is started on runner and update sandbox state accordingly
  //  also used to handle the case where a sandbox is started on a runner and then transferred to a new runner
  private async handleRunnerSandboxStartedStateCheck(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState> {
    const runner = await this.runnerService.findOne(sandbox.runnerId)
    const runnerAdapter = await this.runnerAdapterFactory.create(runner)
    const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)

    switch (sandboxInfo.state) {
      case SandboxState.STARTED: {
        let daemonVersion: string | undefined
        try {
          daemonVersion = await runnerAdapter.getSandboxDaemonVersion(sandbox.id)
        } catch (error) {
          this.logger.error(`Failed to get sandbox daemon version for sandbox ${sandbox.id}:`, error)
        }

        //  if previous backup state is error or completed, set backup state to none
        if ([BackupState.ERROR, BackupState.COMPLETED].includes(sandbox.backupState)) {
          await this.updateSandboxState(
            sandbox.id,
            SandboxState.STARTED,
            lockCode,
            undefined,
            undefined,
            daemonVersion,
            BackupState.NONE,
          )
          return DONT_SYNC_AGAIN
        } else {
          await this.updateSandboxState(sandbox.id, SandboxState.STARTED, lockCode, undefined, undefined, daemonVersion)

          //  if sandbox was transferred to a new runner, remove it from the old runner
          if (sandbox.prevRunnerId) {
            await this.removeSandboxFromPreviousRunner(sandbox)
          }

          return DONT_SYNC_AGAIN
        }
        break
      }
      case SandboxState.STARTING:
        if (await this.checkTimeoutError(sandbox, 5, 'Timeout while starting sandbox')) {
          return DONT_SYNC_AGAIN
        }
        break
      case SandboxState.RESTORING:
        if (await this.checkTimeoutError(sandbox, 30, 'Timeout while starting sandbox')) {
          return DONT_SYNC_AGAIN
        }
        break
      case SandboxState.CREATING: {
        if (await this.checkTimeoutError(sandbox, 15, 'Timeout while creating sandbox')) {
          return DONT_SYNC_AGAIN
        }
        break
      }
      case SandboxState.UNKNOWN: {
        await this.updateSandboxState(sandbox.id, SandboxState.UNKNOWN, lockCode)
        break
      }
      case SandboxState.ERROR: {
        await this.updateSandboxState(
          sandbox.id,
          SandboxState.ERROR,
          lockCode,
          undefined,
          'Sandbox entered error state on runner during startup wait loop',
        )
        break
      }
      case SandboxState.DESTROYED: {
        this.logger.warn(
          `Sandbox ${sandbox.id} is in destroyed state while starting on runner ${sandbox.runnerId}, prev runner ${sandbox.prevRunnerId}`,
        )
        await this.checkTimeoutError(
          sandbox,
          15,
          'Timeout while starting sandbox: Sandbox is in unknown state on runner',
        )
        return DONT_SYNC_AGAIN
      }
      // also any other state that is not STARTED
      default: {
        this.logger.error(`Sandbox ${sandbox.id} is in unexpected state ${sandboxInfo.state}`)
        await this.updateSandboxState(
          sandbox.id,
          SandboxState.ERROR,
          lockCode,
          undefined,
          `Sandbox is in unexpected state: ${sandboxInfo.state}`,
        )
        break
      }
    }

    return SYNC_AGAIN
  }

  private async checkTimeoutError(sandbox: Sandbox, timeoutMinutes: number, errorReason: string): Promise<boolean> {
    if (
      sandbox.lastActivityAt &&
      new Date(sandbox.lastActivityAt).getTime() < Date.now() - 1000 * 60 * timeoutMinutes
    ) {
      sandbox.state = SandboxState.ERROR
      sandbox.errorReason = errorReason
      await this.sandboxRepository.save(sandbox)
      return true
    }
    return false
  }

  private async restoreSandboxOnNewRunner(
    sandbox: Sandbox,
    lockCode: LockCode,
    organization: Organization,
    excludedRunnerId: string,
    isRecovery?: boolean,
  ): Promise<SyncState | null> {
    let lockKey: string | null = null

    // Recovery lock to prevent frequent automatic restore attempts
    if (isRecovery) {
      lockKey = `sandbox-${sandbox.id}-restored-cooldown`
      const sixHoursInSeconds = 6 * 60 * 60
      const acquired = await this.redisLockProvider.lock(lockKey, sixHoursInSeconds)
      if (!acquired) {
        return null
      }
    }

    if (!sandbox.backupRegistryId) {
      throw new Error('No registry found for backup')
    }

    const registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)
    if (!registry) {
      throw new Error('No registry found for backup')
    }

    const existingBackups = sandbox.existingBackupSnapshots
      .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
      .map((existingSnapshot) => existingSnapshot.snapshotName)

    let validBackup: string | null = null
    let exists = false

    while (existingBackups.length > 0) {
      try {
        if (!validBackup) {
          //  last snapshot is the current snapshot, so we don't need to check it
          //  just in case, we'll use the value from the backupSnapshot property
          validBackup = sandbox.backupSnapshot
          existingBackups.pop()
        } else {
          validBackup = existingBackups.pop()
        }
        if (await this.dockerRegistryService.checkImageExistsInRegistry(validBackup, registry)) {
          exists = true
          break
        }
      } catch (error) {
        this.logger.error(
          `Failed to check if backup snapshot ${sandbox.backupSnapshot} exists in registry ${registry.id}:`,
          error,
        )
      }
    }

    if (!exists) {
      if (!isRecovery) {
        await this.updateSandboxState(
          sandbox.id,
          SandboxState.ERROR,
          lockCode,
          undefined,
          'No valid backup snapshot found',
        )
      } else {
        throw new Error('No valid backup snapshot found')
      }
      return SYNC_AGAIN
    }

    //  make sure we pick a runner that has the base snapshot
    let baseSnapshot: Snapshot | null = null
    if (sandbox.snapshot) {
      try {
        baseSnapshot = await this.snapshotService.getSnapshotByName(sandbox.snapshot, sandbox.organizationId)
      } catch (e) {
        if (e instanceof NotFoundException) {
          //  if the base snapshot is not found, we'll use any available runner later
        } else {
          if (isRecovery) {
            return SYNC_AGAIN
          }
          //  for all other errors, throw them
          throw e
        }
      }
    }

    const snapshotRef = baseSnapshot ? baseSnapshot.ref : null

    let availableRunners: Runner[] = []

    const excludedRunnerIds: string[] = excludedRunnerId ? [excludedRunnerId] : []

    const runnersWithBaseSnapshot: Runner[] = snapshotRef
      ? await this.runnerService.findAvailableRunners({
          regions: [sandbox.region],
          sandboxClass: sandbox.class,
          snapshotRef,
          excludedRunnerIds,
        })
      : []
    if (runnersWithBaseSnapshot.length > 0) {
      availableRunners = runnersWithBaseSnapshot
    } else {
      //  if no runner has the base snapshot, get all available runners
      availableRunners = await this.runnerService.findAvailableRunners({
        regions: [sandbox.region],
        excludedRunnerIds,
      })
    }

    //  check if we have any available runners after filtering
    if (availableRunners.length === 0) {
      // Sync state again later. Runners are unavailable
      if (isRecovery) {
        await this.redisLockProvider.unlock(lockKey)
      }
      return DONT_SYNC_AGAIN
    }

    //  get random runner from available runners
    const randomRunnerIndex = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1) + min)
    const runner = availableRunners[randomRunnerIndex(0, availableRunners.length - 1)]

    //  verify the runner is still available and ready
    if (!runner || runner.state !== RunnerState.READY || runner.unschedulable) {
      this.logger.warn(`Selected runner ${runner?.id || 'null'} is no longer available, retrying sandbox assignment`)
      if (isRecovery) {
        await this.redisLockProvider.unlock(lockKey)
      }
      return SYNC_AGAIN
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    await this.updateSandboxState(sandbox.id, SandboxState.RESTORING, lockCode, runner.id)

    sandbox.snapshot = validBackup

    let metadata: { [key: string]: string } | undefined = undefined
    if (organization) {
      metadata = {
        limitNetworkEgress: String(organization.sandboxLimitedNetworkEgress),
        organizationId: organization.id,
        organizationName: organization.name,
        sandboxName: sandbox.name,
      }
    }

    await runnerAdapter.createSandbox(sandbox, registry, undefined, metadata)
    return null
  }

  private async removeSandboxFromPreviousRunner(sandbox: Sandbox): Promise<void> {
    const runner = await this.runnerService.findOne(sandbox.prevRunnerId)
    if (!runner) {
      this.logger.warn(`Previously assigned runner ${sandbox.prevRunnerId} for sandbox ${sandbox.id} not found`)

      await this.sandboxRepository.update(sandbox.id, {
        prevRunnerId: null,
      })
      return
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    try {
      // First try to destroy the sandbox
      await runnerAdapter.destroySandbox(sandbox.id)

      // Wait for sandbox to be destroyed before removing
      let retries = 0
      while (retries < 10) {
        try {
          const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
          if (sandboxInfo.state === SandboxState.DESTROYED) {
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
      await runnerAdapter.removeDestroyedSandbox(sandbox.id)

      await this.sandboxRepository.update(sandbox.id, {
        prevRunnerId: null,
      })
    } catch (error) {
      this.logger.error(`Failed to cleanup sandbox ${sandbox.id} on previous runner ${runner.id}:`, error)
    }
  }
}
