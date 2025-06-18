import { Injectable, Logger } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SYNC_AGAIN, SyncState } from '../sandbox.manager'
import { SnapshotRunnerState } from '../../enums/snapshot-runner-state.enum'
import { BackupState } from '../../enums/backup-state.enum'
import { RunnerState } from '../../enums/runner-state.enum'
import { RunnerSandboxState } from '../../runner-adapter/runnerAdapter'
import { CreateSandboxDTO } from '@daytonaio/runner-api-client'
import { DockerProvider } from '../../docker/docker-provider'
import { BuildInfo } from '../../entities/build-info.entity'
import { SnapshotService } from '../../services/snapshot.service'
import { DockerRegistryService } from '../../../docker-registry/services/docker-registry.service'

@Injectable()
export class SandboxStartAction extends SandboxAction {
  protected readonly logger = new Logger(SandboxStartAction.name)
  constructor(
    runnerService,
    runnerSandboxAdapterFactory,
    sandboxRepository,
    protected readonly dockerProvider: DockerProvider,
    protected readonly snapshotService: SnapshotService,
    protected readonly dockerRegistryService: DockerRegistryService,
  ) {
    super(runnerService, runnerSandboxAdapterFactory, sandboxRepository)
  }

  async run(sandbox: Sandbox) {
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
        const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)
        const sandboxInfo = await runnerAdapter.info(sandbox.id)
        if (sandboxInfo.state === RunnerSandboxState.STARTED) {
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
      //  console.debug(`Runner ${runner.id} is not ready`);
      return DONT_SYNC_AGAIN
    }

    let createSandboxDto: CreateSandboxDTO = {
      id: sandbox.id,
      osUser: sandbox.osUser,
      snapshot: '',
      userId: sandbox.organizationId,
      storageQuota: sandbox.disk,
      memoryQuota: sandbox.mem,
      cpuQuota: sandbox.cpu,
      env: sandbox.env,
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

    const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)
    await runnerAdapter.init(runner)
    await runnerAdapter.start(sandbox.id)
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
              const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)
              await runnerAdapter.removeDestroyed(sandbox.id)
            } catch (e) {
              this.logger.error(`Failed to cleanup sandbox ${sandbox.id} on previous runner ${runner.id}:`)
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
        return DONT_SYNC_AGAIN
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

      const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)

      await runnerAdapter.create(sandbox.id)

      await this.updateSandboxState(sandbox.id, SandboxState.RESTORING, runnerId)
    } else {
      // if sandbox has runner, start sandbox
      const runner = await this.runnerService.findOne(sandbox.runnerId)

      const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)

      await runnerAdapter.start(sandbox.id)

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
    const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)
    const sandboxInfo = await runnerAdapter.info(sandbox.id)

    if (sandboxInfo.state === RunnerSandboxState.PULLING_SNAPSHOT) {
      await this.updateSandboxState(sandbox.id, SandboxState.PULLING_SNAPSHOT)
    } else if (sandboxInfo.state === RunnerSandboxState.ERROR) {
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
    const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)
    const sandboxInfo = await runnerAdapter.info(sandbox.id)

    switch (sandboxInfo.state) {
      case RunnerSandboxState.STARTED: {
        let daemonVersion: string | undefined
        try {
          daemonVersion = await runnerAdapter.getDaemonVersion(sandbox)
        } catch (e) {
          this.logger.error(`Failed to get sandbox daemon version for sandbox ${sandbox.id}:`, e)
        }
        if (daemonVersion != sandbox.daemonVersion) {
          sandbox.daemonVersion = daemonVersion
          await this.sandboxRepository.update(sandbox.id, {
            daemonVersion: daemonVersion,
          })
        }

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
          const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)
          try {
            // First try to destroy the sandbox
            await runnerAdapter.destroy(sandbox.id)

            // Wait for sandbox to be destroyed before removing
            let retries = 0
            while (retries < 10) {
              try {
                const sandboxInfo = await runnerAdapter.info(sandbox.id)
                if (sandboxInfo.state === RunnerSandboxState.DESTROYED) {
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
            await runnerAdapter.removeDestroyed(sandbox.id)
            sandbox.prevRunnerId = null

            const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
              id: sandbox.id,
            })
            sandboxToUpdate.prevRunnerId = null
            await this.sandboxRepository.save(sandboxToUpdate)
          } catch (e) {
            this.logger.error(`Failed to cleanup sandbox ${sandbox.id} on previous runner ${runner.id}:`)
          }
        }
        break
      }
      case RunnerSandboxState.ERROR: {
        await this.updateSandboxState(sandbox.id, SandboxState.ERROR)
        break
      }
    }

    return SYNC_AGAIN
  }

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
    const runnerAdapter = await this.runnerSandboxAdapterFactory.create(runner)

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
