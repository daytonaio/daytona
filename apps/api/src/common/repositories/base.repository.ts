/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Repository, DataSource, FindOptionsWhere, FindOneOptions, FindManyOptions, EntityTarget } from 'typeorm'
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
   * Inserts a new entity into the database and emits a corresponding event.
   *
   * Must use {@link Repository.insert} to insert the entity into the database.
   */
  protected abstract insert(entity: TEntity): Promise<TEntity>

  /**
   * Partially updates an entity in the database and emits a corresponding event based on the changes.
   *
   * Must use {@link Repository.update} to update the entity in the database.
   */
  protected abstract update(entity: TEntity, updateData: Partial<TEntity>): Promise<TEntity>
}
