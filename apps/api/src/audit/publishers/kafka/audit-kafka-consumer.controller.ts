/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Controller, Inject, Logger, UseFilters } from '@nestjs/common'
import { Ctx, EventPattern, KafkaContext, Payload } from '@nestjs/microservices'
import { AuditLog } from '../../entities/audit-log.entity'
import { type AuditLogStorageAdapter } from '../../interfaces/audit-storage.interface'
import { AutoCommitOffset } from '../../../common/decorators/autocommit-offset.decorator'
import { AUDIT_KAFKA_TOPIC, AUDIT_STORAGE_ADAPTER } from '../../constants/audit-tokens'
import { KafkaMaxRetryExceptionFilter } from '../../../filters/kafka-exception.filter'

@Controller('kafka-audit')
@UseFilters(new KafkaMaxRetryExceptionFilter({ retries: 3, sendToDlq: true }))
export class AuditKafkaConsumerController {
  private readonly logger = new Logger(AuditKafkaConsumerController.name)

  constructor(@Inject(AUDIT_STORAGE_ADAPTER) private readonly auditStorageAdapter: AuditLogStorageAdapter) {}

  @EventPattern(AUDIT_KAFKA_TOPIC)
  @AutoCommitOffset()
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public async handleAuditLogMessage(@Payload() message: AuditLog, @Ctx() _: KafkaContext): Promise<void> {
    this.logger.debug('Handling audit log message', { message })
    await this.auditStorageAdapter.write([message])
  }
}
