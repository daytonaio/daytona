/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logger, OnModuleInit } from '@nestjs/common'
import { errors } from '@opensearch-project/opensearch'
import { AuditLog } from '../entities/audit-log.entity'
import { PaginatedList } from '../../common/interfaces/paginated-list.interface'
import { AuditLogStorageAdapter } from '../interfaces/audit-storage.interface'
import { AuditLogFilter } from '../interfaces/audit-filter.interface'
import { TypedConfigService } from '../../config/typed-config.service'
import { OpensearchClient } from 'nestjs-opensearch'
import { PolicyEnvelope } from '@opensearch-project/opensearch/api/_types/ism._common.js'
import { QueryContainer } from '@opensearch-project/opensearch/api/_types/_common.query_dsl.js'
import { Bulk_RequestBody } from '@opensearch-project/opensearch/api/index.js'

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
    await this.initialize()
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

  async getAllLogs(page?: number, limit?: number, filters?: AuditLogFilter): Promise<PaginatedList<AuditLog>> {
    const from = (page - 1) * limit
    const size = limit

    // Build the main query for audit logs
    const query: QueryContainer = {
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

    // Get total number of audit logs
    const totalResponse = await this.client.count({
      index: this.indexName,
      body: {
        query,
      },
    })
    const total = totalResponse.body.count

    // Get the audit logs with proper pagination
    const response = await this.client.search({
      index: this.indexName,
      body: {
        query,
        sort: [{ createdAt: { order: 'desc' } }],
        from,
        size,
      },
    })

    return {
      items: response.body.hits?.hits?.map((hit) => this.mapSourceToAuditLog(hit._source)) || [],
      total,
      page,
      totalPages: Math.ceil(total / limit),
    }
  }

  async getOrganizationLogs(
    organizationId: string,
    page?: number,
    limit?: number,
    filters?: AuditLogFilter,
  ): Promise<PaginatedList<AuditLog>> {
    if (!organizationId) {
      throw new Error('Organization ID is required')
    }

    const from = (page - 1) * limit
    const size = limit

    // Build the main query for audit logs
    const query: QueryContainer = {
      bool: {
        filter: [
          {
            term: { organizationId: organizationId },
          },
          {
            range: {
              createdAt: {
                gt: filters?.from?.toISOString(),
                lt: filters?.to?.toISOString(),
              },
            },
          },
        ],
      },
    }

    // Get total number of audit logs
    const totalResponse = await this.client.count({
      index: this.indexName,
      body: { query },
    })
    const total = totalResponse.body.count

    // Get the audit logs with proper pagination
    const response = await this.client.search({
      index: this.indexName,
      body: {
        query,
        sort: [{ createdAt: { order: 'desc' } }],
        from,
        size,
      },
    })

    return {
      items: response.body.hits?.hits?.map((hit) => this.mapSourceToAuditLog(hit._source)) || [],
      total,
      page,
      totalPages: Math.ceil(total / limit),
    }
  }

  private async initialize() {
    this.logger.log('Initializing OpenSearch audit storage adapter')

    // Step 1: Create index template for the data stream (if it doesn't exist)
    const templateName = `${this.indexName}-template`
    try {
      await this.client.indices.getIndexTemplate({ name: templateName })
      this.logger.log(`Index template already exists: ${templateName}. Skipping creation.`)
    } catch (error) {
      if (error instanceof errors.ResponseError && error.statusCode === 404) {
        await this.createIndexTemplate(templateName)
        this.logger.log(`Created index template: ${templateName}.`)
        return
      }
      throw error
    }

    // Step 2: Create data stream (if it doesn't exist)
    try {
      await this.client.indices.getDataStream({ name: this.indexName })
      this.logger.log(`Data stream already exists: ${this.indexName}. Skipping creation.`)
    } catch (error) {
      if (error instanceof errors.ResponseError && error.statusCode === 404) {
        await this.client.indices.createDataStream({ name: this.indexName })
        this.logger.log(`Created data stream: ${this.indexName}.`)
        return
      }
      throw error
    }

    // Step 3: Set up cleanup (ISM for OpenSearch)
    await this.setupCleanup()

    this.logger.log('OpenSearch audit storage adapter initialized')
  }

  private async createIndexTemplate(templateName: string) {
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
              actorEmail: { type: 'keyword' },
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

  private async setupCleanup(): Promise<void> {
    try {
      const retentionDays = this.configService.get('audit.retentionDays') || 0
      if (!retentionDays || retentionDays < 1) {
        this.logger.log('Audit log retention not configured, skipping ILM setup')
        return
      }

      await this.createISMPolicy(retentionDays)
      await this.applyISMPolicyToIndexTemplate()

      this.logger.log(`OpenSearch ILM policy configured for ${retentionDays} days retention`)
    } catch (error) {
      this.logger.error(`Failed to setup ILM policy: ${error.message}`)
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
      const existingPolicy = await this.client.ism.existsPolicy({
        policy_id: policyName,
      })

      // If policy exists, update it
      if (existingPolicy.body == true) {
        this.logger.log(`ISM policy already exists: ${policyName}. Updating it.`)
        const existingPolicy = await this.client.ism.getPolicy({
          policy_id: policyName,
        })
        await this.client.ism.putPolicy({
          policy_id: policyName,
          if_primary_term: existingPolicy.body._primary_term,
          if_seq_no: existingPolicy.body._seq_no,
          body: policy,
        })
        return
        // If policy does not exist, create it
      } else {
        this.logger.log(`ISM policy does not exist: ${policyName}. Creating it.`)
        await this.client.ism.putPolicy({
          policy_id: policyName,
          body: policy,
        })
        return
      }
    } catch (error) {
      // Non-critical error, log warning and continue
      this.logger.warn(`Failed to create ISM policy: ${error.message}`)
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
        this.logger.warn(`Index template ${templateName} not found, cannot apply ILM policy`)
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

      this.logger.log(`Applied ILM policy ${policyName} to index template ${templateName}`)
    } catch (error) {
      this.logger.error(`Failed to apply ILM policy to index template: ${error.message}`)
    }
  }
}
