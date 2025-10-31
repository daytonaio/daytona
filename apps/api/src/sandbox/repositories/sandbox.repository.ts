/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Repository, DataSource, FindOptionsWhere } from 'typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { ConflictException, Injectable } from '@nestjs/common'
import { InjectDataSource } from '@nestjs/typeorm'

@Injectable()
export class SandboxRepository extends Repository<Sandbox> {
  constructor(@InjectDataSource() private dataSource: DataSource) {
    super(Sandbox, dataSource.createEntityManager(), dataSource.createQueryRunner())
  }

  async saveWhere(sandbox: Sandbox, whereCondition: FindOptionsWhere<Sandbox>): Promise<Sandbox | null> {
    return this.manager.transaction(async (entityManager) => {
      const whereClause = {
        ...whereCondition,
        id: sandbox.id,
      }

      const existingSandbox = await entityManager.findOne(Sandbox, {
        where: whereClause,
        lock: { mode: 'pessimistic_write' },
        relations: [],
        loadEagerRelations: false,
      })

      if (!existingSandbox) {
        throw new ConflictException('Sandbox was modified by another operation, please try again')
      }

      const mergedEntity = entityManager.merge(Sandbox, existingSandbox, sandbox)
      const savedEntity = await entityManager.save(mergedEntity)

      return savedEntity
    })
  }
}
