/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { Runner } from '../entities/runner.entity'
import { SnapshotRegion } from '../entities/snapshot-region.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { JobStatus } from '../enums/job-status.enum'
import { JobType } from '../enums/job-type.enum'
import { Job } from '../entities/job.entity'
import { BackupState } from '../enums/backup-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { RunnerClass } from '../enums/runner-class'

/**
 * Service for handling entity state updates based on job completion (v2 runners only).
 * This service listens to job status changes and updates entity states accordingly.
 */
@Injectable()
export class JobStateHandlerService {
  private readonly logger = new Logger(JobStateHandlerService.name)

  constructor(
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    @InjectRepository(SnapshotRunner)
    private readonly snapshotRunnerRepository: Repository<SnapshotRunner>,
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    @InjectRepository(SnapshotRegion)
    private readonly snapshotRegionRepository: Repository<SnapshotRegion>,
  ) {}

  /**
   * Handle job completion and update entity state accordingly.
   * Called when a job status is updated to COMPLETED or FAILED.
   */
  async handleJobCompletion(job: Job): Promise<void> {
    if (job.status !== JobStatus.COMPLETED && job.status !== JobStatus.FAILED) {
      return
    }

    if (!job.resourceId) {
      return
    }

    switch (job.type) {
      case JobType.CREATE_SANDBOX:
        await this.handleCreateSandboxJobCompletion(job)
        break
      case JobType.START_SANDBOX:
        await this.handleStartSandboxJobCompletion(job)
        break
      case JobType.STOP_SANDBOX:
        await this.handleStopSandboxJobCompletion(job)
        break
      case JobType.DESTROY_SANDBOX:
        await this.handleDestroySandboxJobCompletion(job)
        break
      case JobType.PULL_SNAPSHOT:
        await this.handlePullSnapshotJobCompletion(job)
        break
      case JobType.BUILD_SNAPSHOT:
        await this.handleBuildSnapshotJobCompletion(job)
        break
      case JobType.REMOVE_SNAPSHOT:
        await this.handleRemoveSnapshotJobCompletion(job)
        break
      case JobType.CREATE_BACKUP:
        await this.handleCreateBackupJobCompletion(job)
        break
      case JobType.CREATE_SANDBOX_SNAPSHOT:
        await this.handleCreateSandboxSnapshotJobCompletion(job)
        break
      case JobType.FORK_SANDBOX:
        await this.handleForkSandboxJobCompletion(job)
        break
      case JobType.CLONE_SANDBOX:
        await this.handleCloneSandboxJobCompletion(job)
        break
      default:
        break
    }
  }

  private async handleCreateSandboxJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    if (!sandboxId) return

    try {
      const sandbox = await this.sandboxRepository.findOne({ where: { id: sandboxId } })
      if (!sandbox) {
        this.logger.warn(`Sandbox ${sandboxId} not found for CREATE_SANDBOX job ${job.id}`)
        return
      }

      if (sandbox.desiredState !== SandboxDesiredState.STARTED) {
        this.logger.error(
          `Sandbox ${sandboxId} is not in desired state STARTED for CREATE_SANDBOX job ${job.id}. Desired state: ${sandbox.desiredState}`,
        )
        return
      }

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(
          `CREATE_SANDBOX job ${job.id} completed successfully, marking sandbox ${sandboxId} as STARTED`,
        )
        sandbox.state = SandboxState.STARTED
        sandbox.errorReason = null
        const metadata = job.getResultMetadata()
        if (metadata?.daemonVersion && typeof metadata.daemonVersion === 'string') {
          sandbox.daemonVersion = metadata.daemonVersion
        }
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`CREATE_SANDBOX job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        sandbox.state = SandboxState.ERROR
        sandbox.errorReason = job.errorMessage || 'Failed to create sandbox'
      }

      await this.sandboxRepository.save(sandbox)
    } catch (error) {
      this.logger.error(`Error handling CREATE_SANDBOX job completion for sandbox ${sandboxId}:`, error)
    }
  }

  private async handleStartSandboxJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    if (!sandboxId) return

    try {
      const sandbox = await this.sandboxRepository.findOne({ where: { id: sandboxId } })
      if (!sandbox) {
        this.logger.warn(`Sandbox ${sandboxId} not found for START_SANDBOX job ${job.id}`)
        return
      }

      if (sandbox.desiredState !== SandboxDesiredState.STARTED) {
        this.logger.error(
          `Sandbox ${sandboxId} is not in desired state STARTED for START_SANDBOX job ${job.id}. Desired state: ${sandbox.desiredState}`,
        )
        return
      }

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(`START_SANDBOX job ${job.id} completed successfully, marking sandbox ${sandboxId} as STARTED`)
        sandbox.state = SandboxState.STARTED
        sandbox.errorReason = null
        const metadata = job.getResultMetadata()
        if (metadata?.daemonVersion && typeof metadata.daemonVersion === 'string') {
          sandbox.daemonVersion = metadata.daemonVersion
        }
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`START_SANDBOX job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        sandbox.state = SandboxState.ERROR
        sandbox.errorReason = job.errorMessage || 'Failed to start sandbox'
      }

