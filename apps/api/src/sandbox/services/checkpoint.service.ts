/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException, BadRequestException, ConflictException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { Checkpoint } from '../entities/checkpoint.entity'
import { CheckpointState } from '../enums/checkpoint-state.enum'
import { Sandbox } from '../entities/sandbox.entity'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'
import { CheckpointSortField, CheckpointSortDirection } from '../dto/list-checkpoints-query.dto'

@Injectable()
export class CheckpointService {
  private readonly logger = new Logger(CheckpointService.name)

  constructor(
    @InjectRepository(Checkpoint)
    private readonly checkpointRepository: Repository<Checkpoint>,
    @InjectRepository(SnapshotRunner)
    private readonly snapshotRunnerRepository: Repository<SnapshotRunner>,
  ) {}

  async createPending(
    sandbox: Sandbox,
    name: string,
    runnerId: string,
  ): Promise<Checkpoint> {
    try {
      const checkpoint = this.checkpointRepository.create({
        originSandboxId: sandbox.id,
        organizationId: sandbox.organizationId,
        name,
        state: CheckpointState.CREATING,
        region: sandbox.region,
        osUser: sandbox.osUser,
        cpu: sandbox.cpu,
        gpu: sandbox.gpu,
        mem: sandbox.mem,
        disk: sandbox.disk,
        env: sandbox.env,
        public: sandbox.public,
        networkBlockAll: sandbox.networkBlockAll,
        networkAllowList: sandbox.networkAllowList,
        labels: sandbox.labels,
        volumes: sandbox.volumes || [],
        class: sandbox.class,
        autoStopInterval: sandbox.autoStopInterval,
        autoArchiveInterval: sandbox.autoArchiveInterval,
        autoDeleteInterval: sandbox.autoDeleteInterval,
        buildInfoSnapshotRef: sandbox.buildInfo?.snapshotRef,
        initialRunnerId: runnerId,
      })

      const savedCheckpoint = await this.checkpointRepository.save(checkpoint)
      this.logger.log(`Checkpoint '${name}' created in CREATING state for sandbox ${sandbox.id}`)
      return savedCheckpoint
    } catch (error) {
      if (error.code === '23505') {
        throw new ConflictException(`Checkpoint with name "${name}" already exists for this sandbox`)
      }
      throw error
    }
  }

  async markActive(
    checkpointId: string,
    ref: string,
    runnerId: string,
    sizeBytes?: number,
  ): Promise<Checkpoint> {
    const checkpoint = await this.checkpointRepository.findOne({ where: { id: checkpointId } })
    if (!checkpoint) {
      throw new NotFoundException(`Checkpoint ${checkpointId} not found`)
    }

    checkpoint.ref = ref
    checkpoint.state = CheckpointState.ACTIVE
    checkpoint.size = sizeBytes ? sizeBytes / (1024 * 1024 * 1024) : undefined

    const saved = await this.checkpointRepository.save(checkpoint)

    if (ref) {
      await this.ensureSnapshotRunnerEntry(ref, runnerId)
    }

    this.logger.log(`Checkpoint '${checkpoint.name}' marked ACTIVE with ref ${ref}`)
    return saved
  }

  async markError(checkpointId: string, errorReason: string): Promise<void> {
    const checkpoint = await this.checkpointRepository.findOne({ where: { id: checkpointId } })
    if (!checkpoint) {
      return
    }

    checkpoint.state = CheckpointState.ERROR
    checkpoint.errorReason = errorReason
    await this.checkpointRepository.save(checkpoint)
    this.logger.warn(`Checkpoint '${checkpoint.name}' marked ERROR: ${errorReason}`)
  }

  async getCreatingCheckpoints(): Promise<Checkpoint[]> {
    return this.checkpointRepository.find({
      where: { state: CheckpointState.CREATING },
    })
  }

  async list(
    organizationId: string,
    page = 1,
    limit = 100,
    sandboxId?: string,
    sort?: { field?: CheckpointSortField; direction?: CheckpointSortDirection },
  ): Promise<PaginatedList<Checkpoint>> {
    const pageNum = Number(page)
    const limitNum = Number(limit)
    const { field: sortField, direction: sortDirection } = sort || {}

    const qb = this.checkpointRepository.createQueryBuilder('checkpoint')

    qb.where('checkpoint.organizationId = :organizationId', { organizationId })

    if (sandboxId) {
      qb.andWhere('checkpoint.originSandboxId = :sandboxId', { sandboxId })
    }

    const orderField = sortField || CheckpointSortField.CREATED_AT
    const orderDirection = (sortDirection || CheckpointSortDirection.DESC).toUpperCase() as 'ASC' | 'DESC'
    qb.orderBy(`checkpoint.${orderField}`, orderDirection)

    qb.skip((pageNum - 1) * limitNum).take(limitNum)

    const [items, total] = await qb.getManyAndCount()

    return {
      items,
      total,
      page: pageNum,
      totalPages: Math.ceil(total / limitNum),
    }
  }

  async getCheckpointById(checkpointId: string): Promise<Checkpoint> {
    const checkpoint = await this.checkpointRepository.findOne({
      where: { id: checkpointId },
    })

    if (!checkpoint) {
      throw new NotFoundException(`Checkpoint ${checkpointId} not found`)
    }

    return checkpoint
  }

  async getCheckpoint(checkpointId: string, organizationId: string): Promise<Checkpoint> {
    const checkpoint = await this.checkpointRepository.findOne({
      where: { id: checkpointId, organizationId },
    })

    if (!checkpoint) {
      throw new NotFoundException(`Checkpoint ${checkpointId} not found`)
    }

    return checkpoint
  }

  async deleteCheckpoint(checkpointId: string, organizationId: string): Promise<void> {
    const checkpoint = await this.getCheckpoint(checkpointId, organizationId)

    if (checkpoint.state === CheckpointState.REMOVING) {
      throw new BadRequestException('Checkpoint is already being removed')
    }

    checkpoint.state = CheckpointState.REMOVING
    await this.checkpointRepository.save(checkpoint)
    await this.checkpointRepository.remove(checkpoint)

    this.logger.log(`Checkpoint ${checkpointId} deleted`)
  }

  async getCheckpointsBySandboxId(sandboxId: string): Promise<Checkpoint[]> {
    return this.checkpointRepository.find({
      where: { originSandboxId: sandboxId },
      order: { createdAt: 'DESC' },
    })
  }

  async countByOrganization(organizationId: string): Promise<number> {
    return this.checkpointRepository.count({
      where: { organizationId },
    })
  }

  private async ensureSnapshotRunnerEntry(ref: string, runnerId: string): Promise<void> {
    const existing = await this.snapshotRunnerRepository.findOne({
      where: { snapshotRef: ref, runnerId },
    })

    if (existing) {
      return
    }

    const snapshotRunner = new SnapshotRunner()
    snapshotRunner.snapshotRef = ref
    snapshotRunner.runnerId = runnerId
    snapshotRunner.state = SnapshotRunnerState.READY

    try {
      await this.snapshotRunnerRepository.save(snapshotRunner)
    } catch (error) {
      if (error.code === '23505') {
        return
      }
      throw error
    }
  }
}
