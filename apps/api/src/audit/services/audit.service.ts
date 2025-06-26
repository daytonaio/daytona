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

  @Cron(CronExpression.EVERY_DAY_AT_2AM)
  async cleanupOldAuditLogs(): Promise<void> {
    const lockKey = 'cleanup-old-audit-logs'
    if (!(await this.redisLockProvider.lock(lockKey, 300))) {
      return
    }

    try {
      const retentionDays = this.configService.get('audit.retentionDays')

      if (!retentionDays || retentionDays <= 0) {
        this.logger.debug('Audit log retention not configured, skipping cleanup')
        return
      }

      if (retentionDays < 90) {
        this.logger.warn(
          `Audit log retention period (${retentionDays} days) is less than minimum 90 days, skipping cleanup`,
        )
        return
      }

      const cutoffDate = new Date(Date.now() - retentionDays * 1000 * 60 * 60 * 24)

      this.logger.log(
        `Starting cleanup of audit logs older than ${retentionDays} days (before ${cutoffDate.toISOString()})`,
      )

      let totalDeleted = 0
      const batchSize = 1000

      // Delete in batches to avoid locking the table for too long
      while (true) {
        const deletionResult = await this.auditLogRepository.delete({
          createdAt: LessThan(cutoffDate),
        })

        const deletedCount = deletionResult.affected || 0
        totalDeleted += deletedCount

        this.logger.debug(`Deleted ${deletedCount} audit logs in current batch`)

        // If we deleted fewer records than the batch size, we're done
        if (deletedCount < batchSize) {
          break
        }

        // Small delay between batches to reduce database load
        await new Promise((resolve) => setTimeout(resolve, 100))
      }

      if (totalDeleted > 0) {
        this.logger.log(
          `Cleanup completed: deleted ${totalDeleted} audit logs older than ${retentionDays} days (before ${cutoffDate.toISOString()})`,
        )
      } else {
        this.logger.log(`No old audit logs found for cleanup (before ${cutoffDate.toISOString()})`)
      }
    } catch (error) {
      this.logger.error(`Failed to cleanup old audit logs: ${error.message}`, error.stack)
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }
}
