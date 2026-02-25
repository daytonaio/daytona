/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module } from '@nestjs/common'
import { TypeOrmModule } from '@nestjs/typeorm'
import { OrganizationModule } from '../organization/organization.module'
import { AuditLog } from './entities/audit-log.entity'
import { AuditService } from './services/audit.service'
import { AuditInterceptor } from './interceptors/audit.interceptor'
import { RedisLockProvider } from '../sandbox/common/redis-lock.provider'
import { AuditController } from './controllers/audit.controller'
import { AuditKafkaConsumerController } from './publishers/kafka/audit-kafka-consumer.controller'
import { ClientsModule, Transport } from '@nestjs/microservices'
import { TypedConfigService } from '../config/typed-config.service'
import { Partitioners } from 'kafkajs'
import { OpensearchModule } from 'nestjs-opensearch'
import { AuditStorageAdapterProvider } from './providers/audit-storage.provider'
import { AuditPublisherProvider } from './providers/audit-publisher.provider'
import { AUDIT_KAFKA_SERVICE } from './constants/audit-tokens'

@Module({
  imports: [
    OrganizationModule,
    TypeOrmModule.forFeature([AuditLog]),
    ClientsModule.registerAsync([
      {
        name: AUDIT_KAFKA_SERVICE,
        inject: [TypedConfigService],
        useFactory: (configService: TypedConfigService) => {
          return {
            transport: Transport.KAFKA,
            options: {
              producerOnlyMode: true,
              client: configService.getKafkaClientConfig(),
              producer: {
                allowAutoTopicCreation: true,
                createPartitioner: Partitioners.DefaultPartitioner,
                idempotent: true,
              },
            },
          }
        },
      },
    ]),
    OpensearchModule.forRootAsync({
      inject: [TypedConfigService],
      useFactory: (configService: TypedConfigService) => {
        return configService.getOpenSearchConfig()
      },
    }),
  ],
  controllers: [AuditController, AuditKafkaConsumerController],
  providers: [AuditService, AuditInterceptor, RedisLockProvider, AuditStorageAdapterProvider, AuditPublisherProvider],
  exports: [AuditService, AuditInterceptor],
})
export class AuditModule {}
