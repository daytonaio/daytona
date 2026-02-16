/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Repository,
  DataSource,
  FindOptionsWhere,
  FindOneOptions,
  FindManyOptions,
  SelectQueryBuilder,
  DeleteResult,
} from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { ConflictException, Injectable, NotFoundException } from '@nestjs/common'
import { InjectDataSource } from '@nestjs/typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxCreatedEvent } from '../events/sandbox-create.event'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'
import { SandboxDesiredStateUpdatedEvent } from '../events/sandbox-desired-state-updated.event'
import { SandboxPublicStatusUpdatedEvent } from '../events/sandbox-public-status-updated.event'
import { SandboxOrganizationUpdatedEvent } from '../events/sandbox-organization-updated.event'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { BackupState } from '../enums/backup-state.enum'

@Injectable()
export class SandboxRepository {
  private repository: Repository<Sandbox>

  constructor(
    @InjectDataSource() private dataSource: DataSource,
    private eventEmitter: EventEmitter2,
  ) {
    this.repository = this.dataSource.getRepository(Sandbox)
  }

  /**
   * See reference for {@link Repository.findOne}
   */
  async findOne(options: FindOneOptions<Sandbox>): Promise<Sandbox | null> {
    return this.repository.findOne(options)
  }

  /**
   * See reference for {@link Repository.findOneBy}
   */
  async findOneBy(where: FindOptionsWhere<Sandbox> | FindOptionsWhere<Sandbox>[]): Promise<Sandbox | null> {
    return this.repository.findOneBy(where)
  }

  /**
   * See reference for {@link Repository.findOneByOrFail}
   */
  async findOneByOrFail(where: FindOptionsWhere<Sandbox> | FindOptionsWhere<Sandbox>[]): Promise<Sandbox> {
    return this.repository.findOneByOrFail(where)
  }

  /**
   * See reference for {@link Repository.find}
   */
  async find(options?: FindManyOptions<Sandbox>): Promise<Sandbox[]> {
    return this.repository.find(options)
  }

  /**
   * See reference for {@link Repository.findAndCount}
   */
  async findAndCount(options?: FindManyOptions<Sandbox>): Promise<[Sandbox[], number]> {
    return this.repository.findAndCount(options)
  }

  /**
   * See reference for {@link Repository.count}
   */
  async count(options?: FindManyOptions<Sandbox>): Promise<number> {
    return this.repository.count(options)
  }

  /**
   * See reference for {@link Repository.createQueryBuilder}
   */
  createQueryBuilder(alias = 'sandbox'): SelectQueryBuilder<Sandbox> {
    return this.repository.createQueryBuilder(alias)
  }

  /**
   * See reference for {@link Repository.manager}
   */
  get manager() {
    return this.repository.manager
  }

  /**
   * Inserts a new sandbox into the database and emits a {@link SandboxCreatedEvent} event.
   *
   * Uses {@link Repository.insert} to insert the sandbox into the database.
   */
  async insert(sandbox: Sandbox): Promise<Sandbox> {
    const result = await this.repository.insert(sandbox)

    const insertedSandbox = await this.findOneBy({ id: result.identifiers[0].id })
    if (!insertedSandbox) {
      throw new NotFoundException('Sandbox not found after insert')
    }

    this.eventEmitter.emit(SandboxEvents.CREATED, new SandboxCreatedEvent(insertedSandbox))

    return insertedSandbox
  }

  /**
   * Partially updates a sandbox in the database and emits a corresponding event based on the changes.
   *
   * Uses {@link Repository.update} to update the sandbox in the database.
   */
  async update(sandboxId: string, updateData: Partial<Sandbox>): Promise<Sandbox> {
    const sandbox = await this.findOneBy({ id: sandboxId })
    if (!sandbox) {
      throw new NotFoundException('Sandbox not found')
    }

    // Store old sandbox for event emission
    const oldSandbox = { ...sandbox }

    Object.assign(sandbox, updateData)
    this.applyBeforeUpdateLogic(sandbox)

    const result = await this.repository.update(sandboxId, updateData)
    if (!result.affected) {
      throw new NotFoundException('Sandbox not found after update')
    }

    this.emitUpdateEvents(sandbox, oldSandbox)

    return sandbox
  }

  /**
   * Partially updates a sandbox in the database and emits a corresponding event based on the changes.
   *
   * Performs the update in a transaction with a pessimistic write lock to ensure consistency.
   *
   * @throws {ConflictException} if the sandbox was modified by another operation
   */
  async updateWhere(
    sandboxId: string,
    params: {
      updateData: Partial<Sandbox>
      whereCondition: FindOptionsWhere<Sandbox>
    },
  ): Promise<Sandbox> {
    const { updateData, whereCondition } = params

    return this.manager.transaction(async (entityManager) => {
      const whereClause = {
        ...whereCondition,
        id: sandboxId,
      }

      const sandbox = await entityManager.findOne(Sandbox, {
        where: whereClause,
        lock: { mode: 'pessimistic_write' },
        relations: [],
        loadEagerRelations: false,
      })

      if (!sandbox) {
        throw new ConflictException('Sandbox was modified by another operation, please try again')
      }

      // Store old sandbox for event emission
      const oldSandbox = { ...sandbox }

      Object.assign(sandbox, updateData)
      this.applyBeforeUpdateLogic(sandbox)

      await entityManager.update(Sandbox, sandboxId, updateData)

      this.emitUpdateEvents(sandbox, oldSandbox)

      return sandbox
    })
  }

