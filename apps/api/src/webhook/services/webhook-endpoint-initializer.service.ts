/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { WebhookService } from './webhook.service'
import { OrganizationEvents } from '../../organization/constants/organization-events.constant'
import { Organization } from '../../organization/entities/organization.entity'
import { WebhookEvents } from '../constants/webhook-events.constant'

@Injectable()
export class WebhookEndpointInitializerService {
  private readonly logger = new Logger(WebhookEndpointInitializerService.name)

  constructor(private readonly webhookService: WebhookService) {}

  @OnEvent(OrganizationEvents.CREATED)
  async handleOrganizationCreated(organization: Organization) {
    if (!this.webhookService.isEnabled()) {
      this.logger.debug('Webhook service not enabled, skipping endpoint initialization')
      return
    }

    try {
      await this.createPreconfiguredEndpoints(organization.id)
      this.logger.log(`Created preconfigured webhook endpoints for organization ${organization.id}`)
    } catch (error) {
      this.logger.error(`Failed to create preconfigured webhook endpoints for organization ${organization.id}:`, error)
    }
  }

  /**
   * Create preconfigured webhook endpoints for all supported events
   */
  private async createPreconfiguredEndpoints(organizationId: string): Promise<string[]> {
    const endpoints = [
      // Sandbox events
      {
        url: `https://webhook.daytona.io/${organizationId}/sandbox-events`,
        description: 'Preconfigured endpoint for all sandbox events',
        eventTypes: [
          WebhookEvents.SANDBOX_CREATED,
          WebhookEvents.SANDBOX_STATE_UPDATED,
          WebhookEvents.SANDBOX_DESIRED_STATE_UPDATED,
        ],
      },
      // Snapshot events
      {
        url: `https://webhook.daytona.io/${organizationId}/snapshot-events`,
        description: 'Preconfigured endpoint for all snapshot events',
        eventTypes: [
          WebhookEvents.SNAPSHOT_CREATED,
          WebhookEvents.SNAPSHOT_STATE_UPDATED,
          WebhookEvents.SNAPSHOT_REMOVED,
        ],
      },
      // Volume events
      {
        url: `https://webhook.daytona.io/${organizationId}/volume-events`,
        description: 'Preconfigured endpoint for all volume events',
        eventTypes: [
          WebhookEvents.VOLUME_CREATED,
          WebhookEvents.VOLUME_STATE_UPDATED,
          WebhookEvents.VOLUME_LAST_USED_AT_UPDATED,
        ],
      },
      // Audit events
      {
        url: `https://webhook.daytona.io/${organizationId}/audit-events`,
        description: 'Preconfigured endpoint for all audit log events',
        eventTypes: [WebhookEvents.AUDIT_LOG_CREATED, WebhookEvents.AUDIT_LOG_UPDATED],
      },
      // All events endpoint
      {
        url: `https://webhook.daytona.io/${organizationId}/all-events`,
        description: 'Preconfigured endpoint for all events (comprehensive)',
        eventTypes: [
          WebhookEvents.SANDBOX_CREATED,
          WebhookEvents.SANDBOX_STATE_UPDATED,
          WebhookEvents.SANDBOX_DESIRED_STATE_UPDATED,
          WebhookEvents.SNAPSHOT_CREATED,
          WebhookEvents.SNAPSHOT_STATE_UPDATED,
          WebhookEvents.SNAPSHOT_REMOVED,
          WebhookEvents.VOLUME_CREATED,
          WebhookEvents.VOLUME_STATE_UPDATED,
          WebhookEvents.VOLUME_LAST_USED_AT_UPDATED,
          WebhookEvents.AUDIT_LOG_CREATED,
          WebhookEvents.AUDIT_LOG_UPDATED,
        ],
      },
    ]

    const createdEndpointIds: string[] = []

    // Create each preconfigured endpoint
    for (const endpoint of endpoints) {
      try {
        const createdEndpoint = await this.webhookService.createEndpoint(
          organizationId,
          endpoint.url,
          endpoint.description,
          endpoint.eventTypes,
        )
        createdEndpointIds.push(createdEndpoint.id)
        this.logger.debug(`Created preconfigured endpoint: ${endpoint.description}`)
      } catch (error) {
        this.logger.error(`Failed to create preconfigured endpoint ${endpoint.description}:`, error)
        // Continue with other endpoints even if one fails
      }
    }

    return createdEndpointIds
  }

  /**
   * Manually create preconfigured endpoints for an existing organization
   */
  async createPreconfiguredEndpointsForOrganization(organizationId: string): Promise<string[]> {
    if (!this.webhookService.isEnabled()) {
      throw new Error('Webhook service not enabled')
    }

    return await this.createPreconfiguredEndpoints(organizationId)
  }

  /**
   * Get the list of preconfigured endpoint configurations
   */
  getPreconfiguredEndpointConfigs(): Array<{
    url: string
    description: string
    eventTypes: string[]
  }> {
    return [
      {
        url: 'https://webhook.daytona.io/{organizationId}/sandbox-events',
        description: 'Preconfigured endpoint for all sandbox events',
        eventTypes: [
          WebhookEvents.SANDBOX_CREATED,
          WebhookEvents.SANDBOX_STATE_UPDATED,
          WebhookEvents.SANDBOX_DESIRED_STATE_UPDATED,
        ],
      },
      {
        url: 'https://webhook.daytona.io/{organizationId}/snapshot-events',
        description: 'Preconfigured endpoint for all snapshot events',
        eventTypes: [
          WebhookEvents.SNAPSHOT_CREATED,
          WebhookEvents.SNAPSHOT_STATE_UPDATED,
          WebhookEvents.SNAPSHOT_REMOVED,
        ],
      },
      {
        url: 'https://webhook.daytona.io/{organizationId}/volume-events',
        description: 'Preconfigured endpoint for all volume events',
        eventTypes: [
          WebhookEvents.VOLUME_CREATED,
          WebhookEvents.VOLUME_STATE_UPDATED,
          WebhookEvents.VOLUME_LAST_USED_AT_UPDATED,
        ],
      },
      {
        url: 'https://webhook.daytona.io/{organizationId}/audit-events',
        description: 'Preconfigured endpoint for all audit log events',
        eventTypes: [WebhookEvents.AUDIT_LOG_CREATED, WebhookEvents.AUDIT_LOG_UPDATED],
      },
      {
        url: 'https://webhook.daytona.io/{organizationId}/all-events',
        description: 'Preconfigured endpoint for all events (comprehensive)',
        eventTypes: [
          WebhookEvents.SANDBOX_CREATED,
          WebhookEvents.SANDBOX_STATE_UPDATED,
          WebhookEvents.SANDBOX_DESIRED_STATE_UPDATED,
          WebhookEvents.SNAPSHOT_CREATED,
          WebhookEvents.SNAPSHOT_STATE_UPDATED,
          WebhookEvents.SNAPSHOT_REMOVED,
          WebhookEvents.VOLUME_CREATED,
          WebhookEvents.VOLUME_STATE_UPDATED,
          WebhookEvents.VOLUME_LAST_USED_AT_UPDATED,
          WebhookEvents.AUDIT_LOG_CREATED,
          WebhookEvents.AUDIT_LOG_UPDATED,
        ],
      },
    ]
  }
}
