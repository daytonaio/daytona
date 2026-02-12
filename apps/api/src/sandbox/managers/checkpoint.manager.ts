/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { Checkpoint } from '../entities/checkpoint.entity'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxDestroyedEvent } from '../events/sandbox-destroyed.event'
import { RunnerAdapterFactory } from '../runner-adapter/runnerAdapter'
import { Runner } from '../entities/runner.entity'
import { TrackJobExecution } from '../../common/decorators/track-job-execution.decorator'

/**
 * CheckpointManager handles checkpoint lifecycle management,
 * specifically cleanup of checkpoint images from runners when a sandbox is destroyed.
 * The checkpoint entities themselves are cascade-deleted by the database FK constraint.
 */
@Injectable()
export class CheckpointManager {
  private readonly logger = new Logger(CheckpointManager.name)

  constructor(
    @InjectRepository(Checkpoint)
    private readonly checkpointRepository: Repository<Checkpoint>,
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    private readonly runnerAdapterFactory: RunnerAdapterFactory,
  ) {}

  /**
   * When a sandbox is destroyed, clean up checkpoint images from runners.
   * The checkpoint entities are cascade-deleted by the DB, but images on runners
   * need to be explicitly removed.
   */
  @OnEvent(SandboxEvents.DESTROYED)
  @TrackJobExecution()
  async handleSandboxDestroyed(event: SandboxDestroyedEvent): Promise<void> {
    const sandbox = event.sandbox

    try {
      // Get all checkpoints for this sandbox before cascade delete removes them
      const checkpoints = await this.checkpointRepository.find({
        where: { sandboxId: sandbox.id },
        relations: ['runners'],
      })

      if (checkpoints.length === 0) {
        return
      }

      this.logger.log(`Cleaning up ${checkpoints.length} checkpoint image(s) for destroyed sandbox ${sandbox.id}`)

      // Remove checkpoint images from each runner that has them
      for (const checkpoint of checkpoints) {
        if (!checkpoint.ref) continue

        for (const cr of checkpoint.runners || []) {
          try {
            const runner = await this.runnerRepository.findOne({
              where: { id: cr.runnerId },
            })
            if (!runner) continue

            const runnerAdapter = await this.runnerAdapterFactory.create(runner)
            await runnerAdapter.removeSnapshot(checkpoint.ref)

            this.logger.debug(`Removed checkpoint image ${checkpoint.ref} from runner ${cr.runnerId}`)
          } catch (error) {
            this.logger.warn(
              `Failed to remove checkpoint image ${checkpoint.ref} from runner ${cr.runnerId}: ${error.message}`,
            )
            // Continue with other runners - best effort cleanup
          }
        }
      }
    } catch (error) {
      this.logger.error(`Error cleaning up checkpoints for sandbox ${sandbox.id}: ${error.message}`)
    }
  }
}
