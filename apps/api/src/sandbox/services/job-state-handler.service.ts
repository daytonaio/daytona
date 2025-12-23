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
import { SandboxState } from '../enums/sandbox-state.enum'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { JobStatus } from '../enums/job-status.enum'
import { JobType } from '../enums/job-type.enum'
import { Job } from '../entities/job.entity'
import { BackupState } from '../enums/backup-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'

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
        if (snapshot && (snapshot.state === SnapshotState.PULLING || snapshot.state === SnapshotState.BUILDING)) {
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

      // Update SnapshotRunner state
      const snapshotRunner = await this.snapshotRunnerRepository.findOne({
        where: { snapshotRef, runnerId },
      })

      if (job.status === JobStatus.COMPLETED) {
        this.logger.log(`BUILD_SNAPSHOT job ${job.id} completed successfully for snapshot ref ${snapshotRef}`)

        if (snapshot?.state === SnapshotState.BUILDING) {
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
        this.logger.error(`BUILD_SNAPSHOT job ${job.id} failed for snapshot ref ${snapshotRef}: ${job.errorMessage}`)

        if (snapshot?.state === SnapshotState.BUILDING) {
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
}
