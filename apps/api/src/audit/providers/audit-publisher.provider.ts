/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Provider } from '@nestjs/common'
import { AuditKafkaPublisher } from '../publishers/kafka/audit-kafka-publisher'
import { AuditDirectPublisher } from '../publishers/audit-direct-publisher'
import { AuditLogStorageAdapter } from '../interfaces/audit-storage.interface'
import { AuditLogPublisher } from '../interfaces/audit-publisher.interface'
import { AUDIT_KAFKA_SERVICE, AUDIT_STORAGE_ADAPTER, AUDIT_LOG_PUBLISHER } from '../constants/audit-tokens'
import { TypedConfigService } from '../../config/typed-config.service'
import { ClientKafka } from '@nestjs/microservices'

export const AuditPublisherProvider: Provider = {
  provide: AUDIT_LOG_PUBLISHER,
  useFactory: (
    configService: TypedConfigService,
    kafkaService: ClientKafka,
    auditStorageAdapter: AuditLogStorageAdapter,
  ): AuditLogPublisher => {
    const auditConfig = configService.get('audit')

    if (!auditConfig.publish.enabled) {
      return
    }

    switch (auditConfig.publish.mode) {
      case 'direct':
        return new AuditDirectPublisher(auditStorageAdapter)
      case 'kafka':
        if (!configService.get('kafka.enabled')) {
          throw new Error('Kafka must be enabled to publish audit logs to Kafka')
        }
        return new AuditKafkaPublisher(kafkaService)
      default:
        throw new Error(`Invalid publish mode: ${auditConfig.publish.mode}`)
    }
  },
  inject: [TypedConfigService, AUDIT_KAFKA_SERVICE, AUDIT_STORAGE_ADAPTER],
}
