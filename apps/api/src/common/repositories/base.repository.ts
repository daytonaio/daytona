/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  Repository,
  DataSource,
  FindOptionsWhere,
  FindOneOptions,
  FindManyOptions,
  EntityTarget,
  SelectQueryBuilder,
  DeleteResult,
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
   * See reference for {@link Repository.findOneOrFail}
   */
  async findOneOrFail(options: FindOneOptions<TEntity>): Promise<TEntity> {
    return this.repository.findOneOrFail(options)
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
   * Returns the entity manager for the repository. Use this only when you need to perform raw SQL queries.
   *
   * See reference for {@link Repository.manager}
   */
  get manager() {
    return this.repository.manager
  }

  /**
   * See reference for {@link Repository.createQueryBuilder}
   */
  createQueryBuilder(alias?: string): SelectQueryBuilder<TEntity> {
    return this.repository.createQueryBuilder(alias)
  }

  /**
   * See reference for {@link Repository.delete}
   */
  async delete(criteria: FindOptionsWhere<TEntity> | FindOptionsWhere<TEntity>[]): Promise<DeleteResult> {
    return this.repository.delete(criteria)
  }

  /**
   * Inserts a new entity into the database.
   *
   * Uses {@link Repository.insert} to insert the entity into the database.
   *
   * @returns The inserted entity.
   */
  abstract insert(entity: TEntity): Promise<TEntity>

  /**
   * Partially updates an entity in the database.
   *
   * Uses {@link Repository.update} to update the entity in the database.
   *
   * @param id - The ID of the entity to update.
   * @param params.updateData - The partial data to update.
   * @param params.entity - Optional pre-fetched entity to use instead of fetching from the database when not performing a raw update.
   * @param raw - If true, performs only the database update via {@link Repository.update},
   *   skipping entity fetching, domain logic (validation, derived fields), and event emission.
   *
   * @returns The updated entity or void if `raw` is true.
   */
  abstract update(id: string, params: { updateData: Partial<TEntity>; entity?: TEntity }, raw: true): Promise<void>
  abstract update(id: string, params: { updateData: Partial<TEntity>; entity?: TEntity }, raw?: false): Promise<TEntity>
  abstract update(
    id: string,
    params: { updateData: Partial<TEntity>; entity?: TEntity },
    raw?: boolean,
  ): Promise<TEntity | void>
}
