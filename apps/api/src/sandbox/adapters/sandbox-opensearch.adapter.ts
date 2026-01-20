/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, Logger, OnModuleInit } from '@nestjs/common'
import { OpensearchClient } from 'nestjs-opensearch'
import { errors } from '@opensearch-project/opensearch'
import { Search_RequestBody } from '@opensearch-project/opensearch/api/index.js'
import { QueryContainer } from '@opensearch-project/opensearch/api/_types/_common.query_dsl.js'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { TypedConfigService } from '../../config/typed-config.service'
import { SandboxSearchSortField } from '../dto/search-sandboxes-query.dto'
import {
  SandboxSearchAdapter,
  SandboxSearchFilters,
  SandboxSearchPagination,
  SandboxSearchResult,
  SandboxSearchSort,
} from '../interfaces/sandbox-search.interface'
import { SandboxDto } from '../dto/sandbox.dto'

export class SandboxOpenSearchAdapter implements SandboxSearchAdapter, OnModuleInit {
  private readonly logger = new Logger(SandboxOpenSearchAdapter.name)
  private indexName: string

  constructor(
    configService: TypedConfigService,
    private readonly client: OpensearchClient,
  ) {
    this.indexName = configService.getOrThrow('opensearch.sandboxSearch.indexName')
  }

  async onModuleInit(): Promise<void> {
    await this.createIndex()
    this.logger.log('OpenSearch sandbox search adapter initialized')
  }

  private async createIndex(): Promise<void> {
    try {
      const exists = await this.client.indices.exists({ index: this.indexName })
      if (exists.body) {
        this.logger.debug(`Index already exists: ${this.indexName}. Skipping creation.`)
        return
      }

      await this.client.indices.create({
        index: this.indexName,
        body: {
          settings: {
            index: {
              number_of_shards: 1,
              number_of_replicas: 1,
            },
            analysis: {
              normalizer: {
                lowercase_normalizer: {
                  type: 'custom',
                  filter: ['lowercase'],
                },
              },
            },
          },
          mappings: {
            dynamic: 'strict',
            properties: {
              id: { type: 'keyword', normalizer: 'lowercase_normalizer' },
              organizationId: { type: 'keyword' },
              name: { type: 'keyword', normalizer: 'lowercase_normalizer' },
              regionId: { type: 'keyword' },
              runnerId: { type: 'keyword' },
              class: { type: 'keyword' },
              state: { type: 'keyword' },
              desiredState: { type: 'keyword' },
              snapshot: { type: 'keyword' },
              osUser: { type: 'keyword' },
              errorReason: { type: 'text' },
              recoverable: { type: 'boolean' },
              public: { type: 'boolean' },
              cpu: { type: 'integer' },
              gpu: { type: 'integer' },
              mem: { type: 'integer' },
              disk: { type: 'integer' },
              createdAt: { type: 'date' },
              lastActivityAt: { type: 'date' },
              autoStopInterval: { type: 'integer' },
              autoArchiveInterval: { type: 'integer' },
              autoDeleteInterval: { type: 'integer' },
              labels: { type: 'object', dynamic: 'true' },
              backupState: { type: 'keyword' },
              daemonVersion: { type: 'keyword' },
            },
          },
        },
      })
      this.logger.debug(`Created index: ${this.indexName}`)
    } catch (error) {
      if (error instanceof errors.ResponseError && error.body?.error?.type === 'resource_already_exists_exception') {
        this.logger.debug(`Index already exists: ${this.indexName}. Skipping creation.`)
        return
      }
      throw error
    }
  }

  async search(params: {
    filters: SandboxSearchFilters
    pagination: SandboxSearchPagination
    sort: SandboxSearchSort
  }): Promise<SandboxSearchResult> {
    const query = this.buildSearchQuery(params.filters)
    const searchBody = this.buildSearchBody(query, params.pagination, params.sort)
    const response = await this.executeSearch(searchBody)
    return this.processSearchResponse(response, params.pagination.limit, params.sort.field)
  }

