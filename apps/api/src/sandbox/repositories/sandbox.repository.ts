/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  DataSource,
  EntityManager,
  EntityNotFoundError,
  FindManyOptions,
  FindOneOptions,
  FindOptionsRelations,
  FindOptionsWhere,
} from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
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
import { SandboxLifecycleMigrationService } from '../services/sandbox-lifecycle-migration.service'
import { SandboxFork } from '../entities/sandbox-fork.entity'
import { SandboxLifecycle } from '../entities/sandbox-lifecycle.entity'
import {
  SANDBOX_CONFIG_COLUMNS,
  SANDBOX_DUAL_WRITE_COLUMNS,
  SANDBOX_INERT_AFTER_CUTOVER,
  SANDBOX_LIFECYCLE_COLUMNS,
} from '../constants/sandbox-columns'

@Injectable()
export class SandboxRepository extends BaseRepository<Sandbox> {
  private readonly logger = new Logger(SandboxRepository.name)

  constructor(
    @InjectDataSource() dataSource: DataSource,
    eventEmitter: EventEmitter2,
    private readonly sandboxLookupCacheInvalidationService: SandboxLookupCacheInvalidationService,
    private readonly lifecycleMigrationService: SandboxLifecycleMigrationService,
  ) {
    super(dataSource, eventEmitter, Sandbox)
  }

  async insert(sandbox: Sandbox, parentId?: string): Promise<Sandbox> {
    const now = new Date()
    if (!sandbox.createdAt) {
      sandbox.createdAt = now
    }
    if (!sandbox.updatedAt) {
      sandbox.updatedAt = now
    }

    sandbox.assertValid()
    sandbox.enforceInvariants()

    await this.dataSource.transaction(async (entityManager) => {
      await entityManager.insert(Sandbox, sandbox)
      if (this.lifecycleMigrationService.useLifecycleTableForWrites()) {
        await entityManager.insert(SandboxLifecycle, SandboxLifecycle.fromSandbox(sandbox))
      }
      await this.upsertLastActivity(entityManager, sandbox.id, sandbox.createdAt)
      sandbox.lastActivityAt = { sandboxId: sandbox.id, lastActivityAt: sandbox.createdAt }

      if (parentId) {
        await entityManager.insert(SandboxFork, {
          parentId,
          childId: sandbox.id,
        })
      }
    })

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
   * Optimistic update against a fixed predicate.
   *
   * Use this when the caller has already validated transition semantics on a
   * loaded entity and wants "no concurrent modifier" guarantees.
   *
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

    if (raw) {
      if (this.lifecycleMigrationService.useLifecycleTableForWrites()) {
        // Write to both tables in a single transaction.
        const { configFields, lifecycleFields } = this.splitUpdateFields(updateData)
        await this.dataSource.transaction(async (em) => {
          if (Object.keys(configFields).length > 0) {
            await em.update(Sandbox, id, configFields)
          }
          if (Object.keys(lifecycleFields).length > 0) {
            lifecycleFields.updatedAt = new Date()
            await em.update(SandboxLifecycle, { sandboxId: id }, lifecycleFields)
          }
        })
        return
      } else {
        // Legacy write path - trigger keeps the lifecycle table in sync.
        await this.repository.update(id, updateData)
        return
      }
    }

    const sandbox = entity ?? (await this.findOneBy({ id }))
    if (!sandbox) {
      throw new NotFoundException('Sandbox not found')
    }

    const previousSandbox = { ...sandbox }

    Object.assign(sandbox, updateData)
    sandbox.assertValid()
    const invariantChanges = sandbox.enforceInvariants()
    const fullUpdate = { ...updateData, ...invariantChanges }

    await this.dataSource.transaction(async (entityManager) => {
      if (this.lifecycleMigrationService.useLifecycleTableForWrites()) {
        // Write to both tables in a single transaction.
        const { configFields, lifecycleFields } = this.splitUpdateFields(fullUpdate)
        const { configPredicate, lifecyclePredicate } = this.splitPredicateFields({
          state: previousSandbox.state,
          desiredState: previousSandbox.desiredState,
          pending: previousSandbox.pending,
          organizationId: previousSandbox.organizationId,
        })

        // Lifecycle UPDATE must always run — it carries the OCC gate.
        // Bumping "updatedAt" manually also keeps the SET clause non-empty when only config columns change.
        lifecycleFields.updatedAt = new Date()
        const lifecycleResult = await entityManager.update(
          SandboxLifecycle,
          {
            sandboxId: previousSandbox.id,
            lifecyclePhase: SandboxLifecycle.phaseFor(previousSandbox.state),
            ...lifecyclePredicate,
          },
          lifecycleFields,
        )
        if (!lifecycleResult.affected) {
          throw new SandboxConflictError()
        }

        if (Object.keys(configFields).length > 0) {
          const configResult = await entityManager.update(
            Sandbox,
            { id: previousSandbox.id, ...configPredicate },
            configFields,
          )
          if (!configResult.affected) {
            throw new SandboxConflictError()
          }
        }
      } else {
        // Legacy write path - trigger keeps the lifecycle table in sync.
        const result = await entityManager.update(
          Sandbox,
          {
            id: previousSandbox.id,
            state: previousSandbox.state,
            desiredState: previousSandbox.desiredState,
            pending: previousSandbox.pending,
            organizationId: previousSandbox.organizationId,
          },
          fullUpdate,
        )
        if (!result.affected) {
          throw new SandboxConflictError()
        }
      }
      sandbox.updatedAt = new Date()

      if (previousSandbox.state !== sandbox.state || previousSandbox.organizationId !== sandbox.organizationId) {
        await this.upsertLastActivity(entityManager, id, sandbox.updatedAt)
        sandbox.lastActivityAt = { sandboxId: id, lastActivityAt: sandbox.updatedAt }
      }
    })

    this.emitUpdateEvents(sandbox, previousSandbox)
    this.invalidateLookupCacheOnUpdate(sandbox, previousSandbox)

    return sandbox
  }

