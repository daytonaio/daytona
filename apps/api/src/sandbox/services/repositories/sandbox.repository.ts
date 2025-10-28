/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Repository, DataSource, FindOptionsWhere } from 'typeorm'
import { Sandbox } from '../../entities/sandbox.entity'
import { Injectable } from '@nestjs/common'
import { InjectDataSource } from '@nestjs/typeorm'

@Injectable()
export class SandboxRepository extends Repository<Sandbox> {
  constructor(@InjectDataSource() private dataSource: DataSource) {
    super(Sandbox, dataSource.createEntityManager(), dataSource.createQueryRunner())
  }

  async saveWhere(sandbox: Sandbox, whereCondition: FindOptionsWhere<Sandbox>): Promise<Sandbox | null> {
    // Build the update query
    const queryBuilder = this.createQueryBuilder().update(Sandbox).set(sandbox).where('id = :id', { id: sandbox.id })

    // Add where conditions
    for (const [key, value] of Object.entries(whereCondition)) {
      queryBuilder.andWhere(`${key} = :${key}`, { [key]: value })
    }

    // Execute the update
    const result = await queryBuilder.execute()

    if (result.affected === 0) {
      // The update failed - either conditions weren't met or version changed
      throw new Error(`Failed to update sandbox ${sandbox.id} - entity was modified concurrently or conditions not met`)
    }

    // Fetch and return the updated entity
    const updatedEntity = await this.findOne({ where: { id: sandbox.id } })

    if (!updatedEntity) {
      //  this should never happen
      throw new Error(`Failed to update sandbox ${sandbox.id} - entity was not found after update`)
    }

    return updatedEntity
  }
}
