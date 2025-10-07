/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Provider } from '@nestjs/common'
import { getRepositoryToken } from '@nestjs/typeorm'
import { AuditOpenSearchStorageAdapter } from '../adapters/audit-opensearch.adapter'
import { AuditLogStorageAdapter } from '../interfaces/audit-storage.interface'
import { AUDIT_STORAGE_ADAPTER } from '../constants/audit-tokens'
import { TypedConfigService } from '../../config/typed-config.service'
import { OpensearchClient } from 'nestjs-opensearch'
import { AuditTypeormStorageAdapter } from '../adapters/audit-typeorm.adapter'
import { Repository } from 'typeorm'
import { AuditLog } from '../entities/audit-log.entity'

export const AuditStorageAdapterProvider: Provider = {
  provide: AUDIT_STORAGE_ADAPTER,
  useFactory: (
    configService: TypedConfigService,
    opensearchClient: OpensearchClient,
    auditLogRepository: Repository<AuditLog>,
  ): AuditLogStorageAdapter => {
    const auditConfig = configService.get('audit')

    if (auditConfig.publish.enabled) {
      switch (auditConfig.publish.storageAdapter) {
        case 'opensearch': {
          return new AuditOpenSearchStorageAdapter(configService, opensearchClient)
        }
        default:
          throw new Error(`Invalid storage adapter: ${auditConfig.publish.storageAdapter}`)
      }
    } else {
      return new AuditTypeormStorageAdapter(auditLogRepository)
    }
  },
  inject: [TypedConfigService, OpensearchClient, getRepositoryToken(AuditLog)],
}
