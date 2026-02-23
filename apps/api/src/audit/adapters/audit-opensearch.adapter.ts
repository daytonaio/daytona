/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, Logger, OnModuleInit } from '@nestjs/common'
import { errors } from '@opensearch-project/opensearch'
import { AuditLog } from '../entities/audit-log.entity'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'
import { AuditLogStorageAdapter } from '../interfaces/audit-storage.interface'
import { AuditLogFilter } from '../interfaces/audit-filter.interface'
import { TypedConfigService } from '../../config/typed-config.service'
import { OpensearchClient } from 'nestjs-opensearch'
import { PolicyEnvelope } from '@opensearch-project/opensearch/api/_types/ism._common.js'
import { QueryContainer } from '@opensearch-project/opensearch/api/_types/_common.query_dsl.js'
import { Bulk_RequestBody, Search_RequestBody, Search_Response } from '@opensearch-project/opensearch/api/index.js'
import { TotalHits } from '@opensearch-project/opensearch/api/_types/_core.search.js'
import { isMatch } from 'es-toolkit/compat'

// Safe limit for offset-based pagination to avoid hitting OpenSearch's 10000 limit
const MAX_OFFSET_PAGINATION_LIMIT = 10000

export class AuditOpenSearchStorageAdapter implements AuditLogStorageAdapter, OnModuleInit {
  private readonly logger = new Logger(AuditOpenSearchStorageAdapter.name)
  private indexName: string

  constructor(
    private readonly configService: TypedConfigService,
    private readonly client: OpensearchClient,
  ) {
    this.indexName = configService.getOrThrow('audit.publish.opensearchIndexName')
  }

  async onModuleInit(): Promise<void> {
    await this.putIndexTemplate()
    await this.setupISM()
    await this.createDataStream()

    this.logger.log('OpenSearch audit storage adapter initialized')
  }

  async write(auditLogs: AuditLog[]): Promise<void> {
    try {
      const documents = auditLogs.map((auditLog) => ({
        '@timestamp': new Date(), // Required field for data streams
        ...auditLog,
      }))

      // Include document ID to prevent duplicates
      const bulkBody: Bulk_RequestBody = documents.flatMap((document) => [
        { create: { _index: this.indexName, _id: document.id } },
        document,
      ])

      const response = await this.client.bulk({
        body: bulkBody,
        refresh: false,
      })

      if (response.body.errors) {
        const errors = response.body.items
          .filter((item: any) => item.create?.error)
          .map((item: any) => item.create.error)

        // Check if any errors are not 409 (idempotent errors are OK) or version conflicts (also idempotent)
        const nonIdempotentErrors = errors.filter(
          (error: any) => error.status !== 409 && error.type !== 'version_conflict_engine_exception',
        )

        if (nonIdempotentErrors.length > 0) {
          throw new Error(`OpenSearch bulk operation failed: ${JSON.stringify(nonIdempotentErrors)}`)
        }
      }

      this.logger.debug(`Saved ${auditLogs.length} audit logs to OpenSearch`)
    } catch (error) {
      this.logger.error(`Failed to save audit log to OpenSearch: ${error.message}`)
      throw error
    }
  }

  async getAllLogs(
    page?: number,
    limit?: number,
    filters?: AuditLogFilter,
    nextToken?: string,
  ): Promise<PaginatedList<AuditLog>> {
    const query = this.buildDateRangeQuery(filters)
    const searchBody = this.buildSearchBody(query, page, limit, nextToken)
    const response = await this.executeSearch(searchBody)
    return this.processSearchResponse(response, page, limit, nextToken, query)
  }

  async getOrganizationLogs(
    organizationId: string,
    page?: number,
    limit?: number,
    filters?: AuditLogFilter,
    nextToken?: string,
  ): Promise<PaginatedList<AuditLog>> {
    if (!organizationId) {
      throw new Error('Organization ID is required')
    }

    const query = this.buildOrganizationQuery(organizationId, filters)
    const searchBody = this.buildSearchBody(query, page, limit, nextToken)
    const response = await this.executeSearch(searchBody)
    return this.processSearchResponse(response, page, limit, nextToken, query)
  }

  private async createDataStream() {
    try {
      await this.client.indices.createDataStream({ name: this.indexName })
      this.logger.debug(`Created data stream: ${this.indexName}.`)
    } catch (error) {
      if (error instanceof errors.ResponseError && error.body.error.type === 'resource_already_exists_exception') {
        this.logger.debug(`Data stream already exists: ${this.indexName}. Skipping creation.`)
        return
      }
      throw error
    }
  }

