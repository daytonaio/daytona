/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logger } from '@nestjs/common'
import { AuditLogStorageAdapter } from '../interfaces/audit-storage.interface'
import { InjectRepository } from '@nestjs/typeorm'
import { AuditLog } from '../entities/audit-log.entity'
import { FindManyOptions, Repository } from 'typeorm'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'
import { AuditLogFilter } from '../interfaces/audit-filter.interface'
import { createRangeFilter } from '../../common/utils/range-filter'

export class AuditTypeormStorageAdapter implements AuditLogStorageAdapter {
  private readonly logger = new Logger(AuditTypeormStorageAdapter.name)

  constructor(
    @InjectRepository(AuditLog)
    private readonly auditLogRepository: Repository<AuditLog>,
  ) {}

  async write(auditLogs: AuditLog[]): Promise<void> {
    throw new Error('Typeorm adapter does not support writing audit logs.')
  }

  async getAllLogs(page?: number, limit = 1000, filters?: AuditLogFilter): Promise<PaginatedList<AuditLog>> {
    const options: FindManyOptions<AuditLog> = {
      order: {
        createdAt: 'DESC',
      },
      skip: (page - 1) * limit,
      take: limit,
      where: {
        createdAt: createRangeFilter(filters?.from, filters?.to),
      },
    }

    const [items, total] = await this.auditLogRepository.findAndCount(options)

    return {
      items,
      total,
      page: page,
      totalPages: Math.ceil(total / limit),
    }
  }

  async getOrganizationLogs(
    organizationId: string,
    page?: number,
    limit = 1000,
    filters?: AuditLogFilter,
  ): Promise<PaginatedList<AuditLog>> {
    const options: FindManyOptions<AuditLog> = {
      order: {
        createdAt: 'DESC',
      },
      skip: (page - 1) * limit,
      take: limit,
      where: [
        {
          organizationId,
          createdAt: createRangeFilter(filters?.from, filters?.to),
        },
      ],
    }

    const [items, total] = await this.auditLogRepository.findAndCount(options)

    return {
      items,
      total,
      page: page,
      totalPages: Math.ceil(total / limit),
    }
  }
}
