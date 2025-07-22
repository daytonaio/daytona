/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { LessThan, Repository } from 'typeorm'
import { CreateAuditLogInternalDto } from '../dto/create-audit-log-internal.dto'
import { UpdateAuditLogInternalDto } from '../dto/update-audit-log-internal.dto'
import { AuditLog } from '../entities/audit-log.entity'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'
import { TypedConfigService } from '../../config/typed-config.service'
import { RedisLockProvider } from '../../sandbox/common/redis-lock.provider'

@Injectable()
export class AuditService {
  private readonly logger = new Logger(AuditService.name)

  constructor(
    @InjectRepository(AuditLog)
    private readonly auditLogRepository: Repository<AuditLog>,
    private readonly configService: TypedConfigService,
    private readonly redisLockProvider: RedisLockProvider,
  ) {}

  async createLog(createDto: CreateAuditLogInternalDto): Promise<AuditLog> {
    const auditLog = new AuditLog()
    auditLog.actorId = createDto.actorId
    auditLog.actorEmail = createDto.actorEmail
    auditLog.organizationId = createDto.organizationId
    auditLog.action = createDto.action
    auditLog.targetType = createDto.targetType
    auditLog.targetId = createDto.targetId
    auditLog.statusCode = createDto.statusCode
    auditLog.errorMessage = createDto.errorMessage
    auditLog.ipAddress = createDto.ipAddress
    auditLog.userAgent = createDto.userAgent
    auditLog.source = createDto.source
    auditLog.metadata = createDto.metadata

    return await this.auditLogRepository.save(auditLog)
  }

  async updateLog(id: string, updateDto: UpdateAuditLogInternalDto): Promise<AuditLog> {
    const auditLog = await this.auditLogRepository.findOne({ where: { id } })
    if (!auditLog) {
      throw new NotFoundException(`Audit log with ID ${id} not found`)
    }

    Object.assign(auditLog, updateDto)
    return await this.auditLogRepository.save(auditLog)
  }

  async getAllLogs(page = 1, limit = 10): Promise<PaginatedList<AuditLog>> {
    const pageNum = Number(page)
    const limitNum = Number(limit)

    const [items, total] = await this.auditLogRepository.findAndCount({
      order: {
        createdAt: 'DESC',
      },
      skip: (pageNum - 1) * limitNum,
      take: limitNum,
    })

    return {
      items,
      total,
      page: pageNum,
      totalPages: Math.ceil(total / limitNum),
    }
  }

  async getLogsByOrganization(organizationId: string, page = 1, limit = 10): Promise<PaginatedList<AuditLog>> {
    const pageNum = Number(page)
    const limitNum = Number(limit)

    const [items, total] = await this.auditLogRepository.findAndCount({
      where: [{ organizationId }, { targetId: organizationId }],
      order: {
        createdAt: 'DESC',
      },
      skip: (pageNum - 1) * limitNum,
      take: limitNum,
    })

    return {
      items,
      total,
      page: pageNum,
      totalPages: Math.ceil(total / limitNum),
    }
  }

  @Cron(CronExpression.EVERY_DAY_AT_2AM)
  async cleanupOldAuditLogs(): Promise<void> {
    const lockKey = 'cleanup-old-audit-logs'
    if (!(await this.redisLockProvider.lock(lockKey, 600))) {
      return
    }

    try {
      const retentionDays = this.configService.get('audit.retentionDays')

      if (!retentionDays) {
        this.logger.debug('Audit log retention not configured, skipping cleanup')
        return
      }

      if (retentionDays < 90) {
        this.logger.warn(
          `Audit log retention period (${retentionDays} days) is less than minimum 90 days, skipping cleanup`,
        )
        return
      }

      const cutoffDate = new Date(Date.now() - retentionDays * 24 * 60 * 60 * 1000)

      let totalDeleted = 0
      const batchSize = 1000

      this.logger.log(`Starting cleanup of audit logs older than ${retentionDays} days`)

      while (true) {
        // Find batch of audit logs older than the retention period
        const logsToDelete = await this.auditLogRepository.find({
          where: {
            createdAt: LessThan(cutoffDate),
          },
          take: batchSize,
        })

        if (logsToDelete.length === 0) {
          break
        }

        const idsToDelete = logsToDelete.map((log) => log.id)

        // Delete batch
        const result = await this.auditLogRepository.delete(idsToDelete)
        const deletedCount = result.affected || 0
        totalDeleted += deletedCount

        // If we deleted fewer records than the batch size, we're done
        if (deletedCount < batchSize) {
          break
        }
      }

      this.logger.log(`Completed cleanup of audit logs older than ${retentionDays} days (${totalDeleted} logs deleted)`)
    } catch (error) {
      this.logger.error(`An error occurred during cleanup of old audit logs: ${error.message}`, error.stack)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }
}
