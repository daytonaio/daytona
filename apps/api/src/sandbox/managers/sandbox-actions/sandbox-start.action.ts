/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { Repository } from 'typeorm'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SYNC_AGAIN, SyncState } from './sandbox.action'
import { SnapshotRunnerState } from '../../enums/snapshot-runner-state.enum'
import { BackupState } from '../../enums/backup-state.enum'
import { RunnerState } from '../../enums/runner-state.enum'
import { DockerProvider } from '../../docker/docker-provider'
import { BuildInfo } from '../../entities/build-info.entity'
import { SnapshotService } from '../../services/snapshot.service'
import { DockerRegistryService } from '../../../docker-registry/services/docker-registry.service'
import { DockerRegistry } from '../../../docker-registry/entities/docker-registry.entity'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { ToolboxService } from '../../services/toolbox.service'
import { InjectRepository } from '@nestjs/typeorm'
import { Snapshot } from '../../entities/snapshot.entity'

@Injectable()
export class SandboxStartAction extends SandboxAction {
  protected readonly logger = new Logger(SandboxStartAction.name)
  constructor(
    protected runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    protected sandboxRepository: Repository<Sandbox>,
    protected toolboxService: ToolboxService,
    protected readonly dockerProvider: DockerProvider,
    protected readonly snapshotService: SnapshotService,
    protected readonly dockerRegistryService: DockerRegistryService,
  ) {
    super(runnerService, runnerAdapterFactory, sandboxRepository, toolboxService)
  }

