/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Repository,
  DataSource,
  FindOptionsWhere,
  InsertResult,
  FindOneOptions,
  FindManyOptions,
  EntityTarget,
} from 'typeorm'
import { ObjectLiteral } from 'typeorm/common/ObjectLiteral'
import { EventEmitter2 } from '@nestjs/event-emitter'

/**
 * Abstract base repository class that provides common CRUD operations with event emission.
 *
 * @template TEntity - The entity class this repository manages
 */
export abstract class BaseRepository<TEntity extends ObjectLiteral> {
  protected repository: Repository<TEntity>

  constructor(
    protected readonly dataSource: DataSource,
    protected readonly eventEmitter: EventEmitter2,
    protected readonly entityClass: EntityTarget<TEntity>,
  ) {
    this.repository = this.dataSource.getRepository(entityClass)
  }

  /**
   * See reference for {@link Repository.findOne}
   */
  async findOne(options: FindOneOptions<TEntity>): Promise<TEntity | null> {
    return this.repository.findOne(options)
  }

  /**
   * See reference for {@link Repository.findOneBy}
   */
  async findOneBy(where: FindOptionsWhere<TEntity> | FindOptionsWhere<TEntity>[]): Promise<TEntity | null> {
    return this.repository.findOneBy(where)
  }

  /**
   * See reference for {@link Repository.findOneByOrFail}
   */
  async findOneByOrFail(where: FindOptionsWhere<TEntity> | FindOptionsWhere<TEntity>[]): Promise<TEntity> {
    return this.repository.findOneByOrFail(where)
  }

  /**
   * See reference for {@link Repository.find}
   */
  async find(options?: FindManyOptions<TEntity>): Promise<TEntity[]> {
    return this.repository.find(options)
  }

  /**
   * See reference for {@link Repository.findAndCount}
   */
  async findAndCount(options?: FindManyOptions<TEntity>): Promise<[TEntity[], number]> {
    return this.repository.findAndCount(options)
  }

  /**
   * See reference for {@link Repository.count}
   */
  async count(options?: FindManyOptions<TEntity>): Promise<number> {
    return this.repository.count(options)
  }

  /**
   * See reference for {@link Repository.manager}
   */
  get manager() {
    return this.repository.manager
  }

  /**
   * Inserts a new entity into the database.
   * After insertion, fetches the complete entity (including DB-generated fields)
   * and emits the appropriate creation event.
   *
   * @param entity - The entity to insert
   * @returns The inserted entity with all DB-generated fields populated
   */
  async insert(entity: TEntity): Promise<TEntity> {
    const result = await this.repository.insert(entity)

    const insertedEntity = await this.findOneByIdentifier(this.extractIdentifier(result))
    if (!insertedEntity) {
      throw new Error()
    }

    this.emitCreatedEvent(insertedEntity)

    return insertedEntity
  }

  /**
   * Updates an existing entity with partial data.
   * Stores old values, performs the update, merges changes into the entity object,
   * and emits appropriate events based on what changed.
   *
   * @param entity - The current entity instance
   * @param updateData - Partial entity data containing the fields to update
   * @returns The updated entity
   */
  async update(entity: TEntity, updateData: Partial<TEntity>): Promise<TEntity> {
    const identifier = this.getEntityIdentifier(entity)

    // Capture old values for event emission
    const oldValues = this.captureOldValues(entity, updateData)

    // Perform the update
    await this.repository.update(identifier, updateData as any)

    // Merge changes into the entity (avoid extra DB query)
    Object.assign(entity, updateData)

    // Emit events based on what changed
    this.emitUpdateEvents(entity, oldValues, updateData)

    return entity
  }

  /**
   * Abstract method to find an entity by its identifier.
   * Must be implemented by child classes to handle their specific identifier structure.
   *
   * @param identifier - The entity identifier
   */
  protected abstract findOneByIdentifier(identifier: FindOptionsWhere<TEntity>): Promise<TEntity | null>

  /**
   * Abstract method to extract the identifier from an insert result.
   * Must be implemented by child classes based on their primary key field(s).
   *
   * @param result - The insert result from TypeORM
   */
  protected abstract extractIdentifier(result: InsertResult): FindOptionsWhere<TEntity>

  /**
   * Abstract method to get the identifier from an entity instance.
   * Used for updates to identify which record to update.
   *
   * @param entity - The entity instance
   */
  protected abstract getEntityIdentifier(entity: TEntity): FindOptionsWhere<TEntity>

  /**
   * Abstract method to emit the created event.
   * Must be implemented by child classes to emit their specific event type.
   *
   * @param entity - The newly created entity
   */
  protected abstract emitCreatedEvent(entity: TEntity): void

  /**
   * Abstract method to capture old values before update.
   * Child classes should override to capture specific fields they track for events.
   *
   * @param entity - The current entity
   * @param updateData - The data that will be updated
   * @returns An object containing the old values that need to be tracked
   */
  protected abstract captureOldValues(entity: TEntity, updateData: Partial<TEntity>): Record<string, any>

  /**
   * Abstract method to emit update events based on what changed.
   * Child classes should implement this to emit specific events for field changes they care about.
   *
   * @param entity - The updated entity
   * @param oldValues - The captured old values before the update
   * @param updateData - The data that was updated
   */
  protected abstract emitUpdateEvents(
    entity: TEntity,
    oldValues: Record<string, any>,
    updateData: Partial<TEntity>,
  ): void
}
