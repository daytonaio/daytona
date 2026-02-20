/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException, Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
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
import { sanitizeSandboxError } from '../utils/sanitize-error.util'
import { OrganizationUsageService } from '../../organization/services/organization-usage.service'
import { SandboxRepository } from '../repositories/sandbox.repository'
import { Sandbox } from '../entities/sandbox.entity'

/**
 * Service for handling entity state updates based on job completion (v2 runners only).
 * This service listens to job status changes and updates entity states accordingly.
 */
@Injectable()
export class JobStateHandlerService {
  private readonly logger = new Logger(JobStateHandlerService.name)

  constructor(
    private readonly sandboxRepository: SandboxRepository,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    @InjectRepository(SnapshotRunner)
    private readonly snapshotRunnerRepository: Repository<SnapshotRunner>,
    private readonly organizationUsageService: OrganizationUsageService,
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
      case JobType.RESIZE_SANDBOX:
        await this.handleResizeSandboxJobCompletion(job)
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
      case JobType.RECOVER_SANDBOX:
        await this.handleRecoverSandboxJobCompletion(job)
        break
      default:
        break
    }
  }

  private async handleCreateSandboxJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    if (!sandboxId) return

    try {
      const updateData: Partial<Sandbox> = {}

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(
          `CREATE_SANDBOX job ${job.id} completed successfully, marking sandbox ${sandboxId} as STARTED`,
        )
        updateData.state = SandboxState.STARTED
        updateData.errorReason = null
        const metadata = job.getResultMetadata()
        if (metadata?.daemonVersion && typeof metadata.daemonVersion === 'string') {
          updateData.daemonVersion = metadata.daemonVersion
        }
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`CREATE_SANDBOX job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        updateData.state = SandboxState.ERROR
        const { recoverable, errorReason } = sanitizeSandboxError(job.errorMessage)
        updateData.errorReason = errorReason || 'Failed to create sandbox'
        updateData.recoverable = recoverable
      }

      await this.sandboxRepository.updateWhere(sandboxId, {
        updateData,
        whereCondition: { desiredState: SandboxDesiredState.STARTED },
      })
    } catch (error) {
      if (error instanceof ConflictException) {
        this.logger.warn(
          `Sandbox ${sandboxId} not found or not in desired state STARTED for CREATE_SANDBOX job ${job.id}`,
        )
        return
      }
      this.logger.error(`Error handling CREATE_SANDBOX job completion for sandbox ${sandboxId}:`, error)
    }
  }

  private async handleStartSandboxJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    if (!sandboxId) return

    try {
      const updateData: Partial<Sandbox> = {}

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(`START_SANDBOX job ${job.id} completed successfully, marking sandbox ${sandboxId} as STARTED`)
        updateData.state = SandboxState.STARTED
        updateData.errorReason = null
        const metadata = job.getResultMetadata()
        if (metadata?.daemonVersion && typeof metadata.daemonVersion === 'string') {
          updateData.daemonVersion = metadata.daemonVersion
        }
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`START_SANDBOX job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        updateData.state = SandboxState.ERROR
        const { recoverable, errorReason } = sanitizeSandboxError(job.errorMessage)
        updateData.errorReason = errorReason || 'Failed to start sandbox'
        updateData.recoverable = recoverable
      }

      await this.sandboxRepository.updateWhere(sandboxId, {
        updateData,
        whereCondition: { desiredState: SandboxDesiredState.STARTED },
      })
    } catch (error) {
      if (error instanceof ConflictException) {
        this.logger.warn(
          `Sandbox ${sandboxId} not found or not in desired state STARTED for START_SANDBOX job ${job.id}`,
        )
        return
      }
      this.logger.error(`Error handling START_SANDBOX job completion for sandbox ${sandboxId}:`, error)
    }
  }

  private async handleStopSandboxJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    if (!sandboxId) return

    try {
      const updateData: Partial<Sandbox> = {}

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(`STOP_SANDBOX job ${job.id} completed successfully, marking sandbox ${sandboxId} as STOPPED`)
        updateData.state = SandboxState.STOPPED
        updateData.errorReason = null
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`STOP_SANDBOX job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        updateData.state = SandboxState.ERROR
        const { recoverable, errorReason } = sanitizeSandboxError(job.errorMessage)
        updateData.errorReason = errorReason || 'Failed to stop sandbox'
        updateData.recoverable = recoverable
      }

      await this.sandboxRepository.updateWhere(sandboxId, {
        updateData,
        whereCondition: { desiredState: SandboxDesiredState.STOPPED },
      })
    } catch (error) {
      if (error instanceof ConflictException) {
        this.logger.warn(
          `Sandbox ${sandboxId} not found or not in desired state STOPPED for STOP_SANDBOX job ${job.id}`,
        )
        return
      }
      this.logger.error(`Error handling STOP_SANDBOX job completion for sandbox ${sandboxId}:`, error)
    }
  }

  private async handleDestroySandboxJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    if (!sandboxId) return

    try {
      const updateData: Partial<Sandbox> = {}

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(
          `DESTROY_SANDBOX job ${job.id} completed successfully, marking sandbox ${sandboxId} as DESTROYED`,
        )
        updateData.state = SandboxState.DESTROYED
        updateData.errorReason = null
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`DESTROY_SANDBOX job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        updateData.state = SandboxState.ERROR
        const { recoverable, errorReason } = sanitizeSandboxError(job.errorMessage)
        updateData.errorReason = errorReason || 'Failed to destroy sandbox'
        updateData.recoverable = recoverable
      }

      await this.sandboxRepository.updateWhere(sandboxId, {
        updateData,
        whereCondition: { desiredState: SandboxDesiredState.DESTROYED },
      })
    } catch (error) {
      if (error instanceof ConflictException) {
        // Don't log — sandboxes can be destroyed on runners when archiving or moving to a new runner
        return
      }
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
        this.logger.debug(
          `PULL_SNAPSHOT job ${job.id} completed successfully, marking SnapshotRunner ${snapshotRunner.id} as READY`,
        )
        snapshotRunner.state = SnapshotRunnerState.READY
        snapshotRunner.errorReason = null

        // Check if this is the initial runner for a snapshot and update the snapshot state
        const snapshot = await this.snapshotRepository.findOne({
          where: { initialRunnerId: runnerId, ref: snapshotRef },
        })
        if (snapshot && (snapshot.state === SnapshotState.PULLING || snapshot.state === SnapshotState.BUILDING)) {
          this.logger.debug(`Marking snapshot ${snapshot.id} as ACTIVE after initial pull completed`)
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
        this.logger.debug(`BUILD_SNAPSHOT job ${job.id} completed successfully for snapshot ref ${snapshotRef}`)

        if (snapshot?.state === SnapshotState.BUILDING) {
          snapshot.state = SnapshotState.ACTIVE
          snapshot.errorReason = null
          await this.snapshotRepository.save(snapshot)
          this.logger.debug(`Marked snapshot ${snapshot.id} as ACTIVE after build completed`)
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
        this.logger.debug(
          `REMOVE_SNAPSHOT job ${job.id} completed successfully for snapshot ${snapshotRef} on runner ${runnerId}`,
        )
        const affected = await this.snapshotRunnerRepository.delete({ snapshotRef, runnerId })
        if (affected.affected && affected.affected > 0) {
          this.logger.debug(
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

      const updateData: Partial<Sandbox> = {}

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(
          `CREATE_BACKUP job ${job.id} completed successfully, marking sandbox ${sandboxId} as BACKUP_COMPLETED`,
        )
        Object.assign(updateData, Sandbox.getBackupStateUpdate(sandbox, BackupState.COMPLETED))
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`CREATE_BACKUP job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        Object.assign(
          updateData,
          Sandbox.getBackupStateUpdate(sandbox, BackupState.ERROR, undefined, undefined, job.errorMessage),
        )
      }

      await this.sandboxRepository.update(sandboxId, { updateData, entity: sandbox })
    } catch (error) {
      this.logger.error(`Error handling CREATE_BACKUP job completion for sandbox ${sandboxId}:`, error)
    }
  }

  private async handleRecoverSandboxJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    if (!sandboxId) return

    try {
      const updateData: Partial<Sandbox> = {}

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(
          `RECOVER_SANDBOX job ${job.id} completed successfully, marking sandbox ${sandboxId} as STARTED`,
        )
        updateData.state = SandboxState.STARTED
        updateData.errorReason = null
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`RECOVER_SANDBOX job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)
        updateData.state = SandboxState.ERROR
        updateData.errorReason = job.errorMessage || 'Failed to recover sandbox'
      }

      await this.sandboxRepository.updateWhere(sandboxId, {
        updateData,
        whereCondition: { desiredState: SandboxDesiredState.STARTED },
      })
    } catch (error) {
      if (error instanceof ConflictException) {
        this.logger.warn(
          `Sandbox ${sandboxId} not found or not in desired state STARTED for RECOVER_SANDBOX job ${job.id}`,
        )
        return
      }
      this.logger.error(`Error handling RECOVER_SANDBOX job completion for sandbox ${sandboxId}:`, error)
    }
  }

  private async handleResizeSandboxJobCompletion(job: Job): Promise<void> {
    const sandboxId = job.resourceId
    if (!sandboxId) return

    try {
      const sandbox = await this.sandboxRepository.findOne({ where: { id: sandboxId } })
      if (!sandbox) {
        this.logger.warn(`Sandbox ${sandboxId} not found for RESIZE_SANDBOX job ${job.id}`)
        return
      }

      if (sandbox.state !== SandboxState.RESIZING) {
        this.logger.warn(
          `Sandbox ${sandboxId} is not in RESIZING state for RESIZE_SANDBOX job ${job.id}. State: ${sandbox.state}`,
        )
        return
      }

      // Determine the previous state (STARTED or STOPPED based on desiredState)
      const previousState =
        sandbox.desiredState === SandboxDesiredState.STARTED
          ? SandboxState.STARTED
          : sandbox.desiredState === SandboxDesiredState.STOPPED
            ? SandboxState.STOPPED
            : null

      if (!previousState) {
        this.logger.error(
          `Sandbox ${sandboxId} has unexpected desiredState ${sandbox.desiredState} for RESIZE_SANDBOX job ${job.id}`,
        )
        return
      }

      // Calculate deltas before updating sandbox
      const payload = job.payload as { cpu?: number; memory?: number; disk?: number }

      // For cold resize (previousState === STOPPED), cpu/memory don't affect org quota.
      const isHotResize = previousState === SandboxState.STARTED
      const cpuDeltaForQuota = isHotResize ? (payload.cpu ?? sandbox.cpu) - sandbox.cpu : 0
      const memDeltaForQuota = isHotResize ? (payload.memory ?? sandbox.mem) - sandbox.mem : 0
      const diskDeltaForQuota = (payload.disk ?? sandbox.disk) - sandbox.disk // Disk only increases

      const updateData: Partial<Sandbox> = {}

      if (job.status === JobStatus.COMPLETED) {
        this.logger.debug(`RESIZE_SANDBOX job ${job.id} completed successfully for sandbox ${sandboxId}`)

        // Update sandbox resources
        updateData.cpu = payload.cpu ?? sandbox.cpu
        updateData.mem = payload.memory ?? sandbox.mem
        updateData.disk = payload.disk ?? sandbox.disk
        updateData.state = previousState

        // Apply usage change (handles both positive and negative deltas)
        await this.organizationUsageService.applyResizeUsageChange(
          sandbox.organizationId,
          sandbox.region,
          cpuDeltaForQuota,
          memDeltaForQuota,
          diskDeltaForQuota,
        )
        return
      } else if (job.status === JobStatus.FAILED) {
        this.logger.error(`RESIZE_SANDBOX job ${job.id} failed for sandbox ${sandboxId}: ${job.errorMessage}`)

        // Rollback pending usage (all deltas were tracked, including negative)
        await this.organizationUsageService.decrementPendingSandboxUsage(
          sandbox.organizationId,
          sandbox.region,
          cpuDeltaForQuota !== 0 ? cpuDeltaForQuota : undefined,
          memDeltaForQuota !== 0 ? memDeltaForQuota : undefined,
          diskDeltaForQuota !== 0 ? diskDeltaForQuota : undefined,
        )

        updateData.state = previousState
      }

      await this.sandboxRepository.update(sandboxId, { updateData, entity: sandbox })
    } catch (error) {
      this.logger.error(`Error handling RESIZE_SANDBOX job completion for sandbox ${sandboxId}:`, error)
    }
  }
}
