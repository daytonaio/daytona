/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException, BadRequestException, ConflictException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { Checkpoint } from '../entities/checkpoint.entity'
import { CheckpointRunner } from '../entities/checkpoint-runner.entity'
import { CheckpointState } from '../enums/checkpoint-state.enum'
import { CheckpointRunnerState } from '../enums/checkpoint-runner-state.enum'

@Injectable()
export class CheckpointService {
  private readonly logger = new Logger(CheckpointService.name)

  constructor(
    @InjectRepository(Checkpoint)
    private readonly checkpointRepository: Repository<Checkpoint>,
    @InjectRepository(CheckpointRunner)
    private readonly checkpointRunnerRepository: Repository<CheckpointRunner>,
  ) {}

  /**
   * Creates a checkpoint entity from a completed snapshot-from-sandbox operation.
   * Called by the job completion handler after the runner successfully commits and pushes.
   */
  async createFromJobResult(
    sandboxId: string,
    organizationId: string,
    runnerId: string,
    name: string,
    ref: string,
    sizeBytes?: number,
    hash?: string,
  ): Promise<Checkpoint> {
    try {
      const checkpoint = this.checkpointRepository.create({
        sandboxId,
        organizationId,
        name,
        ref,
        state: CheckpointState.ACTIVE,
        sizeBytes,
        hash,
        runners: [
          {
            runnerId,
            state: CheckpointRunnerState.READY,
          },
        ],
      })

      const savedCheckpoint = await this.checkpointRepository.save(checkpoint)
      this.logger.log(`Checkpoint '${name}' created for sandbox ${sandboxId} with ref ${ref}`)
      return savedCheckpoint
    } catch (error) {
      if (error.code === '23505') {
        throw new ConflictException(`Checkpoint with name "${name}" already exists for this sandbox`)
      }
      throw error
    }
  }

  /**
   * List all checkpoints for a sandbox, sorted by creation date descending.
   */
  async listBySandbox(sandboxId: string, organizationId: string): Promise<Checkpoint[]> {
    return this.checkpointRepository.find({
      where: { sandboxId, organizationId },
      relations: ['runners'],
      order: { createdAt: 'DESC' },
    })
  }

  /**
   * Get a specific checkpoint by ID.
   */
  async getCheckpoint(checkpointId: string, organizationId: string): Promise<Checkpoint> {
    const checkpoint = await this.checkpointRepository.findOne({
      where: { id: checkpointId, organizationId },
      relations: ['runners'],
    })

    if (!checkpoint) {
      throw new NotFoundException(`Checkpoint ${checkpointId} not found`)
    }

    return checkpoint
  }

  /**
   * Delete a checkpoint. Sets state to REMOVING, then deletes.
   * Image cleanup from registry is handled by CheckpointManager.
   */
  async deleteCheckpoint(checkpointId: string, organizationId: string): Promise<void> {
    const checkpoint = await this.getCheckpoint(checkpointId, organizationId)

    if (checkpoint.state === CheckpointState.REMOVING) {
      throw new BadRequestException('Checkpoint is already being removed')
    }

    checkpoint.state = CheckpointState.REMOVING
    await this.checkpointRepository.save(checkpoint)

    // Delete checkpoint entity (cascade deletes CheckpointRunner entries)
    await this.checkpointRepository.remove(checkpoint)

    this.logger.log(`Checkpoint ${checkpointId} deleted`)
  }

  /**
   * Find a checkpoint runner entry for a specific runner.
   */
  async findCheckpointRunner(checkpointId: string, runnerId: string): Promise<CheckpointRunner | null> {
    return this.checkpointRunnerRepository.findOne({
      where: { checkpointId, runnerId },
    })
  }

  /**
   * Create or update a checkpoint runner entry (for when checkpoint needs to be pulled to a new runner).
   */
  async ensureCheckpointRunner(
    checkpointId: string,
    runnerId: string,
    state: CheckpointRunnerState,
  ): Promise<CheckpointRunner> {
    const existing = await this.checkpointRunnerRepository.findOne({
      where: { checkpointId, runnerId },
    })

    if (existing) {
      existing.state = state
      return this.checkpointRunnerRepository.save(existing)
    }

    const checkpointRunner = this.checkpointRunnerRepository.create({
      checkpointId,
      runnerId,
      state,
    })

    return this.checkpointRunnerRepository.save(checkpointRunner)
  }

  /**
   * Update a checkpoint runner state.
   */
  async updateCheckpointRunnerState(
    checkpointId: string,
    runnerId: string,
    state: CheckpointRunnerState,
    errorReason?: string,
  ): Promise<void> {
    await this.checkpointRunnerRepository.update(
      { checkpointId, runnerId },
      { state, errorReason: errorReason ?? null },
    )
  }

  /**
   * Get all checkpoints for a sandbox (used for cleanup on sandbox destroy).
   */
  async getCheckpointsBySandboxId(sandboxId: string): Promise<Checkpoint[]> {
    return this.checkpointRepository.find({
      where: { sandboxId },
      relations: ['runners'],
    })
  }
}
