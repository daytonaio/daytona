/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logger, OnModuleInit } from '@nestjs/common'
import { ClientKafkaProxy } from '@nestjs/microservices'
import { CompressionTypes, Message } from 'kafkajs'
import { AuditLog } from '../../entities/audit-log.entity'
import { AuditLogPublisher } from '../../interfaces/audit-publisher.interface'
import { AUDIT_KAFKA_TOPIC } from '../../constants/audit-tokens'

export class AuditKafkaPublisher implements AuditLogPublisher, OnModuleInit {
  private readonly logger = new Logger(AuditKafkaPublisher.name)

  constructor(private readonly kafkaService: ClientKafkaProxy) {}

  async onModuleInit() {
    await this.kafkaService.connect()
    this.logger.debug('Kafka audit log publisher initialized')
  }

  async write(auditLogs: AuditLog[]): Promise<void> {
    const messages: Message[] = auditLogs.map((auditLog) => ({
      key: auditLog.organizationId,
      value: JSON.stringify(auditLog),
    }))

    try {
      await this.kafkaService.producer.send({
        topic: AUDIT_KAFKA_TOPIC,
        messages: messages,
        acks: -1,
        compression: CompressionTypes.GZIP,
      })
    } catch (error) {
      this.logger.error('Failed to write audit log to Kafka:', error)
      throw error
    }
    this.logger.debug(`${auditLogs.length} audit logs written to Kafka`)
  }
}
