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
    const opensearchConfig = configService.get('opensearch.sandboxSearch')

    if (opensearchConfig?.enabled) {
      return new SandboxOpenSearchAdapter(configService, opensearchClient)
    }

    return new SandboxTypeormSearchAdapter(sandboxRepository)
  },
  inject: [TypedConfigService, OpensearchClient, getRepositoryToken(Sandbox)],
}
