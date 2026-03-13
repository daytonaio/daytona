/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  ForbiddenException,
  Inject,
  Injectable,
  Logger,
  NotFoundException,
  OnApplicationBootstrap,
  Optional,
} from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression, SchedulerRegistry } from '@nestjs/schedule'
import { LessThan, Repository, IsNull, Not, QueryFailedError } from 'typeorm'
import { CreateAuditLogInternalDto } from '../dto/create-audit-log-internal.dto'
import { AuditLog } from '../entities/audit-log.entity'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'
import { TypedConfigService } from '../../config/typed-config.service'
import { RedisLockProvider } from '../../sandbox/common/redis-lock.provider'
import { AUDIT_LOG_PUBLISHER, AUDIT_STORAGE_ADAPTER } from '../constants/audit-tokens'
import { AuditLogStorageAdapter } from '../interfaces/audit-storage.interface'
import { AuditLogPublisher } from '../interfaces/audit-publisher.interface'
import { AuditLogFilter } from '../interfaces/audit-filter.interface'
import { DistributedLock } from '../../common/decorators/distributed-lock.decorator'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'
import { LogExecution } from '../../common/decorators/log-execution.decorator'

@Injectable()
export class AuditService implements OnApplicationBootstrap {
  private readonly logger = new Logger(AuditService.name)

  constructor(
    @InjectRepository(AuditLog)
    private readonly auditLogRepository: Repository<AuditLog>,
    private readonly configService: TypedConfigService,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly schedulerRegistry: SchedulerRegistry,
    @Inject(AUDIT_STORAGE_ADAPTER)
    private readonly auditStorageAdapter: AuditLogStorageAdapter,
    @Optional()
    @Inject(AUDIT_LOG_PUBLISHER)
    private readonly auditLogPublisher?: AuditLogPublisher,
  ) {}

  onApplicationBootstrap() {
    const auditConfig = this.configService.get('audit')

    // Enable publish cron job if publish is enabled
    if (auditConfig.publish.enabled) {
      this.schedulerRegistry.getCronJob('publish-audit-logs').start()
      return
    }

    // Enable cleanup cron job if retention days is configured and publish is disabled
    if (auditConfig.retentionDays && auditConfig.retentionDays > 0) {
      this.schedulerRegistry.getCronJob('cleanup-old-audit-logs').start()
    }

    const batchSize = this.configService.getOrThrow('audit.publish.batchSize')

    if (batchSize > 50000) {
      throw new Error('Audit publish batch size cannot be greater than 50000')
    }
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

    return this.auditLogRepository.save(auditLog, { transaction: false })
  }

  /**
   * Updates a pending audit log with request handler outcome data.
   * Once finalized (statusCode is set), the log becomes immutable (enforced by database trigger).
   *
   * @param auditLog - The audit log to update.
   * @param updateData - The data to update the audit log with.
   * @throws {NotFoundException} - If the audit log is not found.
   * @throws {ForbiddenException} - If the audit log is already finalized.
   *
   */
  async updateLog(auditLog: AuditLog, updateData: Partial<AuditLog>): Promise<void> {
    try {
      const result = await this.auditLogRepository.update(auditLog.id, updateData)

      if (result.affected === 0) {
        throw new NotFoundException(`Audit log with ID ${auditLog.id} not found`)
      }
    } catch (error) {
      if (error instanceof QueryFailedError && error.message.includes('immutable')) {
        throw new ForbiddenException('Finalized audit logs are immutable.')
      }
      throw error
    }

    if (this.configService.get('audit.consoleLogEnabled')) {
      this.logger.log(`AUDIT_ENTRY: ${JSON.stringify({ ...auditLog, ...updateData })}`)
    }
  }

  async getAllLogs(
    page = 1,
    limit = 10,
    filters?: AuditLogFilter,
    nextToken?: string,
  ): Promise<PaginatedList<AuditLog>> {
    return this.auditStorageAdapter.getAllLogs(page, limit, filters, nextToken)
  }

  async getOrganizationLogs(
    organizationId: string,
    page = 1,
    limit = 10,
    filters?: AuditLogFilter,
    nextToken?: string,
  ): Promise<PaginatedList<AuditLog>> {
    return this.auditStorageAdapter.getOrganizationLogs(organizationId, page, limit, filters, nextToken)
  }

  @Cron(CronExpression.EVERY_DAY_AT_2AM, {
    name: 'cleanup-old-audit-logs',
    waitForCompletion: true,
    disabled: true,
  })
  @DistributedLock()
  @LogExecution('cleanup-old-audit-logs')
  async cleanupOldAuditLogs(): Promise<void> {
    try {
      const retentionDays = this.configService.get('audit.retentionDays')
      if (!retentionDays) {
        return
      }

      const cutoffDate = new Date(Date.now() - retentionDays * 24 * 60 * 60 * 1000)
      this.logger.log(`Starting cleanup of audit logs older than ${retentionDays} days`)

      const deletedLogs = await this.auditLogRepository.delete({
        createdAt: LessThan(cutoffDate),
      })

      const totalDeleted = deletedLogs.affected

      this.logger.log(`Completed cleanup of audit logs older than ${retentionDays} days (${totalDeleted} logs deleted)`)
    } catch (error) {
      this.logger.error(`An error occurred during cleanup of old audit logs: ${error.message}`, error.stack)
    }
  }

  // Resolve dangling audit logs where status code is not set and created at is more than half an hour ago
  @Cron(CronExpression.EVERY_MINUTE, {
    name: 'resolve-dangling-audit-logs',
    waitForCompletion: true,
  })
  @DistributedLock()
  @WithInstrumentation()
  @LogExecution('resolve-dangling-audit-logs')
  async resolveDanglingLogs() {
    const danglingLogs = await this.auditLogRepository.find({
      where: {
        statusCode: IsNull(),
        createdAt: LessThan(new Date(Date.now() - 30 * 60 * 1000)),
      },
    })

    for (const log of danglingLogs) {
      // set status code to unknown
      log.statusCode = 0
      await this.auditLogRepository.save(log)
      if (this.configService.get('audit.consoleLogEnabled')) {
        this.logger.log(`AUDIT_ENTRY: ${JSON.stringify(log)}`)
      }
    }
    this.logger.debug(`Resolved ${danglingLogs.length} dangling audit logs`)
  }

  @Cron(CronExpression.EVERY_SECOND, {
    name: 'publish-audit-logs',
    waitForCompletion: true,
    disabled: true,
  })
  @LogExecution('publish-audit-logs')
  @DistributedLock()
  @WithInstrumentation()
  async publishAuditLogs() {
    // Safeguard
    if (!this.auditLogPublisher) {
      this.logger.warn('Audit log publisher not configured, skipping publish')
      return
    }

    const auditLogs = await this.auditLogRepository.find({
      where: {
        statusCode: Not(IsNull()),
      },
      take: this.configService.getOrThrow('audit.publish.batchSize'),
      order: {
        createdAt: 'ASC',
      },
    })

    if (auditLogs.length === 0) {
      return
    }

    await this.auditLogPublisher.write(auditLogs)
    await this.auditLogRepository.delete(auditLogs.map((log) => log.id))
  }
}
