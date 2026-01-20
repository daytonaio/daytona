/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Provider } from '@nestjs/common'
import { getRepositoryToken } from '@nestjs/typeorm'
import { OpensearchClient } from 'nestjs-opensearch'
import { Repository } from 'typeorm'
import { SandboxOpenSearchAdapter } from '../adapters/sandbox-opensearch.adapter'
import { SandboxTypeormSearchAdapter } from '../adapters/sandbox-typeorm.adapter'
import { SANDBOX_SEARCH_ADAPTER } from '../constants/sandbox-tokens'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxSearchAdapter } from '../interfaces/sandbox-search.interface'
import { TypedConfigService } from '../../config/typed-config.service'

export const SandboxSearchAdapterProvider: Provider = {
  provide: SANDBOX_SEARCH_ADAPTER,
  useFactory: (
    configService: TypedConfigService,
    opensearchClient: OpensearchClient,
    sandboxRepository: Repository<Sandbox>,
  ): SandboxSearchAdapter => {
    const sandboxSearchConfig = configService.get('sandboxSearch')

    if (sandboxSearchConfig.publish.enabled) {
      switch (sandboxSearchConfig.publish.storageAdapter) {
        case 'opensearch': {
          return new SandboxOpenSearchAdapter(configService, opensearchClient)
        }
        default:
          throw new Error(`Invalid storage adapter: ${sandboxSearchConfig.publish.storageAdapter}`)
      }
    } else {
      return new SandboxTypeormSearchAdapter(sandboxRepository)
    }
  },
  inject: [TypedConfigService, OpensearchClient, getRepositoryToken(Sandbox)],
}
