/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DataSource, EntityManager, FindOptionsWhere, SelectQueryBuilder } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxStateEntity } from '../entities/sandbox-state.entity'
import { SandboxBackupEntity } from '../entities/sandbox-backup.entity'
import { SandboxAggregate } from '../types/sandbox-aggregate.type'
import { BackupState } from '../enums/backup-state.enum'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { SandboxLastActivity } from '../entities/sandbox-last-activity.entity'
import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { SandboxConflictError } from '../errors/sandbox-conflict.error'
import { InjectDataSource } from '@nestjs/typeorm'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { BaseRepository } from '../../common/repositories/base.repository'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'
import { SandboxDesiredStateUpdatedEvent } from '../events/sandbox-desired-state-updated.event'
import { SandboxPublicStatusUpdatedEvent } from '../events/sandbox-public-status-updated.event'
import { SandboxOrganizationUpdatedEvent } from '../events/sandbox-organization-updated.event'
import { SandboxLookupCacheInvalidationService } from '../services/sandbox-lookup-cache-invalidation.service'

const STATE_KEYS = new Set([
  'state',
  'desiredState',
  'pending',
  'errorReason',
  'recoverable',
  'runnerId',
  'prevRunnerId',
  'daemonVersion',
])

const BACKUP_KEYS = new Set([
  'backupState',
  'backupSnapshot',
  'backupRegistryId',
  'lastBackupAt',
  'backupErrorReason',
  'existingBackupSnapshots',
])

function partitionUpdate(data: Partial<SandboxAggregate>): {
  stateFields: Partial<SandboxStateEntity>
  backupFields: Partial<SandboxBackupEntity>
  configFields: Partial<Sandbox>
} {
  const stateFields: Record<string, unknown> = {}
  const backupFields: Record<string, unknown> = {}
  const configFields: Record<string, unknown> = {}

  for (const [key, value] of Object.entries(data)) {
    if (value === undefined) continue
    if (STATE_KEYS.has(key)) {
      stateFields[key] = value
    } else if (BACKUP_KEYS.has(key)) {
      backupFields[key] = value
    } else {
      configFields[key] = value
    }
  }

  return {
    stateFields: stateFields as Partial<SandboxStateEntity>,
    backupFields: backupFields as Partial<SandboxBackupEntity>,
    configFields: configFields as Partial<Sandbox>,
  }
}