  /**
   * Optimistic update against a caller-specified predicate.
   *
   * Use this when the caller's correctness depends on a subset of fields
   * and any other field is allowed to change concurrently.
   *
   * @param id - The ID of the sandbox to update.
   * @param params.updateData - The partial data to update.
   * @param params.whereCondition - The where condition to use for the update.
   *
   * @throws {SandboxConflictError} if the sandbox was modified by another operation
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
      const useLifecycleTableForWrites = this.lifecycleMigrationService.useLifecycleTableForWrites()

      let sandbox: Sandbox | null
      if (useLifecycleTableForWrites) {
        // SELECT joins sandbox_lifecycle and routes the caller's state-machine
        // predicate fields to `lifecycle.*` because writes route directly to
        // `sandbox_lifecycle`, leaving `sandbox.*` stale.
        const { configPredicate, lifecyclePredicate } = this.splitPredicateFields(whereCondition)
        const result = await entityManager.findOne(Sandbox, {
          where: {
            ...configPredicate,
            id,
            lifecycle: lifecyclePredicate as FindOptionsWhere<SandboxLifecycle>,
          },
          relations: { lifecycle: true },
          loadEagerRelations: false,
        })
        sandbox = result ? this.hydrate(result) : null
      } else {
        // Legacy read path — all writes still go to `sandbox`.
        sandbox = await entityManager.findOne(Sandbox, {
          where: { ...whereCondition, id },
          relations: [],
          loadEagerRelations: false,
        })
      }

      if (!sandbox) {
        throw new SandboxConflictError()
      }

      const previousSandbox = { ...sandbox }

      Object.assign(sandbox, updateData)
      sandbox.assertValid()
      const invariantChanges = sandbox.enforceInvariants()
      const fullUpdate = { ...updateData, ...invariantChanges }

      if (useLifecycleTableForWrites) {
        // Write to both tables in a single transaction.
        const { configFields, lifecycleFields } = this.splitUpdateFields(fullUpdate)
        const { configPredicate, lifecyclePredicate } = this.splitPredicateFields(whereCondition)

        // Lifecycle UPDATE must always run — it carries the OCC gate.
        // Bumping "updatedAt" manually also keeps the SET clause non-empty when only config columns change.
        lifecycleFields.updatedAt = new Date()
        const lifecycleResult = await entityManager.update(
          SandboxLifecycle,
          {
            sandboxId: previousSandbox.id,
            lifecyclePhase: SandboxLifecycle.phaseFor(previousSandbox.state),
            ...lifecyclePredicate,
          },
          lifecycleFields,
        )
        if (!lifecycleResult.affected) {
          throw new SandboxConflictError()
        }

        if (Object.keys(configFields).length > 0) {
          const configResult = await entityManager.update(
            Sandbox,
            { id: previousSandbox.id, ...configPredicate },
            configFields,
          )
          if (!configResult.affected) {
            throw new SandboxConflictError()
          }
        }
      } else {
        // Legacy write path - trigger keeps the lifecycle table in sync.
        const result = await entityManager.update(Sandbox, { ...whereCondition, id }, fullUpdate)
        if (!result.affected) {
          throw new SandboxConflictError()
        }
      }
      sandbox.updatedAt = new Date()

      if (previousSandbox.state !== sandbox.state || previousSandbox.organizationId !== sandbox.organizationId) {
        await this.upsertLastActivity(entityManager, id, sandbox.updatedAt)
        sandbox.lastActivityAt = { sandboxId: id, lastActivityAt: sandbox.updatedAt }
      }

      this.emitUpdateEvents(sandbox, previousSandbox)
      this.invalidateLookupCacheOnUpdate(sandbox, previousSandbox)

      return sandbox
    })
  }

  /**
   * Splits an update into config-table columns and
   * lifecycle-table columns. Dual-write columns (currently only `organizationId`)
   * appear on both sides.
   *
   * If `state` is being updated, the derived `lifecyclePhase` is also written
   * to the lifecycle side so the row migrates between `sandbox_lifecycle_active`
   * and `sandbox_lifecycle_terminal` automatically.
   */
  private splitUpdateFields(updateData: Partial<Sandbox>): {
    configFields: Record<string, unknown>
    lifecycleFields: Record<string, unknown>
  } {
    const configFields: Record<string, unknown> = {}
    const lifecycleFields: Record<string, unknown> = {}

    for (const [key, value] of Object.entries(updateData)) {
      if (key === 'id') continue
      if ((SANDBOX_DUAL_WRITE_COLUMNS as readonly string[]).includes(key)) {
        configFields[key] = value
        lifecycleFields[key] = value
      } else if ((SANDBOX_CONFIG_COLUMNS as readonly string[]).includes(key)) {
        configFields[key] = value
      } else if ((SANDBOX_LIFECYCLE_COLUMNS as readonly string[]).includes(key)) {
        lifecycleFields[key] = value
      }
    }

    if (updateData.state !== undefined) {
      lifecycleFields.lifecyclePhase = SandboxLifecycle.phaseFor(updateData.state)
    }

    return { configFields, lifecycleFields }
  }

