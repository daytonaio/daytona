/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException, OnModuleInit } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { CronExpression, SchedulerRegistry } from '@nestjs/schedule'
import { CronJob } from 'cron'
import { FindManyOptions, LessThan, Repository } from 'typeorm'
import { CreateAuditLogInternalDto } from '../dto/create-audit-log-internal.dto'
import { UpdateAuditLogInternalDto } from '../dto/update-audit-log-internal.dto'
import { AuditLog } from '../entities/audit-log.entity'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'
import { TypedConfigService } from '../../config/typed-config.service'
import { RedisLockProvider } from '../../sandbox/common/redis-lock.provider'

@Injectable()
export class AuditService implements OnModuleInit {
  private readonly logger = new Logger(AuditService.name)

  constructor(
    @InjectRepository(AuditLog)
    private readonly auditLogRepository: Repository<AuditLog>,
    private readonly configService: TypedConfigService,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly schedulerRegistry: SchedulerRegistry,
  ) {}

  onModuleInit() {
    this.schedulerRegistry.addCronJob(
      'cleanup-old-audit-logs',
      new CronJob(
        CronExpression.EVERY_DAY_AT_2AM,
        () => {
          this.cleanupOldAuditLogs()
        },
        null,
        true,
        this.configService.get('cronTimeZone'),
      ),
    )
  }

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

    if (this.configService.get('audit.consoleLogEnabled')) {
      this.logger.log(`Creating audit log: ${JSON.stringify(auditLog)}`)
    }

    return await this.auditLogRepository.save(auditLog)
  }

  async updateLog(id: string, updateDto: UpdateAuditLogInternalDto): Promise<AuditLog> {
    const auditLog = await this.auditLogRepository.findOne({ where: { id } })
    if (!auditLog) {
      throw new NotFoundException(`Audit log with ID ${id} not found`)
    }

    if (updateDto.statusCode) {
      auditLog.statusCode = updateDto.statusCode
    }

    if (updateDto.errorMessage) {
      auditLog.errorMessage = updateDto.errorMessage
    }

    if (updateDto.targetId) {
      auditLog.targetId = updateDto.targetId
    }

    if (updateDto.organizationId) {
      auditLog.organizationId = updateDto.organizationId
    }

    if (this.configService.get('audit.consoleLogEnabled')) {
      this.logger.log(`Updating audit log: ${JSON.stringify(auditLog)}`)
    }

    return await this.auditLogRepository.save(auditLog)
  }

  async getLogs(page = 1, limit = 10, organizationId?: string): Promise<PaginatedList<AuditLog>> {
    const pageNum = Number(page)
    const limitNum = Number(limit)

    const options: FindManyOptions<AuditLog> = {
      order: {
        createdAt: 'DESC',
      },
      skip: (pageNum - 1) * limitNum,
      take: limitNum,
    }

    if (organizationId) {
      options.where = [{ organizationId }, { targetId: organizationId }]
    }

    const [items, total] = await this.auditLogRepository.findAndCount(options)

    return {
      items,
      total,
      page: pageNum,
      totalPages: Math.ceil(total / limitNum),
    }
  }

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
          order: {
            createdAt: 'ASC',
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