function hasFields(obj: Record<string, unknown>): boolean {
  return Object.keys(obj).length > 0
}

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

  async insert(
    sandbox: Sandbox,
    stateData?: Partial<SandboxStateEntity>,
    backupData?: Partial<SandboxBackupEntity>,
  ): Promise<Sandbox> {
    const now = new Date()
    if (!sandbox.createdAt) {
      sandbox.createdAt = now
    }
    if (!sandbox.updatedAt) {
      sandbox.updatedAt = now
    }

    const stateEntity = new SandboxStateEntity()
    Object.assign(stateEntity, {
      sandboxId: sandbox.id,
      state: SandboxState.UNKNOWN,
      desiredState: SandboxDesiredState.STARTED,
      pending: false,
      recoverable: false,
      ...stateData,
    })
    stateEntity.assertValid(sandbox.id)
    const stateInvariants = stateEntity.enforceInvariants()
    Object.assign(stateEntity, stateInvariants)

    const backupEntity = new SandboxBackupEntity()
    Object.assign(backupEntity, {
      sandboxId: sandbox.id,
      backupState: BackupState.NONE,
      backupSnapshot: null,
      backupRegistryId: null,
      lastBackupAt: null,
      backupErrorReason: null,
      existingBackupSnapshots: [],
      ...backupData,
    })

    await this.dataSource.transaction(async (entityManager) => {
      await entityManager.insert(Sandbox, sandbox)
      await entityManager.insert(SandboxStateEntity, stateEntity)
      await entityManager.insert(SandboxBackupEntity, backupEntity)
      await this.upsertLastActivity(entityManager, sandbox.id, sandbox.createdAt)
    })

    sandbox.sandboxState = stateEntity
    sandbox.sandboxBackup = backupEntity

    this.invalidateLookupCacheOnInsert(sandbox)

    return sandbox
  }

  async update(id: string, params: { updateData: Partial<SandboxAggregate> }, raw: true): Promise<void>
  async update(
    id: string,
    params: { updateData: Partial<SandboxAggregate>; entity?: Sandbox },
    raw?: false,
  ): Promise<Sandbox>
  async update(
    id: string,
    params: { updateData: Partial<SandboxAggregate>; entity?: Sandbox },
    raw = false,
  ): Promise<Sandbox | void> {
    const { updateData, entity } = params

    if (raw) {
      const { stateFields, backupFields, configFields } = partitionUpdate(updateData)
      const stateRepo = this.dataSource.getRepository(SandboxStateEntity)
      const backupRepo = this.dataSource.getRepository(SandboxBackupEntity)

      if (hasFields(stateFields as Record<string, unknown>)) {
        await stateRepo.update({ sandboxId: id }, stateFields)
      }
      if (hasFields(backupFields as Record<string, unknown>)) {
        await backupRepo.update({ sandboxId: id }, backupFields)
      }
      if (hasFields(configFields as Record<string, unknown>)) {
        await this.repository.update(id, configFields)
      }
      return
    }

    const sandbox =
      entity ??
      (await this.findOne({
        where: { id },
        relations: ['sandboxState', 'sandboxBackup'],
      }))
    if (!sandbox) {
      throw new NotFoundException('Sandbox not found')
    }

    if (!sandbox.sandboxState) {
      const stateRepo = this.dataSource.getRepository(SandboxStateEntity)
      const row = await stateRepo.findOneBy({ sandboxId: id })
      if (row) sandbox.sandboxState = row
    }
    if (!sandbox.sandboxBackup) {
      const backupRepo = this.dataSource.getRepository(SandboxBackupEntity)
      const row = await backupRepo.findOneBy({ sandboxId: id })
      if (row) sandbox.sandboxBackup = row
    }

    const prevState = sandbox.sandboxState.state
    const prevDesiredState = sandbox.sandboxState.desiredState
    const prevPending = sandbox.sandboxState.pending
    const prevPublic = sandbox.public
    const prevOrganizationId = sandbox.organizationId
    const prevName = sandbox.name
    const prevAuthToken = sandbox.authToken

    const { stateFields, backupFields, configFields } = partitionUpdate(updateData)

    if (hasFields(configFields as Record<string, unknown>)) {
      Object.assign(sandbox, configFields)
    }
    if (hasFields(stateFields as Record<string, unknown>)) {
      Object.assign(sandbox.sandboxState, stateFields)
    }
    if (hasFields(backupFields as Record<string, unknown>)) {
      Object.assign(sandbox.sandboxBackup, backupFields)
    }

    sandbox.sandboxState.assertValid(id)
    const invariantChanges = sandbox.sandboxState.enforceInvariants()
    Object.assign(sandbox.sandboxState, invariantChanges)

    let crossTableBackup: Partial<SandboxBackupEntity> = {}
    if (sandbox.sandboxState.state === SandboxState.DESTROYED) {
      sandbox.sandboxBackup.backupState = BackupState.NONE
      crossTableBackup = { backupState: BackupState.NONE }
    }

    const allStateFields = { ...stateFields, ...invariantChanges } as Partial<SandboxStateEntity>
    const allBackupFields = { ...backupFields, ...crossTableBackup } as Partial<SandboxBackupEntity>

    await this.dataSource.transaction(async (entityManager) => {
      if (hasFields(allStateFields as Record<string, unknown>)) {
        const result = await entityManager.update(
          SandboxStateEntity,
          {
            sandboxId: id,
            state: prevState,
            desiredState: prevDesiredState,
            pending: prevPending,
          },
          allStateFields,
        )
        if (!result.affected) {
          throw new SandboxConflictError()
        }
      }

      if (hasFields(configFields as Record<string, unknown>)) {
        const result = await entityManager.update(Sandbox, { id, organizationId: prevOrganizationId }, configFields)
        if (!result.affected) {
          throw new SandboxConflictError()
        }
      }

      if (hasFields(allBackupFields as Record<string, unknown>)) {
        await entityManager.update(SandboxBackupEntity, { sandboxId: id }, allBackupFields)
      }

      sandbox.updatedAt = new Date()

      if (prevState !== sandbox.sandboxState.state || prevOrganizationId !== sandbox.organizationId) {
        await this.upsertLastActivity(entityManager, id, sandbox.updatedAt)
      }
    })

    this.emitUpdateEvents(sandbox, {
      state: prevState,
      desiredState: prevDesiredState,
      public: prevPublic,
      organizationId: prevOrganizationId,
    })
    this.invalidateLookupCacheOnUpdate(sandbox, {
      organizationId: prevOrganizationId,
      name: prevName,
      authToken: prevAuthToken,
    })

    return sandbox
  }

  async updateWhere(
    id: string,
    params: {
      updateData: Partial<SandboxAggregate>
      whereCondition: Partial<SandboxAggregate>
    },
  ): Promise<Sandbox> {
    const { updateData, whereCondition } = params

    const { stateFields: stateWhere, configFields: configWhere } = partitionUpdate(whereCondition)

    const result = await this.dataSource.transaction(async (entityManager) => {
      const stateRow = await entityManager.findOne(SandboxStateEntity, {
        where: { sandboxId: id, ...(stateWhere as FindOptionsWhere<SandboxStateEntity>) },
        lock: { mode: 'pessimistic_write' },
      })
      if (!stateRow) {
        throw new SandboxConflictError()
      }

      let sandbox: Sandbox
      if (hasFields(configWhere as Record<string, unknown>)) {
        const found = await entityManager.findOne(Sandbox, {
          where: { id, ...(configWhere as FindOptionsWhere<Sandbox>) },
          loadEagerRelations: false,
        })
        if (!found) {
          throw new SandboxConflictError()
        }
        sandbox = found
      } else {
        const found = await entityManager.findOne(Sandbox, {
          where: { id },
          loadEagerRelations: false,
        })
        if (!found) {
          throw new NotFoundException('Sandbox not found')
        }
        sandbox = found
      }

      sandbox.sandboxState = stateRow
      const backupRow = await entityManager.findOne(SandboxBackupEntity, { where: { sandboxId: id } })
      if (backupRow) {
        sandbox.sandboxBackup = backupRow
      }

      const prevState = sandbox.sandboxState.state
      const prevDesiredState = sandbox.sandboxState.desiredState
      const prevPublic = sandbox.public
      const prevOrganizationId = sandbox.organizationId
      const prevName = sandbox.name
      const prevAuthToken = sandbox.authToken

      const { stateFields, backupFields, configFields } = partitionUpdate(updateData)

      if (hasFields(configFields as Record<string, unknown>)) Object.assign(sandbox, configFields)
      if (hasFields(stateFields as Record<string, unknown>)) Object.assign(sandbox.sandboxState, stateFields)
      if (hasFields(backupFields as Record<string, unknown>)) Object.assign(sandbox.sandboxBackup, backupFields)

      sandbox.sandboxState.assertValid(id)
      const invariantChanges = sandbox.sandboxState.enforceInvariants()
      Object.assign(sandbox.sandboxState, invariantChanges)

      if (sandbox.sandboxState.state === SandboxState.DESTROYED) {
        sandbox.sandboxBackup.backupState = BackupState.NONE
      }

      const allStateFields = { ...stateFields, ...invariantChanges } as Partial<SandboxStateEntity>
      const allBackupFields =
        sandbox.sandboxState.state === SandboxState.DESTROYED
          ? ({ ...backupFields, backupState: BackupState.NONE } as Partial<SandboxBackupEntity>)
          : backupFields

      if (hasFields(allStateFields as Record<string, unknown>)) {
        await entityManager.update(SandboxStateEntity, { sandboxId: id }, allStateFields)
      }
      if (hasFields(configFields as Record<string, unknown>)) {
        await entityManager.update(Sandbox, id, configFields)
      }
      if (hasFields(allBackupFields as Record<string, unknown>)) {
        await entityManager.update(SandboxBackupEntity, { sandboxId: id }, allBackupFields)
      }

      sandbox.updatedAt = new Date()

      if (prevState !== sandbox.sandboxState.state || prevOrganizationId !== sandbox.organizationId) {
        await this.upsertLastActivity(entityManager, id, sandbox.updatedAt)
      }

      return {
        sandbox,
        prev: {
          state: prevState,
          desiredState: prevDesiredState,
          public: prevPublic,
          organizationId: prevOrganizationId,
          name: prevName,
          authToken: prevAuthToken,
        },
      }
    })

    this.emitUpdateEvents(result.sandbox, result.prev)
    this.invalidateLookupCacheOnUpdate(result.sandbox, result.prev)

    return result.sandbox
  }

  async updateState(
    id: string,
    updateData: Partial<SandboxStateEntity>,
    whereCondition: FindOptionsWhere<SandboxStateEntity>,
  ): Promise<Sandbox> {
    const result = await this.dataSource.transaction(async (entityManager) => {
      const stateRow = await entityManager.findOne(SandboxStateEntity, {
        where: { sandboxId: id, ...whereCondition },
        lock: { mode: 'pessimistic_write' },
      })
      if (!stateRow) {
        throw new SandboxConflictError()
      }

      const previousState = stateRow.state
      const previousDesiredState = stateRow.desiredState

      Object.assign(stateRow, updateData)
      const invariants = stateRow.enforceInvariants()
      const allStateChanges = { ...updateData, ...invariants }

      await entityManager.update(SandboxStateEntity, { sandboxId: id }, allStateChanges)

      if (stateRow.state === SandboxState.DESTROYED) {
        await entityManager.update(SandboxBackupEntity, { sandboxId: id }, { backupState: BackupState.NONE })
      }

      const now = new Date()
      if (previousState !== stateRow.state) {
        await this.upsertLastActivity(entityManager, id, now)
      }

      const sandbox = await entityManager.findOne(Sandbox, {
        where: { id },
        loadEagerRelations: false,
      })
      if (!sandbox) {
        throw new NotFoundException('Sandbox not found')
      }
      sandbox.sandboxState = stateRow

      return { sandbox, previousState, previousDesiredState }
    })

    if (result.previousState !== result.sandbox.sandboxState.state) {
      this.eventEmitter.emit(
        SandboxEvents.STATE_UPDATED,
        new SandboxStateUpdatedEvent(result.sandbox, result.previousState, result.sandbox.sandboxState.state),
      )
    }
    if (result.previousDesiredState !== result.sandbox.sandboxState.desiredState) {
      this.eventEmitter.emit(
        SandboxEvents.DESIRED_STATE_UPDATED,
        new SandboxDesiredStateUpdatedEvent(
          result.sandbox,
          result.previousDesiredState,
          result.sandbox.sandboxState.desiredState,
        ),
      )
    }

    return result.sandbox
  }

  async updateBackup(id: string, updateData: Partial<SandboxBackupEntity>): Promise<void> {
    await this.dataSource.getRepository(SandboxBackupEntity).update({ sandboxId: id }, updateData)
  }

  createAggregateQueryBuilder(alias = 'sandbox'): SelectQueryBuilder<Sandbox> {
    return this.repository
      .createQueryBuilder(alias)
      .innerJoin('sandbox_state', 'ss', `ss."sandboxId" = ${alias}."id"`)
      .innerJoin('sandbox_backup', 'sb', `sb."sandboxId" = ${alias}."id"`)
  }

  private async upsertLastActivity(
    entityManager: EntityManager,
    sandboxId: string,
    lastActivityAt: Date,
  ): Promise<void> {
    await entityManager.upsert(SandboxLastActivity, { sandboxId, lastActivityAt }, ['sandboxId'])
  }

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

  private invalidateLookupCacheOnUpdate(
    updatedSandbox: Sandbox,
    previousSandbox: { organizationId: string; name: string; authToken: string },
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

  private emitUpdateEvents(
    updatedSandbox: Sandbox,
    previous: { state: SandboxState; desiredState: SandboxDesiredState; public: boolean; organizationId: string },
  ): void {
    if (previous.state !== updatedSandbox.sandboxState.state) {
      this.eventEmitter.emit(
        SandboxEvents.STATE_UPDATED,
        new SandboxStateUpdatedEvent(updatedSandbox, previous.state, updatedSandbox.sandboxState.state),
      )
    }

    if (previous.desiredState !== updatedSandbox.sandboxState.desiredState) {
      this.eventEmitter.emit(
        SandboxEvents.DESIRED_STATE_UPDATED,
        new SandboxDesiredStateUpdatedEvent(
          updatedSandbox,
          previous.desiredState,
          updatedSandbox.sandboxState.desiredState,
        ),
      )
    }

    if (previous.public !== updatedSandbox.public) {
      this.eventEmitter.emit(
        SandboxEvents.PUBLIC_STATUS_UPDATED,
        new SandboxPublicStatusUpdatedEvent(updatedSandbox, previous.public, updatedSandbox.public),
      )
    }

    if (previous.organizationId !== updatedSandbox.organizationId) {
      this.eventEmitter.emit(
        SandboxEvents.ORGANIZATION_UPDATED,
        new SandboxOrganizationUpdatedEvent(updatedSandbox, previous.organizationId, updatedSandbox.organizationId),
      )
    }
  }
}