  /**
   * Splits an optimistic-concurrency predicate into the fields that
   * apply against `sandbox` vs. those that apply against `sandbox_lifecycle`.
   * `id` is intentionally dropped — callers add `id` (or `sandboxId`) to the
   * relevant predicate themselves so it always lands in the right column.
   */
  private splitPredicateFields(predicate: FindOptionsWhere<Sandbox>): {
    configPredicate: Record<string, unknown>
    lifecyclePredicate: Record<string, unknown>
  } {
    const configPredicate: Record<string, unknown> = {}
    const lifecyclePredicate: Record<string, unknown> = {}

    for (const [key, value] of Object.entries(predicate)) {
      if (key === 'id') continue
      if ((SANDBOX_DUAL_WRITE_COLUMNS as readonly string[]).includes(key)) {
        configPredicate[key] = value
        lifecyclePredicate[key] = value
      } else if ((SANDBOX_CONFIG_COLUMNS as readonly string[]).includes(key)) {
        configPredicate[key] = value
      } else if ((SANDBOX_LIFECYCLE_COLUMNS as readonly string[]).includes(key)) {
        lifecyclePredicate[key] = value
      }
    }

    return { configPredicate, lifecyclePredicate }
  }

  override async find(options?: FindManyOptions<Sandbox>): Promise<Sandbox[]> {
    const effectiveOptions = this.withLifecycleRelation(options)
    const results = await super.find(effectiveOptions)
    return this.hydrateAll(results)
  }

