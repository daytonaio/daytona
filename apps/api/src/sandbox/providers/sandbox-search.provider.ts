/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Provider } from '@nestjs/common'
import { OpensearchClient } from 'nestjs-opensearch'
import { SandboxOpenSearchSearchAdapter } from '../adapters/sandbox-opensearch-search.adapter'
import { SandboxTypeormSearchAdapter } from '../adapters/sandbox-typeorm-search.adapter'
import { SANDBOX_SEARCH_ADAPTER } from '../constants/sandbox-tokens'
import { SandboxSearchAdapter } from '../interfaces/sandbox-search.interface'
import { TypedConfigService } from '../../config/typed-config.service'
import { SandboxRepository } from '../repositories/sandbox.repository'

export const SandboxSearchAdapterProvider: Provider = {
  provide: SANDBOX_SEARCH_ADAPTER,
  useFactory: (
    configService: TypedConfigService,
    opensearchClient: OpensearchClient,
    sandboxRepository: SandboxRepository,
  ): SandboxSearchAdapter => {
    const sandboxSearchConfig = configService.get('sandboxSearch')

    if (sandboxSearchConfig.publish.enabled) {
      switch (sandboxSearchConfig.publish.storageAdapter) {
        case 'opensearch':
          return new SandboxOpenSearchSearchAdapter(configService, opensearchClient, () =>
            sandboxRepository.metadata.columns.filter((col) => col.type === 'jsonb').map((col) => col.propertyName),
          )
        default:
          throw new Error(`Invalid storage adapter: ${sandboxSearchConfig.publish.storageAdapter}`)
      }
    } else {
      return new SandboxTypeormSearchAdapter(sandboxRepository)
    }
  },
  inject: [TypedConfigService, OpensearchClient, SandboxRepository],
}
