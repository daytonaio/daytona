/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DataSource, FindOptionsWhere } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { ConflictException, Injectable, NotFoundException } from '@nestjs/common'
import { InjectDataSource } from '@nestjs/typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { BaseRepository } from '../../common/repositories/base.repository'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxCreatedEvent } from '../events/sandbox-create.event'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'
import { SandboxDesiredStateUpdatedEvent } from '../events/sandbox-desired-state-updated.event'
import { SandboxPublicStatusUpdatedEvent } from '../events/sandbox-public-status-updated.event'
import { SandboxOrganizationUpdatedEvent } from '../events/sandbox-organization-updated.event'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { BackupState } from '../enums/backup-state.enum'
import { SandboxLookupCacheInvalidationService } from '../services/sandbox-lookup-cache-invalidation.service'

@Injectable()
export class SandboxRepository extends BaseRepository<Sandbox> {
  constructor(
    @InjectDataSource() dataSource: DataSource,
    eventEmitter: EventEmitter2,
    private readonly sandboxLookupCacheInvalidationService: SandboxLookupCacheInvalidationService,
  ) {
    super(dataSource, eventEmitter, Sandbox)
  }

  async insert(sandbox: Sandbox): Promise<Sandbox> {
    const result = await this.repository.insert(sandbox)

    const insertedSandbox = await this.findOneBy({ id: result.identifiers[0].id })
    if (!insertedSandbox) {
      throw new NotFoundException('Sandbox not found after insert')
    }

    this.eventEmitter.emit(SandboxEvents.CREATED, new SandboxCreatedEvent(insertedSandbox))

    return insertedSandbox
  }

  async update(id: string, params: { updateData: Partial<Sandbox>; entity?: Sandbox }, raw: true): Promise<void>
  async update(id: string, params: { updateData: Partial<Sandbox>; entity?: Sandbox }, raw?: false): Promise<Sandbox>
  async update(
    id: string,
    params: { updateData: Partial<Sandbox>; entity?: Sandbox },
    raw = false,
  ): Promise<Sandbox | void> {
    const { updateData, entity } = params

    if (raw) {
      await this.repository.update(id, updateData)
      return
    }

    const sandbox = entity ?? (await this.findOneBy({ id }))
    if (!sandbox) {
      throw new NotFoundException('Sandbox not found')
    }

    // Store old sandbox for event emission
    const oldSandbox = { ...sandbox }

    Object.assign(sandbox, updateData)
    this.applyBeforeUpdateLogic(sandbox, updateData)

    const result = await this.repository.update(id, updateData)
    if (!result.affected) {
      throw new NotFoundException('Sandbox not found after update')
    }

    this.emitUpdateEvents(sandbox, oldSandbox)
    this.invalidateLookupCache(sandbox, oldSandbox)

    return sandbox
  }

  /**
   * Partially updates a sandbox in the database and emits a corresponding event based on the changes.
   *
   * Performs the update in a transaction with a pessimistic write lock to ensure consistency.
   *
   * @param id - The ID of the sandbox to update.
   * @param params.updateData - The partial data to update.
   * @param params.whereCondition - The where condition to use for the update.
   *
   * @throws {ConflictException} if the sandbox was modified by another operation
   */
  async updateWhere(
    id: string,
    params: {
      updateData: Partial<Sandbox>
      whereCondition: FindOptionsWhere<Sandbox>
    },
  ): Promise<Sandbox> {
    const { updateData, whereCondition } = params

    return this.manager.transaction(async (entityManager) => {
      const whereClause = {
        ...whereCondition,
        id,
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
      this.applyBeforeUpdateLogic(sandbox, updateData)

      await entityManager.update(Sandbox, id, updateData)

      this.emitUpdateEvents(sandbox, oldSandbox)
      this.invalidateLookupCache(sandbox, oldSandbox)

      return sandbox
    })
  }

  /**
   * Applies the necessary validations and changes in preparation for saving the sandbox changes to the database.
   * Derived fields are written to both the in-memory entity and `updateData` so they are persisted.
   */
  private applyBeforeUpdateLogic(updatedSandbox: Sandbox, updateData: Partial<Sandbox>): void {
    this.validateDesiredState(updatedSandbox)
    this.updatePendingFlag(updatedSandbox, updateData)
    this.handleDestroyedState(updatedSandbox, updateData)
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
            SandboxState.RESIZING,
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
  private updatePendingFlag(sandbox: Sandbox, updateData: Partial<Sandbox>): void {
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
    updateData.pending = sandbox.pending
  }

  /**
   * Performs cleanup when a sandbox reaches destroyed state.
   */
  private handleDestroyedState(sandbox: Sandbox, updateData: Partial<Sandbox>): void {
    if (sandbox.state === SandboxState.DESTROYED) {
      sandbox.runnerId = null
      sandbox.backupState = BackupState.NONE
      updateData.runnerId = null
      updateData.backupState = BackupState.NONE
    }
  }

  /**
   * Invalidates the sandbox lookup cache for the updated sandbox.
   */
  private invalidateLookupCache(updatedSandbox: Sandbox, oldSandbox: Pick<Sandbox, 'organizationId' | 'name'>): void {
    this.sandboxLookupCacheInvalidationService.invalidate({
      sandboxId: updatedSandbox.id,
      organizationId: updatedSandbox.organizationId,
      previousOrganizationId: oldSandbox.organizationId,
      name: updatedSandbox.name,
      previousName: oldSandbox.name,
    })
  }

  /**
   * Emits events based on the changes made to a sandbox.
   */
  private emitUpdateEvents(
    updatedSandbox: Sandbox,
    oldSandbox: Pick<Sandbox, 'state' | 'desiredState' | 'public' | 'organizationId'>,
  ): void {
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