  override async findAndCount(options?: FindManyOptions<Sandbox>): Promise<[Sandbox[], number]> {
    const effectiveOptions = this.withLifecycleRelation(options)
    const [results, count] = await super.findAndCount(effectiveOptions)
    return [this.hydrateAll(results), count]
  }

  override async findOne(options: FindOneOptions<Sandbox>): Promise<Sandbox | null> {
    const effectiveOptions = this.withLifecycleRelation(options)
    const result = await super.findOne(effectiveOptions)
    return result ? this.hydrate(result) : null
  }

  override async findOneOrFail(options: FindOneOptions<Sandbox>): Promise<Sandbox> {
    const result = await this.findOne(options)
    if (result === null) {
      throw new EntityNotFoundError(Sandbox, options.where ?? {})
    }
    return result
  }

  override async findOneBy(where: FindOptionsWhere<Sandbox> | FindOptionsWhere<Sandbox>[]): Promise<Sandbox | null> {
    return this.findOne({ where })
  }

  override async findOneByOrFail(where: FindOptionsWhere<Sandbox> | FindOptionsWhere<Sandbox>[]): Promise<Sandbox> {
    const result = await this.findOneBy(where)
    if (result === null) {
      throw new EntityNotFoundError(Sandbox, where)
    }
    return result
  }

  /**
   * Augments `FindOptions` to eager-load the `sandbox_lifecycle` relation so
   * {@link hydrate} can overlay its columns onto the returned sandbox.
   *
   * No-op when {@link SandboxLifecycleMigrationService.useLifecycleTableForReads}
   * is off, or when the caller passed a narrow `select` (forcing the
   * relation would widen the projection against caller intent).
   */
  private withLifecycleRelation<T extends FindOneOptions<Sandbox> | FindManyOptions<Sandbox> | undefined>(
    options: T,
  ): T {
    if (!this.lifecycleMigrationService.useLifecycleTableForReads()) {
      return options
    }
    if (options?.select) {
      return options
    }

    const existing = options?.relations
    let relations: FindOptionsRelations<Sandbox>
    if (Array.isArray(existing)) {
      relations = {}
      for (const r of existing) {
        ;(relations as Record<string, boolean>)[r] = true
      }
      relations.lifecycle = true
    } else {
      relations = { ...(existing ?? {}), lifecycle: true }
    }

    return { ...(options ?? {}), relations } as T
  }

  /**
   * Batched {@link hydrate} for many-result reads. Short-circuits at the top
   * to avoid a per-row flag check on large result sets.
   */
  hydrateAll(sandboxes: Sandbox[]): Sandbox[] {
    if (!this.lifecycleMigrationService.useLifecycleTableForReads()) {
      return sandboxes
    }
    for (const sandbox of sandboxes) {
      this.hydrate(sandbox)
    }
    return sandboxes
  }

  /**
   * Overlays `sandbox_lifecycle` columns onto the in-memory sandbox so
   * callers reading `sandbox.state`/`sandbox.pending`/etc. see fresh values
   * regardless of which table holds the source of truth.
   *
   * No-op when the read shim is off or the `lifecycle` relation is missing.
   * Mutates in-place; returns the same instance.
   */
  hydrate(sandbox: Sandbox): Sandbox {
    if (!this.lifecycleMigrationService.useLifecycleTableForReads()) {
      return sandbox
    }
    const lifecycle = sandbox.lifecycle
    if (!lifecycle) {
      return sandbox
    }
    for (const col of SANDBOX_INERT_AFTER_CUTOVER) {
      ;(sandbox as unknown as Record<string, unknown>)[col] = (lifecycle as unknown as Record<string, unknown>)[col]
    }
    return sandbox
  }

  /**
   * Upserts the last activity for a sandbox.
   */
  private async upsertLastActivity(
    entityManager: EntityManager,
    sandboxId: string,
    lastActivityAt: Date,
  ): Promise<void> {
    await entityManager.upsert(SandboxLastActivity, { sandboxId, lastActivityAt }, ['sandboxId'])
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