  private buildSearchQuery(filters: SandboxSearchFilters): QueryContainer {
    const must: QueryContainer[] = []
    const mustNot: QueryContainer[] = []

    // Organization filter (required)
    must.push({ term: { organizationId: filters.organizationId } })

    // Exclude errored/deleted unless explicitly requested
    if (!filters.includeErroredDeleted) {
      mustNot.push({
        bool: {
          must: [{ term: { state: SandboxState.ERROR } }, { term: { desiredState: SandboxDesiredState.DESTROYED } }],
        },
      })
    }

    // ID prefix filter (case-insensitive)
    if (filters.idPrefix) {
      must.push({
        prefix: {
          id: filters.idPrefix.toLowerCase(),
        },
      })
    }

    // Name prefix filter (case-insensitive)
    if (filters.namePrefix) {
      must.push({
        prefix: {
          name: filters.namePrefix.toLowerCase(),
        },
      })
    }

    // States filter
    if (filters.states?.length) {
      must.push({ terms: { state: filters.states } })
    }

    // Snapshots filter
    if (filters.snapshots?.length) {
      must.push({ terms: { snapshot: filters.snapshots } })
    }

    // Regions filter
    if (filters.regionIds?.length) {
      must.push({ terms: { regionId: filters.regionIds } })
    }

    // CPU range filter
    if (filters.minCpu !== undefined || filters.maxCpu !== undefined) {
      must.push({
        range: {
          cpu: {
            ...(filters.minCpu !== undefined && { gte: filters.minCpu }),
            ...(filters.maxCpu !== undefined && { lte: filters.maxCpu }),
          },
        },
      })
    }

    // Memory range filter
    if (filters.minMemoryGiB !== undefined || filters.maxMemoryGiB !== undefined) {
      must.push({
        range: {
          mem: {
            ...(filters.minMemoryGiB !== undefined && { gte: filters.minMemoryGiB }),
            ...(filters.maxMemoryGiB !== undefined && { lte: filters.maxMemoryGiB }),
          },
        },
      })
    }

    // Disk range filter
    if (filters.minDiskGiB !== undefined || filters.maxDiskGiB !== undefined) {
      must.push({
        range: {
          disk: {
            ...(filters.minDiskGiB !== undefined && { gte: filters.minDiskGiB }),
            ...(filters.maxDiskGiB !== undefined && { lte: filters.maxDiskGiB }),
          },
        },
      })
    }

    // Public filter
    if (filters.isPublic !== undefined) {
      must.push({ term: { public: filters.isPublic } })
    }

    // Recoverable filter
    if (filters.isRecoverable !== undefined) {
      must.push({ term: { recoverable: filters.isRecoverable } })
    }

    // Creation range filter
    if (filters.createdAtAfter || filters.createdAtBefore) {
      must.push({
        range: {
          createdAt: {
            ...(filters.createdAtAfter && { gte: filters.createdAtAfter.toISOString() }),
            ...(filters.createdAtBefore && { lte: filters.createdAtBefore.toISOString() }),
          },
        },
      })
    }

    // Last activity range filter
    if (filters.lastEventAfter || filters.lastEventBefore) {
      must.push({
        range: {
          lastActivityAt: {
            ...(filters.lastEventAfter && { gte: filters.lastEventAfter.toISOString() }),
            ...(filters.lastEventBefore && { lte: filters.lastEventBefore.toISOString() }),
          },
        },
      })
    }

    // Labels filter (using object field with dynamic mapping)
    if (filters.labels) {
      for (const [key, value] of Object.entries(filters.labels)) {
        must.push({ term: { [`labels.${key}`]: value } })
      }
    }

    return {
      bool: {
        must,
        must_not: mustNot.length > 0 ? mustNot : undefined,
      },
    }
  }

  private getSortFieldMapping(sortField: SandboxSearchSortField): string {
    const fieldMapping: Record<SandboxSearchSortField, string> = {
      [SandboxSearchSortField.NAME]: 'name',
      [SandboxSearchSortField.CPU]: 'cpu',
      [SandboxSearchSortField.MEMORY]: 'mem',
      [SandboxSearchSortField.DISK]: 'disk',
      [SandboxSearchSortField.LAST_ACTIVITY_AT]: 'lastActivityAt',
      [SandboxSearchSortField.CREATED_AT]: 'createdAt',
    }
    return fieldMapping[sortField]
  }

  private buildSearchBody(
    query: QueryContainer,
    pagination: SandboxSearchPagination,
    sort: SandboxSearchSort,
  ): Search_RequestBody {
    const opensearchSortField = this.getSortFieldMapping(sort.field)
    const searchBody: Search_RequestBody = {
      query,
      sort: [{ [opensearchSortField]: { order: sort.direction } }, { _id: { order: 'desc' } }],
      size: pagination.limit + 1, // Request one extra to check if there are more results
    }

    if (pagination.cursor) {
      try {
        const searchAfter = JSON.parse(Buffer.from(pagination.cursor, 'base64').toString())
        searchBody.search_after = searchAfter
        this.logger.debug(`Using cursor-based pagination with search_after: ${JSON.stringify(searchAfter)}`)
      } catch {
        throw new BadRequestException(`Invalid cursor provided: ${pagination.cursor}`)
      }
    }

    return searchBody
  }

  private async executeSearch(searchBody: Search_RequestBody) {
    return await this.client.search({
      index: this.indexName,
      body: searchBody,
    })
  }

  private processSearchResponse(response: any, limit: number, sortField: SandboxSearchSortField): SandboxSearchResult {
    const hits = response.body.hits?.hits || []
    const hasMore = hits.length > limit
    const items = hasMore ? hits.slice(0, limit) : hits

    let nextCursor: string | null = null
    if (hasMore && items.length > 0) {
      const lastItem = items[items.length - 1]
      const opensearchSortField = this.getSortFieldMapping(sortField)
      const searchAfter = [lastItem._source[opensearchSortField], lastItem._source.id]
      nextCursor = Buffer.from(JSON.stringify(searchAfter)).toString('base64')
    }

    return {
      items: items.map((hit: any) => this.mapSourceToSandbox(hit._source)),
      nextCursor,
    }
  }

  private mapSourceToSandbox(source: any): SandboxDto {
    const sandbox = new Sandbox(source.regionId, source.name)
    sandbox.id = source.id
    sandbox.organizationId = source.organizationId
    sandbox.runnerId = source.runnerId
    sandbox.class = source.class
    sandbox.state = source.state
    sandbox.desiredState = source.desiredState
    sandbox.snapshot = source.snapshot
    sandbox.osUser = source.osUser
    sandbox.errorReason = source.errorReason
    sandbox.recoverable = source.recoverable
    sandbox.public = source.public
    sandbox.cpu = source.cpu
    sandbox.gpu = source.gpu
    sandbox.mem = source.mem
    sandbox.disk = source.disk
    sandbox.labels = source.labels
    sandbox.backupState = source.backupState
    sandbox.autoStopInterval = source.autoStopInterval
    sandbox.autoArchiveInterval = source.autoArchiveInterval
    sandbox.autoDeleteInterval = source.autoDeleteInterval
    sandbox.createdAt = source.createdAt ? new Date(source.createdAt) : undefined
    sandbox.updatedAt = source.lastActivityAt ? new Date(source.lastActivityAt) : undefined
    sandbox.lastActivityAt = source.lastActivityAt ? new Date(source.lastActivityAt) : undefined
    return SandboxDto.fromSandbox(sandbox)
  }
}
