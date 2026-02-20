/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { Repository } from 'typeorm'
import { Checkpoint } from '../entities/checkpoint.entity'
import { CheckpointState } from '../enums/checkpoint-state.enum'
import { Sandbox } from '../entities/sandbox.entity'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { Runner } from '../entities/runner.entity'
import { RunnerAdapterFactory, RunnerSnapshotInfo } from '../runner-adapter/runnerAdapter'
import { CheckpointService } from '../services/checkpoint.service'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'

const CHECKPOINT_CREATION_TIMEOUT_MS = 30 * 60 * 1000 // 30 minutes

@Injectable()
export class CheckpointManager {
  private readonly logger = new Logger(CheckpointManager.name)

  constructor(
    @InjectRepository(Checkpoint)
    private readonly checkpointRepository: Repository<Checkpoint>,
    @InjectRepository(SnapshotRunner)
    private readonly snapshotRunnerRepository: Repository<SnapshotRunner>,
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    private readonly runnerAdapterFactory: RunnerAdapterFactory,
    private readonly checkpointService: CheckpointService,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly dockerRegistryService: DockerRegistryService,
  ) {}

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'poll-creating-checkpoints' })
  async pollCreatingCheckpoints(): Promise<void> {
    const lockKey = 'poll-creating-checkpoints-lock'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    try {
      const creatingCheckpoints = await this.checkpointService.getCreatingCheckpoints()

      for (const checkpoint of creatingCheckpoints) {
        try {
          await this.pollCheckpointCreation(checkpoint)
        } catch (error) {
          this.logger.error(`Error polling checkpoint ${checkpoint.id}: ${error.message}`)
        }
      }
    } finally {
      await this.redisLockProvider.unlock('poll-creating-checkpoints-lock')
    }
  }

  private async pollCheckpointCreation(checkpoint: Checkpoint): Promise<void> {
    const elapsed = Date.now() - checkpoint.createdAt.getTime()
    if (elapsed > CHECKPOINT_CREATION_TIMEOUT_MS) {
      await this.checkpointService.markError(checkpoint.id, 'Checkpoint creation timed out')
      await this.releaseSandboxPending(checkpoint.originSandboxId)
      return
    }

    if (!checkpoint.initialRunnerId) {
      await this.checkpointService.markError(checkpoint.id, 'No runner assigned for checkpoint creation')
      await this.releaseSandboxPending(checkpoint.originSandboxId)
      return
    }

    const runner = await this.runnerRepository.findOne({
      where: { id: checkpoint.initialRunnerId },
    })
    if (!runner) {
      await this.checkpointService.markError(checkpoint.id, 'Runner not found')
      await this.releaseSandboxPending(checkpoint.originSandboxId)
      return
    }

    // For V2 runners, job-state-handler handles completion - skip polling
    if (runner.apiVersion === '2') {
      return
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    let info: RunnerSnapshotInfo
    try {
      const internalRegistry = await this.dockerRegistryService.getAvailableInternalRegistry(runner.region)
      const expectedRef = this.buildExpectedCheckpointRef(checkpoint, internalRegistry)
      if (!expectedRef) {
        return
      }

      info = await runnerAdapter.getCheckpointInfo(expectedRef)
    } catch {
      // Not ready yet - still creating
      return
    }

    const internalRegistry = await this.dockerRegistryService.getAvailableInternalRegistry(runner.region)
    const ref = this.buildExpectedCheckpointRef(checkpoint, internalRegistry)

    await this.checkpointService.markActive(
      checkpoint.id,
      ref,
      checkpoint.initialRunnerId,
      info.sizeGB ? info.sizeGB * 1024 * 1024 * 1024 : undefined,
    )
    await this.releaseSandboxPending(checkpoint.originSandboxId)

    this.logger.log(`Checkpoint ${checkpoint.id} creation completed on runner ${runner.id}`)
  }

  private buildExpectedCheckpointRef(
    checkpoint: Checkpoint,
    registry?: { url: string; project?: string },
  ): string | undefined {
    if (!registry) return undefined

    if (registry.project) {
      return `${registry.url.replace(/^(https?:\/\/)/, '')}/${registry.project}/${checkpoint.organizationId}/${checkpoint.name}:latest`
    }
    return `${registry.url.replace(/^(https?:\/\/)/, '')}/${checkpoint.organizationId}/${checkpoint.name}:latest`
  }

  private async releaseSandboxPending(sandboxId: string): Promise<void> {
    try {
      const sandbox = await this.sandboxRepository.findOne({ where: { id: sandboxId } })
      if (sandbox && sandbox.pending) {
        sandbox.pending = false
        await this.sandboxRepository.save(sandbox)
      }
    } catch (error) {
      this.logger.warn(`Failed to release pending state for sandbox ${sandboxId}: ${error.message}`)
    }
  }

  async cleanupCheckpointImages(checkpoint: Checkpoint): Promise<void> {
    if (!checkpoint.ref) return

    const snapshotRunners = await this.snapshotRunnerRepository.find({
      where: { snapshotRef: checkpoint.ref },
    })

    for (const sr of snapshotRunners) {
      try {
        const runner = await this.runnerRepository.findOne({
          where: { id: sr.runnerId },
        })
        if (!runner) continue

        const runnerAdapter = await this.runnerAdapterFactory.create(runner)
        await runnerAdapter.removeCheckpoint(checkpoint.ref)
        await this.snapshotRunnerRepository.remove(sr)

        this.logger.debug(`Removed checkpoint image ${checkpoint.ref} from runner ${sr.runnerId}`)
      } catch (error) {
        this.logger.warn(
          `Failed to remove checkpoint image ${checkpoint.ref} from runner ${sr.runnerId}: ${error.message}`,
        )
      }
    }
  }

  @Cron(CronExpression.EVERY_30_SECONDS, { name: 'cleanup-removing-checkpoints' })
  async cleanupRemoving(): Promise<void> {
    const lockKey = 'cleanup-removing-checkpoints-lock'
    if (!(await this.redisLockProvider.lock(lockKey, 30))) {
      return
    }

    try {
      const checkpoints = await this.checkpointRepository.find({
        where: { state: CheckpointState.REMOVING },
      })

      for (const checkpoint of checkpoints) {
        try {
          await this.cleanupCheckpointImages(checkpoint)
          await this.checkpointRepository.remove(checkpoint)
          this.logger.log(`Cleaned up REMOVING checkpoint ${checkpoint.id}`)
        } catch (error) {
          this.logger.error(`Error cleaning up checkpoint ${checkpoint.id}: ${error.message}`)
        }
      }
    } finally {
      await this.redisLockProvider.unlock('cleanup-removing-checkpoints-lock')
    }
  }
}