  private async putIndexTemplate() {
    const templateName = `${this.indexName}-template`
    await this.client.indices.putIndexTemplate({
      name: templateName,
      body: {
        index_patterns: [`${this.indexName}*`],
        data_stream: {},
        template: {
          settings: {
            index: {
              number_of_shards: 1,
              number_of_replicas: 1,
            },
          },
          mappings: {
            dynamic: 'true',
            dynamic_templates: [
              {
                ids_as_keyword: {
                  match: '*Id',
                  mapping: { type: 'keyword', index: true },
                },
              },
              {
                default_strings: {
                  match: '*',
                  match_mapping_type: 'string',
                  mapping: { type: 'keyword', index: false },
                },
              },
              {
                non_queryable_fields: {
                  match: '*',
                  match_mapping_type: 'object',
                  mapping: {
                    type: 'object',
                    enabled: false,
                  },
                },
              },
            ],
            properties: {
              id: { type: 'keyword' },
              actorEmail: { type: 'keyword' },
              actorApiKeyPrefix: { type: 'keyword' },
              actorApiKeySuffix: { type: 'keyword' },
              action: { type: 'keyword' },
              targetType: { type: 'keyword' },
              statusCode: { type: 'integer' },
              createdAt: { type: 'date' },
            },
          },
        },
      },
    })
  }

  private mapSourceToAuditLog(source: any): AuditLog {
    const auditLog = new AuditLog()
    auditLog.id = source.id
    auditLog.actorId = source.actorId
    auditLog.actorEmail = source.actorEmail
    auditLog.actorApiKeyPrefix = source.actorApiKeyPrefix
    auditLog.actorApiKeySuffix = source.actorApiKeySuffix
    auditLog.organizationId = source.organizationId
    auditLog.action = source.action
    auditLog.targetType = source.targetType
    auditLog.targetId = source.targetId
    auditLog.statusCode = source.statusCode
    auditLog.errorMessage = source.errorMessage
    auditLog.ipAddress = source.ipAddress
    auditLog.userAgent = source.userAgent
    auditLog.source = source.source
    auditLog.metadata = source.metadata
    auditLog.createdAt = new Date(source.createdAt)
    return auditLog
  }

  private async setupISM(): Promise<void> {
    try {
      const retentionDays = this.configService.get('audit.retentionDays') || 0
      if (!retentionDays || retentionDays < 1) {
        this.logger.debug('Audit log retention not configured, skipping ISM setup')
        return
      }

      await this.createISMPolicy(retentionDays)
      await this.applyISMPolicyToIndexTemplate()

      this.logger.debug(`OpenSearch ISM policy configured for ${retentionDays} days retention`)
    } catch (error) {
      this.logger.warn(`Failed to setup ISM policy: ${error.message}`)
    }
  }

  private async createISMPolicy(retentionDays: number): Promise<void> {
    const policyName = `${this.indexName}-lifecycle-policy`

    const policy: PolicyEnvelope = {
      policy: {
        description: `Lifecycle policy for audit logs with ${retentionDays} days retention`,
        default_state: 'hot',
        states: [
          {
            name: 'hot',
            actions: [
              {
                rollover: {
                  // incorrect client type definitions
                  // ref: https://github.com/opensearch-project/opensearch-js/issues/1001
                  min_index_age: '30d' as any,
                  min_primary_shard_size: '20gb' as any,
                  min_doc_count: 20_000_000,
                },
              },
            ],
            transitions: [
              {
                state_name: 'delete',
                conditions: {
                  min_index_age: `${retentionDays}d`, // Delete after retention period
                },
              },
            ],
          },
          {
            name: 'delete',
            actions: [
              {
                delete: {},
              },
            ],
          },
        ],
        ism_template: [
          {
            index_patterns: [`${this.indexName}*`],
            priority: 100,
          },
        ],
      },
    }

    try {
      // Check does policy already exist
      const existingPolicy = await this.client.ism.getPolicy({
        policy_id: policyName,
      })

      // Check does policy need to be updated
      if (isMatch(existingPolicy.body, policy)) {
        this.logger.debug(`ISM policy ${policyName} is up to date`)
      } else {
        this.logger.debug(`ISM policy ${policyName} is out of date. Updating it.`)
        await this.client.ism.putPolicy({
          policy_id: policyName,
          if_primary_term: existingPolicy.body._primary_term,
          if_seq_no: existingPolicy.body._seq_no,
          body: policy,
        })
        this.logger.debug(`ISM policy ${policyName} updated`)
      }
    } catch (error) {
      if (error instanceof errors.ResponseError && error.statusCode === 404) {
        this.logger.debug(`ISM policy ${policyName} not found, creating it.`)
        await this.client.ism.putPolicy({
          policy_id: policyName,
          body: policy,
        })
        this.logger.debug(`ISM policy ${policyName} created`)
        return
      }
      this.logger.error(`Failed to create ISM policy`, error)
      throw error
    }
  }

