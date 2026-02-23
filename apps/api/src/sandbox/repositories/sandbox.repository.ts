/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DataSource, FindOptionsWhere } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { ConflictException, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectDataSource } from '@nestjs/typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { BaseRepository } from '../../common/repositories/base.repository'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'
import { SandboxDesiredStateUpdatedEvent } from '../events/sandbox-desired-state-updated.event'
import { SandboxPublicStatusUpdatedEvent } from '../events/sandbox-public-status-updated.event'
import { SandboxOrganizationUpdatedEvent } from '../events/sandbox-organization-updated.event'
import { SandboxLookupCacheInvalidationService } from '../services/sandbox-lookup-cache-invalidation.service'

@Injectable()
export class SandboxRepository extends BaseRepository<Sandbox> {
  private readonly logger = new Logger(SandboxRepository.name)

  constructor(
    @InjectDataSource() dataSource: DataSource,
    eventEmitter: EventEmitter2,
    private readonly sandboxLookupCacheInvalidationService: SandboxLookupCacheInvalidationService,
  ) {
    super(dataSource, eventEmitter, Sandbox)
  }

  async insert(sandbox: Sandbox): Promise<Sandbox> {
    const now = new Date()
    if (!sandbox.createdAt) {
      sandbox.createdAt = now
    }
    if (!sandbox.updatedAt) {
      sandbox.updatedAt = now
    }
    if (!sandbox.lastActivityAt) {
      sandbox.lastActivityAt = now
    }

    sandbox.assertValid()
    sandbox.enforceInvariants()

    await this.repository.insert(sandbox)

    this.invalidateLookupCacheOnInsert(sandbox)

    return sandbox
  }

  /**
   * @param id - The ID of the sandbox to update.
   * @param params.updateData - The partial data to update.
   *
   * @returns `void` because a raw update is performed.
   */
  async update(id: string, params: { updateData: Partial<Sandbox> }, raw: true): Promise<void>
  /**
   * @param id - The ID of the sandbox to update.
   * @param params.updateData - The partial data to update.
   * @param params.entity - Optional pre-fetched sandbox to use instead of fetching from the database.
   *
   * @returns The updated sandbox.
   */
  async update(id: string, params: { updateData: Partial<Sandbox>; entity?: Sandbox }, raw?: false): Promise<Sandbox>
  async update(
    id: string,
    params: { updateData: Partial<Sandbox>; entity?: Sandbox },
    raw = false,
  ): Promise<Sandbox | void> {
    const { updateData, entity } = params

    if (updateData.state && !updateData.lastActivityAt) {
      updateData.lastActivityAt = new Date()
    }

    if (raw) {
      await this.repository.update(id, updateData)
      return
    }

    const sandbox = entity ?? (await this.findOneBy({ id }))
    if (!sandbox) {
      throw new NotFoundException('Sandbox not found')
    }

    const previousSandbox = { ...sandbox }

    Object.assign(sandbox, updateData)
    sandbox.assertValid()
    const invariantChanges = sandbox.enforceInvariants()

    const result = await this.repository.update(id, { ...updateData, ...invariantChanges })
    if (!result.affected) {
      throw new NotFoundException('Sandbox not found after update')
    }
    sandbox.updatedAt = new Date()

    this.emitUpdateEvents(sandbox, previousSandbox)
    this.invalidateLookupCacheOnUpdate(sandbox, previousSandbox)

    return sandbox
  }

  /**
   * Partially updates a sandbox in the database and optionally emits a corresponding event based on the changes.
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

    if (updateData.state && !updateData.lastActivityAt) {
      updateData.lastActivityAt = new Date()
    }

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

      const previousSandbox = { ...sandbox }

      Object.assign(sandbox, updateData)
      sandbox.assertValid()
      const invariantChanges = sandbox.enforceInvariants()

      await entityManager.update(Sandbox, id, { ...updateData, ...invariantChanges })
      sandbox.updatedAt = new Date()

      this.emitUpdateEvents(sandbox, previousSandbox)
      this.invalidateLookupCacheOnUpdate(sandbox, previousSandbox)

      return sandbox
    })
  }

  /**
   * Invalidates the sandbox lookup cache for the inserted sandbox.
   */
  private invalidateLookupCacheOnInsert(sandbox: Sandbox): void {
    try {
      this.sandboxLookupCacheInvalidationService.invalidateOrgId({
        sandboxId: sandbox.id,
        organizationId: sandbox.organizationId,
        name: sandbox.name,
      })
    } catch (error) {
      this.logger.warn(
        `Failed to enqueue sandbox lookup cache invalidation on insert (id, organizationId, name) for ${sandbox.id}: ${error instanceof Error ? error.message : String(error)}`,
      )
    }
  }

  /**
   * Invalidates the sandbox lookup cache for the updated sandbox.
   */
  private invalidateLookupCacheOnUpdate(
    updatedSandbox: Sandbox,
    previousSandbox: Pick<Sandbox, 'organizationId' | 'name' | 'authToken'>,
  ): void {
    try {
      this.sandboxLookupCacheInvalidationService.invalidate({
        sandboxId: updatedSandbox.id,
        organizationId: updatedSandbox.organizationId,
        previousOrganizationId: previousSandbox.organizationId,
        name: updatedSandbox.name,
        previousName: previousSandbox.name,
      })
    } catch (error) {
      this.logger.warn(
        `Failed to enqueue sandbox lookup cache invalidation on update (id, organizationId, name) for ${updatedSandbox.id}: ${error instanceof Error ? error.message : String(error)}`,
      )
    }

    try {
      if (updatedSandbox.authToken !== previousSandbox.authToken) {
        this.sandboxLookupCacheInvalidationService.invalidate({
          authToken: updatedSandbox.authToken,
        })
      }
    } catch (error) {
      this.logger.warn(
        `Failed to enqueue sandbox lookup cache invalidation on update (authToken) for ${updatedSandbox.id}: ${error instanceof Error ? error.message : String(error)}`,
      )
    }
  }

  /**
   * Emits events based on the changes made to a sandbox.
   */
  private emitUpdateEvents(
    updatedSandbox: Sandbox,
    previousSandbox: Pick<Sandbox, 'state' | 'desiredState' | 'public' | 'organizationId'>,
  ): void {
    if (previousSandbox.state !== updatedSandbox.state) {
      this.eventEmitter.emit(
        SandboxEvents.STATE_UPDATED,
        new SandboxStateUpdatedEvent(updatedSandbox, previousSandbox.state, updatedSandbox.state),
      )
    }

    if (previousSandbox.desiredState !== updatedSandbox.desiredState) {
      this.eventEmitter.emit(
        SandboxEvents.DESIRED_STATE_UPDATED,
        new SandboxDesiredStateUpdatedEvent(updatedSandbox, previousSandbox.desiredState, updatedSandbox.desiredState),
      )
    }

    if (previousSandbox.public !== updatedSandbox.public) {
      this.eventEmitter.emit(
        SandboxEvents.PUBLIC_STATUS_UPDATED,
        new SandboxPublicStatusUpdatedEvent(updatedSandbox, previousSandbox.public, updatedSandbox.public),
      )
    }

    if (previousSandbox.organizationId !== updatedSandbox.organizationId) {
      this.eventEmitter.emit(
        SandboxEvents.ORGANIZATION_UPDATED,
        new SandboxOrganizationUpdatedEvent(
          updatedSandbox,
          previousSandbox.organizationId,
          updatedSandbox.organizationId,
        ),
      )
    }
  }
}