  /**
   * See reference for {@link Repository.delete}
   */
  async delete(criteria: FindOptionsWhere<Sandbox> | FindOptionsWhere<Sandbox>[]): Promise<DeleteResult> {
    return this.repository.delete(criteria)
  }

  /**
   * Applies the necessary validations and changes in preparation for saving the sandbox changes to the database.
   */
  private applyBeforeUpdateLogic(updatedSandbox: Sandbox): void {
    this.validateDesiredState(updatedSandbox)
    this.updatePendingFlag(updatedSandbox)
    this.handleDestroyedState(updatedSandbox)
  }

  /**
   * @throws {Error} if the sandbox is not in a valid state to transition to the desired state
   */
  private validateDesiredState(sandbox: Sandbox): void {
    switch (sandbox.desiredState) {
      case SandboxDesiredState.STARTED:
        if (
          [
            SandboxState.STARTED,
            SandboxState.STOPPED,
            SandboxState.STARTING,
            SandboxState.ARCHIVED,
            SandboxState.CREATING,
            SandboxState.UNKNOWN,
            SandboxState.RESTORING,
            SandboxState.PENDING_BUILD,
            SandboxState.BUILDING_SNAPSHOT,
            SandboxState.PULLING_SNAPSHOT,
            SandboxState.ARCHIVING,
            SandboxState.ERROR,
            SandboxState.BUILD_FAILED,
          ].includes(sandbox.state)
        ) {
          break
        }
        throw new Error(`Sandbox ${sandbox.id} is not in a valid state to be started. State: ${sandbox.state}`)
      case SandboxDesiredState.STOPPED:
        if (
          [
            SandboxState.STARTED,
            SandboxState.STOPPING,
            SandboxState.STOPPED,
            SandboxState.ERROR,
            SandboxState.BUILD_FAILED,
          ].includes(sandbox.state)
        ) {
          break
        }
        throw new Error(`Sandbox ${sandbox.id} is not in a valid state to be stopped. State: ${sandbox.state}`)
      case SandboxDesiredState.ARCHIVED:
        if (
          [
            SandboxState.ARCHIVED,
            SandboxState.ARCHIVING,
            SandboxState.STOPPED,
            SandboxState.ERROR,
            SandboxState.BUILD_FAILED,
          ].includes(sandbox.state)
        ) {
          break
        }
        throw new Error(`Sandbox ${sandbox.id} is not in a valid state to be archived. State: ${sandbox.state}`)
      case SandboxDesiredState.DESTROYED:
        if (
          [
            SandboxState.DESTROYED,
            SandboxState.DESTROYING,
            SandboxState.STOPPED,
            SandboxState.STARTED,
            SandboxState.ARCHIVED,
            SandboxState.ERROR,
            SandboxState.BUILD_FAILED,
            SandboxState.ARCHIVING,
          ].includes(sandbox.state)
        ) {
          break
        }
        throw new Error(`Sandbox ${sandbox.id} is not in a valid state to be destroyed. State: ${sandbox.state}`)
    }
  }

  /**
   * Sets the pending flag for specific combinations of state and desired state.
   */
  private updatePendingFlag(sandbox: Sandbox): void {
    if (!sandbox.pending && String(sandbox.state) !== String(sandbox.desiredState)) {
      sandbox.pending = true
    }
    if (sandbox.pending && String(sandbox.state) === String(sandbox.desiredState)) {
      sandbox.pending = false
    }
    if (
      sandbox.state === SandboxState.ERROR ||
      sandbox.state === SandboxState.BUILD_FAILED ||
      sandbox.desiredState === SandboxDesiredState.ARCHIVED
    ) {
      sandbox.pending = false
    }
  }

  /**
   * Performs cleanup when a sandbox reaches destroyed state.
   */
  private handleDestroyedState(sandbox: Sandbox): void {
    if (sandbox.state === SandboxState.DESTROYED) {
      sandbox.runnerId = null
      sandbox.backupState = BackupState.NONE
    }
  }

  /**
   * Emits events based on the changes made to a sandbox.
   */
  private emitUpdateEvents(updatedSandbox: Sandbox, oldSandbox: Sandbox): void {
    if (oldSandbox.state !== updatedSandbox.state) {
      this.eventEmitter.emit(
        SandboxEvents.STATE_UPDATED,
        new SandboxStateUpdatedEvent(updatedSandbox, oldSandbox.state, updatedSandbox.state),
      )
    }

    if (oldSandbox.desiredState !== updatedSandbox.desiredState) {
      this.eventEmitter.emit(
        SandboxEvents.DESIRED_STATE_UPDATED,
        new SandboxDesiredStateUpdatedEvent(updatedSandbox, oldSandbox.desiredState, updatedSandbox.desiredState),
      )
    }

    // TODO: do we need both old and new status since it's a boolean?
    if (oldSandbox.public !== updatedSandbox.public) {
      this.eventEmitter.emit(
        SandboxEvents.PUBLIC_STATUS_UPDATED,
        new SandboxPublicStatusUpdatedEvent(updatedSandbox, oldSandbox.public, updatedSandbox.public),
      )
    }

    if (oldSandbox.organizationId !== updatedSandbox.organizationId) {
      this.eventEmitter.emit(
        SandboxEvents.ORGANIZATION_UPDATED,
        new SandboxOrganizationUpdatedEvent(updatedSandbox, oldSandbox.organizationId, updatedSandbox.organizationId),
      )
    }
  }
}