  private async applyISMPolicyToIndexTemplate(): Promise<void> {
    const templateName = `${this.indexName}-template`
    const policyName = `${this.indexName}-lifecycle-policy`

    try {
      // Get existing template
      const existingTemplate = await this.client.indices.getIndexTemplate({
        name: templateName,
      })

      if (!existingTemplate.body?.index_templates?.[0]) {
        this.logger.debug(`Index template ${templateName} not found, cannot apply ILM policy`)
        return
      }

      // Update template with ILM policy
      const template = existingTemplate.body.index_templates[0].index_template

      // Add ILM settings to the template
      if (!template.template) template.template = {}
      if (!template.template.settings) template.template.settings = {}
      if (!template.template.settings.index) template.template.settings.index = {}

      template.template.settings.index = {
        ...template.template.settings.index,
        'plugins.index_state_management.policy_id': policyName,
        'plugins.index_state_management.rollover_alias': this.indexName,
        number_of_shards: 1,
        number_of_replicas: 1,
        refresh_interval: '5s',
      }

      // Update the template
      await this.client.indices.putIndexTemplate({
        name: templateName,
        body: template,
      })

      this.logger.debug(`Applied ILM policy ${policyName} to index template ${templateName}`)
    } catch (error) {
      this.logger.error(`Failed to apply ILM policy to index template: ${error.message}`)
    }
  }

  private buildDateRangeQuery(filters?: AuditLogFilter): QueryContainer {
    return {
      bool: {
        filter: [
          {
            range: {
              createdAt: {
                gte: filters?.from?.toISOString(),
                lte: filters?.to?.toISOString(),
              },
            },
          },
        ],
      },
    }
  }

  private buildOrganizationQuery(organizationId: string, filters?: AuditLogFilter): QueryContainer {
    return {
      bool: {
        filter: [
          {
            term: { organizationId },
          },
          {
            range: {
              createdAt: {
                gte: filters?.from?.toISOString(),
                lte: filters?.to?.toISOString(),
              },
            },
          },
        ],
      },
    }
  }

  private buildSearchBody(
    query: QueryContainer,
    page?: number,
    limit?: number,
    nextToken?: string,
  ): Search_RequestBody {
    const size = limit
    const searchBody: Search_RequestBody = {
      query,
      sort: [{ createdAt: { order: 'desc' } }, { id: { order: 'desc' } }],
      size: size + 1, // Request one extra to check if there are more results
    }

    if (nextToken) {
      // Cursor-based pagination using search_after
      try {
        const searchAfter = JSON.parse(Buffer.from(nextToken, 'base64').toString())
        searchBody.search_after = searchAfter
        this.logger.debug(`Using cursor-based pagination with search_after: ${JSON.stringify(searchAfter)}`)
      } catch {
        throw new BadRequestException(`Invalid nextToken provided: ${nextToken}`)
      }
    } else {
      // Offset-based pagination - only use when within safe limits
      const from = (page - 1) * limit
      if (from + size <= MAX_OFFSET_PAGINATION_LIMIT) {
        searchBody.from = from
        this.logger.debug(`Using offset-based pagination: from=${from}, size=${size + 1}`)
      } else {
        throw new BadRequestException(
          `Offset-based pagination not supported for page ${page} with limit ${limit}. Please use cursor-based pagination with nextToken parameter instead.`,
        )
      }
    }

    return searchBody
  }

  private async executeSearch(searchBody: Search_RequestBody) {
    return await this.client.search({
      index: this.indexName,
      body: searchBody,
      track_total_hits: MAX_OFFSET_PAGINATION_LIMIT,
    })
  }

  private async processSearchResponse(
    response: Search_Response,
    page?: number,
    limit?: number,
    nextToken?: string,
    query?: QueryContainer,
  ): Promise<PaginatedList<AuditLog>> {
    const size = limit
    const hits = response.body.hits?.hits || []
    const totalHits = response.body.hits?.total as TotalHits
    const hasMore = hits.length > size
    const items = hasMore ? hits.slice(0, size) : hits

    // Generate nextToken when there are more results and we're approaching limits
    let nextTokenResult: string | undefined
    const currentOffset = nextToken ? 0 : (page - 1) * limit // If using cursor, we don't know the exact offset
    const nextPageOffset = currentOffset + limit
    const wouldExceedLimit = nextPageOffset >= MAX_OFFSET_PAGINATION_LIMIT

    // Only generate nextToken if we're already using cursor pagination OR if the next page would exceed the limit
    if (hasMore && items.length > 0 && (nextToken || wouldExceedLimit)) {
      const lastItem = items[items.length - 1]
      const searchAfter = [lastItem._source.createdAt, lastItem._source.id]
      nextTokenResult = Buffer.from(JSON.stringify(searchAfter)).toString('base64')
    }

    let total = totalHits?.value
    let totalPages = Math.ceil(total / limit)
    if (totalHits?.relation === 'gte') {
      // TODO: This should be cached to avoid hitting OpenSearch for every request
      const totalResponse = await this.client.count({
        index: this.indexName,
        body: { query },
      })
      total = totalResponse.body.count
      totalPages = Math.ceil(total / limit)
    }

    return {
      items: items.map((hit) => this.mapSourceToAuditLog(hit._source)),
      total,
      page: page || 1,
      totalPages,
      nextToken: nextTokenResult,
    }
  }
}
