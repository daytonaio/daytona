/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, Logger, OnModuleInit } from '@nestjs/common'
import { OpensearchClient } from 'nestjs-opensearch'
import { Search_RequestBody } from '@opensearch-project/opensearch/api/index.js'
import { QueryContainer } from '@opensearch-project/opensearch/api/_types/_common.query_dsl.js'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { TypedConfigService } from '../../config/typed-config.service'
import {
  SandboxSearchAdapter,
  SandboxSearchFilters,
  SandboxSearchPagination,
  SandboxSearchResult,
  SandboxSearchSort,
  SandboxSearchSortField,
} from '../interfaces/sandbox-search.interface'
import { SandboxListItemDto } from '../dto/sandbox-list-item.dto'

export class SandboxOpenSearchSearchAdapter implements SandboxSearchAdapter, OnModuleInit {
  private readonly logger = new Logger(SandboxOpenSearchSearchAdapter.name)
  private readonly indexName: string
  private readonly numberOfShards: number
  private readonly numberOfReplicas: number
  private jsonbFields: string[] = []

  private get templateName(): string {
    return `${this.indexName}-template`
  }

  private get pipelineName(): string {
    return `${this.indexName}-ingest-pipeline`
  }

  constructor(
    configService: TypedConfigService,
    private readonly client: OpensearchClient,
    private readonly resolveJsonbFields: () => string[],
  ) {
    this.indexName = configService.getOrThrow('sandboxSearch.publish.opensearchIndexName')
    this.numberOfShards = configService.getOrThrow('sandboxSearch.publish.numberOfShards')
    this.numberOfReplicas = configService.getOrThrow('sandboxSearch.publish.numberOfReplicas')
  }

  async onModuleInit(): Promise<void> {
    this.jsonbFields = this.resolveJsonbFields()
    await this.putIngestPipeline()
    await this.putIndexTemplate()
    await this.applyPipelineToExistingIndex()
    this.logger.log('OpenSearch sandbox search adapter initialized')
  }

  private async putIngestPipeline(): Promise<void> {
    await this.client.ingest.putPipeline({
      id: this.pipelineName,
      body: {
        description: 'Parses JSONB string fields into native JSON objects for indexing',
        processors: this.jsonbFields.map((field) => ({
          json: {
            field,
            if: `ctx.${field} instanceof String`,
            ignore_failure: true,
          },
        })),
      },
    })
    this.logger.debug(`Created ingest pipeline: ${this.pipelineName}`)
  }

  private async applyPipelineToExistingIndex(): Promise<void> {
    try {
      await this.client.indices.putSettings({
        index: this.indexName,
        body: {
          index: {
            default_pipeline: this.pipelineName,
          },
        },
      })
      this.logger.debug(`Applied ingest pipeline to existing index '${this.indexName}'`)
    } catch {
      this.logger.debug(`Index '${this.indexName}' does not exist yet, pipeline will apply via template`)
    }
  }

  private async putIndexTemplate(): Promise<void> {
    await this.client.indices.putIndexTemplate({
      name: this.templateName,
      body: {
        index_patterns: [`${this.indexName}*`],
        template: {
          settings: {
            index: {
              number_of_shards: this.numberOfShards,
              number_of_replicas: this.numberOfReplicas,
              default_pipeline: this.pipelineName,
            },
          },
          mappings: {
            dynamic: 'false',
            properties: {
              id: { type: 'keyword' },
              organizationId: { type: 'keyword' },
              name: { type: 'keyword' },
              region: { type: 'keyword' },
              runnerId: { type: 'keyword' },
              sandboxClass: { type: 'keyword' },
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
              labels: { type: 'flat_object' },
              backupState: { type: 'keyword' },
              daemonVersion: { type: 'keyword' },
            },
          },
        },
      },
    })
    this.logger.debug(`Created index template: ${this.templateName}`)
  }

  async search(params: {
    filters: SandboxSearchFilters
    pagination: SandboxSearchPagination
    sort: SandboxSearchSort
  }): Promise<SandboxSearchResult> {
    const query = this.buildSearchQuery(params.filters)
    const searchBody = this.buildSearchBody(query, params.pagination, params.sort)
    const response = await this.executeSearch(searchBody)
    return this.processSearchResponse(response, params.pagination.limit)
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
          must: [
            { terms: { state: [SandboxState.ERROR, SandboxState.BUILD_FAILED] } },
            { term: { desiredState: SandboxDesiredState.DESTROYED } },
          ],
        },
      })
    }

    if (filters.idPrefix) {
      must.push({
        prefix: {
          id: { value: filters.idPrefix, case_insensitive: true },
        },
      })
    }

    if (filters.namePrefix) {
      must.push({
        prefix: {
          name: { value: filters.namePrefix, case_insensitive: true },
        },
      })
    }

    // States filter
    if (filters.states?.length) {
      must.push({ terms: { state: filters.states } })
    } else {
      mustNot.push({ term: { state: SandboxState.DESTROYED } })
    }

    // Snapshots filter
    if (filters.snapshots?.length) {
      must.push({ terms: { snapshot: filters.snapshots } })
    }

    // Regions filter
    if (filters.regionIds?.length) {
      must.push({ terms: { region: filters.regionIds } })
    }

    // Sandbox class filter
    if (filters.sandboxClasses?.length) {
      must.push({ terms: { sandboxClass: filters.sandboxClasses } })
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

    // Labels filter (term queries on flat_object keys)
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
      sort: [{ [opensearchSortField]: { order: sort.direction } }, { id: { order: sort.direction } }],
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

  private processSearchResponse(response: any, limit: number): SandboxSearchResult {
    const hits = response.body.hits?.hits || []
    const hasMore = hits.length > limit
    const items = hasMore ? hits.slice(0, limit) : hits

    let nextCursor: string | null = null
    if (hasMore && items.length > 0) {
      const lastItem = items[items.length - 1]
      if (Array.isArray(lastItem.sort) && lastItem.sort.length > 0) {
        nextCursor = Buffer.from(JSON.stringify(lastItem.sort)).toString('base64')
      }
    }

    return {
      items: items.map((hit: any) => this.mapSourceToDto(hit._source)),
      nextCursor,
    }
  }

  private mapSourceToDto(source: any): SandboxListItemDto {
    const labels: { [key: string]: string } =
      typeof source.labels === 'string'
        ? JSON.parse(source.labels || '{}')
        : ((source.labels || {}) as { [key: string]: string })

    return new SandboxListItemDto({
      id: source.id,
      organizationId: source.organizationId,
      name: source.name,
      target: source.region,
      runnerId: source.runnerId,
      sandboxClass: source.sandboxClass,
      state: source.state as SandboxState,
      desiredState: source.desiredState as SandboxDesiredState | undefined,
      snapshot: source.snapshot,
      user: source.osUser,
      errorReason: source.errorReason,
      recoverable: source.recoverable,
      public: source.public,
      cpu: source.cpu,
      gpu: source.gpu,
      gpuType: source.gpu_type ?? undefined,
      memory: source.mem,
      disk: source.disk,
      labels,
      backupState: source.backupState,
      autoStopInterval: source.autoStopInterval,
      autoArchiveInterval: source.autoArchiveInterval,
      autoDeleteInterval: source.autoDeleteInterval,
      createdAt: source.createdAt ? new Date(source.createdAt).toISOString() : undefined,
      updatedAt: source.updatedAt ? new Date(source.updatedAt).toISOString() : undefined,
      lastActivityAt: source.lastActivityAt ? new Date(source.lastActivityAt).toISOString() : undefined,
      daemonVersion: source.daemonVersion,
    })
  }
}