  async run(sandbox: Sandbox): Promise<SyncState> {
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
      case SandboxState.ERROR: {
        const runner = await this.runnerService.findOne(sandbox.runnerId)
        const runnerAdapter = await this.runnerAdapterFactory.create(runner)

        const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
        if (sandboxInfo.state === SandboxState.STARTED) {
          const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
            id: sandbox.id,
          })
          sandboxToUpdate.state = SandboxState.STARTED
          sandboxToUpdate.backupState = BackupState.NONE

          try {
            const daemonVersion = await runnerAdapter.getSandboxDaemonVersion(sandbox.id)
            sandboxToUpdate.daemonVersion = daemonVersion
          } catch (error) {
            this.logger.error(`Failed to get sandbox daemon version for sandbox ${sandbox.id}:`, error)
          }

          await this.sandboxRepository.save(sandboxToUpdate)
        }
      }
    }

    return DONT_SYNC_AGAIN
  }

  private async handleUnassignedBuildSandbox(sandbox: Sandbox): Promise<SyncState> {
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

  private async handleRunnerSandboxBuildingSnapshotStateOnDesiredStateStart(sandbox: Sandbox): Promise<SyncState> {
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

  private async handleRunnerSandboxUnknownStateOnDesiredStateStart(sandbox: Sandbox): Promise<SyncState> {
    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      return DONT_SYNC_AGAIN
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    let registry: DockerRegistry
    let entrypoint: string[]
    if (!sandbox.buildInfo) {
      //  get internal snapshot name
      const snapshot = await this.snapshotService.getSnapshotByName(sandbox.snapshot, sandbox.organizationId)
      const internalSnapshotName = snapshot.internalName

      registry = await this.dockerRegistryService.findOneBySnapshotImageName(
        internalSnapshotName,
        sandbox.organizationId,
      )
      if (!registry) {
        throw new Error('No registry found for snapshot')
      }

      sandbox.snapshot = internalSnapshotName
      entrypoint = snapshot.entrypoint
    } else {
      sandbox.snapshot = sandbox.buildInfo.snapshotRef
      entrypoint = this.getEntrypointFromDockerfile(sandbox.buildInfo.dockerfileContent)
    }

    await runnerAdapter.createSandbox(sandbox, registry, entrypoint)

    await this.updateSandboxState(sandbox.id, SandboxState.CREATING)
    //  sync states again immediately for sandbox
    return SYNC_AGAIN
  }

  private async handleRunnerSandboxStoppedOrArchivedStateOnDesiredStateStart(sandbox: Sandbox): Promise<SyncState> {
    //  check if sandbox is assigned to a runner and if that runner is unschedulable
    //  if it is, move sandbox to prevRunnerId, and set runnerId to null
    //  this will assign a new runner to the sandbox and restore the sandbox from the latest backup
    if (sandbox.runnerId) {
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      const originalRunnerId = sandbox.runnerId // Store original value

      // if the runner is unschedulable/not ready and sandbox has a valid backup, move sandbox to a new runner
      if (
        (runner.unschedulable || runner.state != RunnerState.READY) &&
        sandbox.backupState === BackupState.COMPLETED
      ) {
        sandbox.prevRunnerId = originalRunnerId
        sandbox.runnerId = null

        const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
          id: sandbox.id,
        })
        sandboxToUpdate.prevRunnerId = originalRunnerId
        sandboxToUpdate.runnerId = null
        await this.sandboxRepository.save(sandboxToUpdate)
      }

      // If the sandbox is on a runner and its backupState is COMPLETED
      // but there are too many running sandboxes on that runner, move it to a less used runner
      if (sandbox.backupState === BackupState.COMPLETED) {
        const usageThreshold = 35
        const runningSandboxsCount = await this.sandboxRepository.count({
          where: {
            runnerId: originalRunnerId,
            state: SandboxState.STARTED,
          },
        })
        if (runningSandboxsCount > usageThreshold) {
          //  TODO: usage should be based on compute usage

          const availableRunners = await this.runnerService.findAvailableRunners({
            region: sandbox.region,
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
          undefined,
          'Sandbox has no runner and backup is not completed',
        )
        return DONT_SYNC_AGAIN
      }

      const registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)
      if (!registry) {
        throw new Error('No registry found for backup')
      }

      const existingBackups = sandbox.existingBackupSnapshots.map((existingSnapshot) => existingSnapshot.snapshotName)
      let validBackup
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
          if (await this.dockerProvider.checkImageExistsInRegistry(validBackup, registry)) {
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
        await this.updateSandboxState(sandbox.id, SandboxState.ERROR, undefined, 'No valid backup snapshot found')
        return SYNC_AGAIN
      }

      //  make sure we pick a runner that has the base snapshot
      let baseSnapshot: Snapshot | null = null
      try {
        baseSnapshot = await this.snapshotService.getSnapshotByName(sandbox.snapshot, sandbox.organizationId)
      } catch (e) {
        if (e instanceof NotFoundException) {
          //  if the base snapshot is not found, we'll use any available runner later
        } else {
          //  for all other errors, throw them
          throw e
        }
      }

      const snapshotRef = baseSnapshot ? baseSnapshot.internalName : null

      let availableRunners = []
      const runnersWithBaseSnapshot = await this.runnerService.findAvailableRunners({
        region: sandbox.region,
        sandboxClass: sandbox.class,
        snapshotRef,
      })
      if (runnersWithBaseSnapshot.length > 0) {
        availableRunners = runnersWithBaseSnapshot
      } else {
        //  if no runner has the base snapshot, get all available runners
        availableRunners = await this.runnerService.findAvailableRunners({
          region: sandbox.region,
          sandboxClass: sandbox.class,
        })
      }

      //  check if we have any available runners after filtering
      if (availableRunners.length === 0) {
        await this.updateSandboxState(
          sandbox.id,
          SandboxState.ERROR,
          undefined,
          'No available runners found for sandbox restoration',
        )
        return DONT_SYNC_AGAIN
      }

      //  get random runner from available runners
      const randomRunnerIndex = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1) + min)
      const runnerId = availableRunners[randomRunnerIndex(0, availableRunners.length - 1)].id

      const runner = await this.runnerService.findOne(runnerId)

      //  verify the runner is still available and ready
      if (!runner || runner.state !== RunnerState.READY || runner.unschedulable || runner.used >= runner.capacity) {
        this.logger.warn(`Selected runner ${runnerId} is no longer available, retrying sandbox assignment`)
        return SYNC_AGAIN
      }

      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      await this.updateSandboxState(sandbox.id, SandboxState.RESTORING, runnerId)

      sandbox.snapshot = validBackup
      await runnerAdapter.createSandbox(sandbox, registry)
    } else {
      // if sandbox has runner, start sandbox
      const runner = await this.runnerService.findOne(sandbox.runnerId)
      const runnerAdapter = await this.runnerAdapterFactory.create(runner)

      await runnerAdapter.startSandbox(sandbox.id)

      await this.updateSandboxState(sandbox.id, SandboxState.STARTING)

      return SYNC_AGAIN
    }

    return SYNC_AGAIN
  }

  //  used to check if sandbox is pulling snapshot on runner and update sandbox state accordingly
  private async handleRunnerSandboxPullingSnapshotStateCheck(sandbox: Sandbox): Promise<SyncState> {
    //  edge case when sandbox is being transferred to a new runner
    if (!sandbox.runnerId) {
      return SYNC_AGAIN
    }

    const runner = await this.runnerService.findOne(sandbox.runnerId)
    const runnerAdapter = await this.runnerAdapterFactory.create(runner)
    const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)

    if (sandboxInfo.state === SandboxState.PULLING_SNAPSHOT) {
      await this.updateSandboxState(sandbox.id, SandboxState.PULLING_SNAPSHOT)
    } else if (sandboxInfo.state === SandboxState.ERROR) {
      await this.updateSandboxState(sandbox.id, SandboxState.ERROR)
    } else {
      await this.updateSandboxState(sandbox.id, SandboxState.STARTING)
    }

    return SYNC_AGAIN
  }

  //  used to check if sandbox is started on runner and update sandbox state accordingly
  //  also used to handle the case where a sandbox is started on a runner and then transferred to a new runner
  private async handleRunnerSandboxStartedStateCheck(sandbox: Sandbox): Promise<SyncState> {
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
          sandbox.backupState = BackupState.NONE

          const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
            id: sandbox.id,
          })
          sandboxToUpdate.state = SandboxState.STARTED
          sandboxToUpdate.backupState = BackupState.NONE
          if (daemonVersion) {
            sandboxToUpdate.daemonVersion = daemonVersion
          }
          await this.sandboxRepository.save(sandboxToUpdate)
        } else {
          await this.updateSandboxState(sandbox.id, SandboxState.STARTED, undefined, undefined, daemonVersion)
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

            sandbox.prevRunnerId = null

            const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
              id: sandbox.id,
            })

            sandboxToUpdate.prevRunnerId = null

            await this.sandboxRepository.save(sandboxToUpdate)
          } catch (error) {
            this.logger.error(`Failed to cleanup sandbox ${sandbox.id} on previous runner ${runner.id}:`, error)
          }
        }
        break
      }
      case SandboxState.ERROR: {
        await this.updateSandboxState(sandbox.id, SandboxState.ERROR)
        break
      }
    }

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

  // Initiates the snapshot build on the runner and creates an SnapshotRunner depending on the result
  async buildOnRunner(buildInfo: BuildInfo, runnerId: string, organizationId: string) {
    const runner = await this.runnerService.findOne(runnerId)
    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    let retries = 0

    while (retries < 10) {
      try {
        await runnerAdapter.buildSnapshot(buildInfo, organizationId)
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

    const exists = await runnerAdapter.snapshotExists(buildInfo.snapshotRef)
    let state = SnapshotRunnerState.BUILDING_SNAPSHOT
    if (exists) {
      state = SnapshotRunnerState.READY
    }

    await this.runnerService.createSnapshotRunner(runnerId, buildInfo.snapshotRef, state)
  }
}