      await this.sandboxRepository.save(sandbox)
    } catch (error) {
      this.logger.error(`Error handling START_SANDBOX job completion for sandbox ${sandboxId}:`, error)
    }
  }

  private async handleStopSandboxJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    if (!sandboxId) return

    try {
      const sandbox = await this.sandboxRepository.findOne({ where: { id: sandboxId } })
      if (!sandbox) {
        this.logger.warn(`Sandbox ${sandboxId} not found for STOP_SANDBOX job ${job.id}`)
        return
      }

      if (sandbox.desiredState !== SandboxDesiredState.STOPPED) {
        this.logger.error(
          `Sandbox ${sandboxId} is not in desired state STOPPED for STOP_SANDBOX job ${job.id}. Desired state: ${sandbox.desiredState}`,
        )
        return
      }

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(`STOP_SANDBOX job ${job.id} completed successfully, marking sandbox ${sandboxId} as STOPPED`)
        sandbox.state = SandboxState.STOPPED
        sandbox.errorReason = null
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`STOP_SANDBOX job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        sandbox.state = SandboxState.ERROR
        sandbox.errorReason = job.errorMessage || 'Failed to stop sandbox'
      }

      await this.sandboxRepository.save(sandbox)
    } catch (error) {
      this.logger.error(`Error handling STOP_SANDBOX job completion for sandbox ${sandboxId}:`, error)
    }
  }

  private async handleDestroySandboxJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    if (!sandboxId) return

    try {
      const sandbox = await this.sandboxRepository.findOne({ where: { id: sandboxId } })
      if (!sandbox) {
        this.logger.warn(`Sandbox ${sandboxId} not found for DESTROY_SANDBOX job ${job.id}`)
        return
      }
      if (sandbox.desiredState !== SandboxDesiredState.DESTROYED) {
        // Don't log anything because sandboxes can be destroyed on runners when archiving or moving to a new runner
        return
      }

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(
          `DESTROY_SANDBOX job ${job.id} completed successfully, marking sandbox ${sandboxId} as DESTROYED`,
        )
        sandbox.state = SandboxState.DESTROYED
        sandbox.errorReason = null
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`DESTROY_SANDBOX job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        sandbox.state = SandboxState.ERROR
        sandbox.errorReason = job.errorMessage || 'Failed to destroy sandbox'
      }

      await this.sandboxRepository.save(sandbox)
    } catch (error) {
      this.logger.error(`Error handling DESTROY_SANDBOX job completion for sandbox ${sandboxId}:`, error)
    }
  }

  private async handlePullSnapshotJobCompletion(job: Job): Promise<void> {
    const snapshotRef = job.resourceId
    const runnerId = job.runnerId
    if (!snapshotRef || !runnerId) return

    try {
      const snapshotRunner = await this.snapshotRunnerRepository.findOne({
        where: { snapshotRef, runnerId },
      })

      if (!snapshotRunner) {
        this.logger.warn(`SnapshotRunner not found for snapshot ${snapshotRef} on runner ${runnerId}`)
        return
      }

      if (job.status === JobStatus.COMPLETED) {
        this.logger.log(
          `PULL_SNAPSHOT job ${job.id} completed successfully, marking SnapshotRunner ${snapshotRunner.id} as READY`,
        )
        snapshotRunner.state = SnapshotRunnerState.READY
        snapshotRunner.errorReason = null

        // Check if this is the initial runner for a snapshot and update the snapshot state
        const snapshot = await this.snapshotRepository.findOne({
          where: { initialRunnerId: runnerId, ref: snapshotRef },
        })
        if (snapshot && snapshot.state === SnapshotState.PULLING) {
          this.logger.log(`Marking snapshot ${snapshot.id} as ACTIVE after initial pull completed`)
          snapshot.state = SnapshotState.ACTIVE
          snapshot.errorReason = null
          await this.snapshotRepository.save(snapshot)
        }
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`PULL_SNAPSHOT job ${job.id} failed for snapshot ${snapshotRef}: ${job.errorMessage}`)
        snapshotRunner.state = SnapshotRunnerState.ERROR
        snapshotRunner.errorReason = job.errorMessage || 'Failed to pull snapshot'

        // Check if this is the initial runner for a snapshot and update the snapshot state
        const snapshot = await this.snapshotRepository.findOne({
          where: { initialRunnerId: runnerId, ref: snapshotRef },
        })
        if (snapshot && snapshot.state === SnapshotState.PULLING) {
          this.logger.error(`Marking snapshot ${snapshot.id} as ERROR after initial pull failed`)
          snapshot.state = SnapshotState.ERROR
          snapshot.errorReason = job.errorMessage || 'Failed to pull snapshot on initial runner'
          await this.snapshotRepository.save(snapshot)
        }
      }

      await this.snapshotRunnerRepository.save(snapshotRunner)
    } catch (error) {
      this.logger.error(`Error handling PULL_SNAPSHOT job completion for snapshot ${snapshotRef}:`, error)
    }
  }

  private async handleBuildSnapshotJobCompletion(job: Job): Promise<void> {
    const snapshotRef = job.resourceId
    const runnerId = job.runnerId
    if (!snapshotRef || !runnerId) return

    try {
      // For BUILD_SNAPSHOT, find snapshot by buildInfo.snapshotRef
      const snapshot = await this.snapshotRepository
        .createQueryBuilder('snapshot')
        .leftJoinAndSelect('snapshot.buildInfo', 'buildInfo')
        .where('snapshot.initialRunnerId = :runnerId', { runnerId })
        .andWhere('buildInfo.snapshotRef = :snapshotRef', { snapshotRef })
        .getOne()

      if (!snapshot) {
        this.logger.warn(`Snapshot not found for build ref ${snapshotRef} on runner ${runnerId}`)
        return
      }

      // Update SnapshotRunner state
      const snapshotRunner = await this.snapshotRunnerRepository.findOne({
        where: { snapshotRef, runnerId },
      })

      if (job.status === JobStatus.COMPLETED) {
        this.logger.log(`BUILD_SNAPSHOT job ${job.id} completed successfully for snapshot ${snapshot.id}`)

        if (snapshot.state === SnapshotState.BUILDING) {
          snapshot.state = SnapshotState.ACTIVE
          snapshot.errorReason = null
          await this.snapshotRepository.save(snapshot)
          this.logger.log(`Marked snapshot ${snapshot.id} as ACTIVE after build completed`)
        }

        if (snapshotRunner) {
          snapshotRunner.state = SnapshotRunnerState.READY
          snapshotRunner.errorReason = null
          await this.snapshotRunnerRepository.save(snapshotRunner)
        }
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`BUILD_SNAPSHOT job ${job.id} failed for snapshot ${snapshot.id}: ${job.errorMessage}`)

        if (snapshot.state === SnapshotState.BUILDING) {
          snapshot.state = SnapshotState.ERROR
          snapshot.errorReason = job.errorMessage || 'Failed to build snapshot'
          await this.snapshotRepository.save(snapshot)
        }

        if (snapshotRunner) {
          snapshotRunner.state = SnapshotRunnerState.ERROR
          snapshotRunner.errorReason = job.errorMessage || 'Failed to build snapshot'
          await this.snapshotRunnerRepository.save(snapshotRunner)
        }
      }
    } catch (error) {
      this.logger.error(`Error handling BUILD_SNAPSHOT job completion for snapshot ref ${snapshotRef}:`, error)
    }
  }

  private async handleRemoveSnapshotJobCompletion(job: Job): Promise<void> {
    const snapshotRef = job.resourceId
    const runnerId = job.runnerId
    if (!snapshotRef || !runnerId) return

    try {
      if (job.status === JobStatus.COMPLETED) {
        this.logger.log(
          `REMOVE_SNAPSHOT job ${job.id} completed successfully for snapshot ${snapshotRef} on runner ${runnerId}`,
        )
        const affected = await this.snapshotRunnerRepository.delete({ snapshotRef, runnerId })
        if (affected.affected && affected.affected > 0) {
          this.logger.log(
            `Removed ${affected.affected} snapshot runners for snapshot ${snapshotRef} on runner ${runnerId}`,
          )
        }
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(
          `REMOVE_SNAPSHOT job ${job.id} failed for snapshot ${snapshotRef} on runner ${runnerId}: ${job.errorMessage}`,
        )
      }
    } catch (error) {
      this.logger.error(`Error handling REMOVE_SNAPSHOT job completion for snapshot ${snapshotRef}:`, error)
    }
  }

  private async handleCreateBackupJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    if (!sandboxId) return

    try {
      const sandbox = await this.sandboxRepository.findOne({ where: { id: sandboxId } })
      if (!sandbox) {
        this.logger.warn(`Sandbox ${sandboxId} not found for CREATE_BACKUP job ${job.id}`)
        return
      }

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(
          `CREATE_BACKUP job ${job.id} completed successfully, marking sandbox ${sandboxId} as BACKUP_COMPLETED`,
        )
        sandbox.setBackupState(BackupState.COMPLETED)
        await this.sandboxRepository.save(sandbox)
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`CREATE_BACKUP job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        sandbox.setBackupState(BackupState.ERROR, undefined, undefined, job.errorMessage)
        await this.sandboxRepository.save(sandbox)
      }
    } catch (error) {
      this.logger.error(`Error handling CREATE_BACKUP job completion for sandbox ${sandboxId}:`, error)
    }
  }

  private async handleCreateSandboxSnapshotJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    const runnerId = job.runnerId
    if (!sandboxId || !runnerId) return

    try {
      // Parse job payload to get snapshot name
      const payload = job.payload ? JSON.parse(job.payload) : {}
      const snapshotName = payload.name

      if (!snapshotName) {
        this.logger.error(`CREATE_SANDBOX_SNAPSHOT job ${job.id} has no snapshot name in payload`)
        return
      }

      // Get sandbox to retrieve organization and other info
      const sandbox = await this.sandboxRepository.findOne({ where: { id: sandboxId } })
      if (!sandbox) {
        this.logger.warn(`Sandbox ${sandboxId} not found for CREATE_SANDBOX_SNAPSHOT job ${job.id}`)
        return
      }

      if (job.status === JobStatus.COMPLETED) {
        this.logger.log(`CREATE_SANDBOX_SNAPSHOT job ${job.id} completed successfully for sandbox ${sandboxId}`)

        // Get result metadata from job
        const metadata = job.getResultMetadata()
        const snapshotPath = metadata?.snapshotPath || metadata?.SnapshotPath
        const sizeBytes = metadata?.sizeBytes || metadata?.SizeBytes

        // Get runner to determine runner class
        const runner = await this.runnerRepository.findOne({ where: { id: runnerId } })
        if (!runner) {
          this.logger.warn(`Runner ${runnerId} not found for CREATE_SANDBOX_SNAPSHOT job ${job.id}`)
          return
        }

        // Create Snapshot entity with resources from source sandbox
        const snapshot = new Snapshot()
        snapshot.organizationId = sandbox.organizationId
        snapshot.name = snapshotName
        snapshot.imageName = snapshotPath || snapshotName
        snapshot.ref = snapshotPath || snapshotName
        snapshot.runnerClass = runner.class
        snapshot.state = SnapshotState.ACTIVE // Mark as active since it's ready to use
        // Copy resource specifications from source sandbox
        snapshot.cpu = sandbox.cpu
        snapshot.mem = sandbox.mem
        snapshot.disk = sandbox.disk
        snapshot.gpu = sandbox.gpu

        await this.snapshotRepository.save(snapshot)
        this.logger.log(`Created snapshot entity ${snapshot.id} with name ${snapshotName}`)

        // Create SnapshotRegion entry - associate snapshot with sandbox's region
        if (sandbox.region) {
          const snapshotRegion = new SnapshotRegion()
          snapshotRegion.snapshotId = snapshot.id
          snapshotRegion.regionId = sandbox.region
          await this.snapshotRegionRepository.save(snapshotRegion)
          this.logger.log(`Created SnapshotRegion for snapshot ${snapshotName} in region ${sandbox.region}`)
        }

        // Create SnapshotRunner entry for the runner that created the snapshot
        const snapshotRunner = new SnapshotRunner()
        snapshotRunner.snapshotRef = snapshot.ref || snapshotName
        snapshotRunner.runnerId = runnerId
        snapshotRunner.state = SnapshotRunnerState.READY // 'ready' in DB enum
        await this.snapshotRunnerRepository.save(snapshotRunner)
        this.logger.log(`Created SnapshotRunner for snapshot ${snapshotName} on runner ${runnerId}`)
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`CREATE_SANDBOX_SNAPSHOT job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        // TODO: Notify user that snapshot creation failed
      }

      // in all cases, set the sandbox state to the desired state
      if (job.status === JobStatus.COMPLETED || job.status === JobStatus.FAILED) {
        switch (sandbox.desiredState) {
          case SandboxDesiredState.STARTED:
            sandbox.state = SandboxState.STARTED
            break
          case SandboxDesiredState.STOPPED:
            sandbox.state = SandboxState.STOPPED
            break
          default:
            console.error(
              `Unknown desired state ${sandbox.desiredState} for sandbox ${sandboxId} after CREATE_SANDBOX_SNAPSHOT job completed`,
            )
            sandbox.state = SandboxState.ERROR
            break
        }
        sandbox.pending = false
        await this.sandboxRepository.save(sandbox)
      }
    } catch (error) {
      this.logger.error(`Error handling CREATE_SANDBOX_SNAPSHOT job completion for sandbox ${sandboxId}:`, error)
    }
  }

  private async handleForkSandboxJobCompletion(job: Job): Promise<void> {
    const forkedSandboxId = job.resourceId
    if (!forkedSandboxId) return

    try {
      // Parse job payload to get source sandbox ID
      const payload = job.payload ? JSON.parse(job.payload) : {}
      const sourceSandboxId = payload.sourceSandboxId

      // Get the source sandbox to clear its backupState (do this first so we always clear it)
      let sourceSandbox: Sandbox | null = null
      if (sourceSandboxId) {
        sourceSandbox = await this.sandboxRepository.findOne({ where: { id: sourceSandboxId } })

        switch (sourceSandbox?.desiredState) {
          case SandboxDesiredState.STARTED:
            sourceSandbox.state = SandboxState.STARTED
            sourceSandbox.pending = false
            break
          case SandboxDesiredState.STOPPED:
            sourceSandbox.state = SandboxState.STOPPED
            sourceSandbox.pending = false
            break
          default:
            console.error(
              `Unknown desired state ${sourceSandbox?.desiredState} for source sandbox ${sourceSandboxId} after FORK_SANDBOX job completed`,
            )
            sourceSandbox.state = SandboxState.ERROR
            sourceSandbox.pending = false
            break
        }

        await this.sandboxRepository.save(sourceSandbox)
      } else {
        this.logger.warn(`Source sandbox ${sourceSandboxId} not found for FORK_SANDBOX job ${job.id}`)
        return
      }

      // Get the forked sandbox
      const forkedSandbox = await this.sandboxRepository.findOne({ where: { id: forkedSandboxId } })
      if (!forkedSandbox) {
        this.logger.warn(`Forked sandbox ${forkedSandboxId} not found for FORK_SANDBOX job ${job.id}`)
        return
      }

      if (forkedSandbox.desiredState !== SandboxDesiredState.STARTED) {
        this.logger.error(
          `Forked sandbox ${forkedSandboxId} is not in desired state STARTED for FORK_SANDBOX job ${job.id}. Desired state: ${forkedSandbox.desiredState}`,
        )
        return
      }

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(
          `FORK_SANDBOX job ${job.id} completed successfully, marking forked sandbox ${forkedSandboxId} as STARTED`,
        )
        forkedSandbox.state = SandboxState.STARTED
        forkedSandbox.errorReason = null
        const metadata = job.getResultMetadata()
        if (metadata?.daemonVersion && typeof metadata.daemonVersion === 'string') {
          forkedSandbox.daemonVersion = metadata.daemonVersion
        }
        await this.sandboxRepository.save(forkedSandbox)
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`FORK_SANDBOX job ${job.id} failed for sandbox ${forkedSandboxId}: ${job.errorMessage}`)
        forkedSandbox.state = SandboxState.ERROR
        forkedSandbox.errorReason = job.errorMessage || 'Failed to fork sandbox'
        await this.sandboxRepository.save(forkedSandbox)
      }
    } catch (error) {
      this.logger.error(`Error handling FORK_SANDBOX job completion for sandbox ${forkedSandboxId}:`, error)
    }
  }

  private async handleCloneSandboxJobCompletion(job: Job): Promise<void> {
    const clonedSandboxId = job.resourceId
    if (!clonedSandboxId) return

    try {
      // Parse job payload to get source sandbox ID
      const payload = job.payload ? JSON.parse(job.payload) : {}
      const sourceSandboxId = payload.sourceSandboxId

      // Get the source sandbox to clear its state (do this first so we always clear it)
      let sourceSandbox: Sandbox | null = null
      if (sourceSandboxId) {
        sourceSandbox = await this.sandboxRepository.findOne({ where: { id: sourceSandboxId } })

        if (sourceSandbox) {
          switch (sourceSandbox.desiredState) {
            case SandboxDesiredState.STARTED:
              sourceSandbox.state = SandboxState.STARTED
              sourceSandbox.pending = false
              break
            case SandboxDesiredState.STOPPED:
              sourceSandbox.state = SandboxState.STOPPED
              sourceSandbox.pending = false
              break
            default:
              console.error(
                `Unknown desired state ${sourceSandbox.desiredState} for source sandbox ${sourceSandboxId} after CLONE_SANDBOX job completed`,
              )
              sourceSandbox.state = SandboxState.ERROR
              sourceSandbox.pending = false
              break
          }

          await this.sandboxRepository.save(sourceSandbox)
        }
      } else {
        this.logger.warn(`Source sandbox ${sourceSandboxId} not found for CLONE_SANDBOX job ${job.id}`)
        return
      }

      // Get the cloned sandbox
      const clonedSandbox = await this.sandboxRepository.findOne({ where: { id: clonedSandboxId } })
      if (!clonedSandbox) {
        this.logger.warn(`Cloned sandbox ${clonedSandboxId} not found for CLONE_SANDBOX job ${job.id}`)
        return
      }

      if (clonedSandbox.desiredState !== SandboxDesiredState.STARTED) {
        this.logger.error(
          `Cloned sandbox ${clonedSandboxId} is not in desired state STARTED for CLONE_SANDBOX job ${job.id}. Desired state: ${clonedSandbox.desiredState}`,
        )
        return
      }

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(
          `CLONE_SANDBOX job ${job.id} completed successfully, marking cloned sandbox ${clonedSandboxId} as STARTED`,
        )
        clonedSandbox.state = SandboxState.STARTED
        clonedSandbox.pending = false
        clonedSandbox.errorReason = null
        const metadata = job.getResultMetadata()
        if (metadata?.daemonVersion && typeof metadata.daemonVersion === 'string') {
          clonedSandbox.daemonVersion = metadata.daemonVersion
        }
        await this.sandboxRepository.save(clonedSandbox)
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`CLONE_SANDBOX job ${job.id} failed for sandbox ${clonedSandboxId}: ${job.errorMessage}`)
        clonedSandbox.state = SandboxState.ERROR
        clonedSandbox.pending = false
        clonedSandbox.errorReason = job.errorMessage || 'Failed to clone sandbox'
        await this.sandboxRepository.save(clonedSandbox)
      }
    } catch (error) {
      this.logger.error(`Error handling CLONE_SANDBOX job completion for sandbox ${clonedSandboxId}:`, error)
    }
  }
}
