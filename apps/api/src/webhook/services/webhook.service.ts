/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnModuleInit, NotFoundException, ServiceUnavailableException } from '@nestjs/common'
import { TypedConfigService } from '../../config/typed-config.service'
import { Svix } from 'svix'
import { Organization } from '../../organization/entities/organization.entity'
import { InjectRepository } from '@nestjs/typeorm'
import { WebhookInitialization } from '../entities/webhook-initialization.entity'
import { Repository } from 'typeorm'

@Injectable()
export class WebhookService implements OnModuleInit {
  private readonly logger = new Logger(WebhookService.name)
  private svix: Svix | null = null

  constructor(
    private readonly configService: TypedConfigService,
    @InjectRepository(WebhookInitialization)
    private readonly webhookInitializationRepository: Repository<WebhookInitialization>,
  ) {}

  async onModuleInit() {
    const svixAuthToken = this.configService.get('webhook.authToken')
    if (svixAuthToken) {
      const serverUrl = this.configService.get('webhook.serverUrl')
      if (serverUrl) {
        this.svix = new Svix(svixAuthToken, { serverUrl })
      } else {
        this.svix = new Svix(svixAuthToken)
        //this.svix.eventType.importOpenapi
      }
      this.logger.log('Svix webhook service initialized')
    } else {
      this.logger.warn('SVIX_AUTH_TOKEN not configured, webhook service disabled')
    }
  }

  /**
   * Get webhook initialization status for an organization
   */
  async getInitializationStatus(organizationId: string): Promise<WebhookInitialization | null> {
    return this.webhookInitializationRepository.findOne({
      where: { organizationId },
    })
  }

  // TODO: Remove this once we decide to open webhooks to all organizations
  // @OnEvent(OrganizationEvents.CREATED)
  async handleOrganizationCreated(organization: Organization) {
    if (!this.svix) {
      this.logger.debug('Svix not configured, skipping webhook creation')
      return
    }

    try {
      // Create a new Svix application for this organization
      const svixAppId = await this.createSvixApplication(organization)
      this.logger.log(`Created Svix application for organization ${organization.id}: ${svixAppId}`)
    } catch (error) {
      this.logger.error(`Failed to create Svix application for organization ${organization.id}:`, error)
    }
  }

  /**
   * Create a Svix application for an organization
   */
  async createSvixApplication(organization: Organization): Promise<string> {
    if (!this.svix) {
      throw new ServiceUnavailableException('Webhook service is not configured')
    }

    let existingWebhookInitialization = await this.getInitializationStatus(organization.id)
    if (existingWebhookInitialization && existingWebhookInitialization.svixApplicationId) {
      this.logger.warn(
        `Svix application already exists for organization ${organization.id}: ${existingWebhookInitialization.svixApplicationId}`,
      )
      return existingWebhookInitialization.svixApplicationId
    } else {
      existingWebhookInitialization = new WebhookInitialization()
      existingWebhookInitialization.organizationId = organization.id
      existingWebhookInitialization.svixApplicationId = null
      existingWebhookInitialization.retryCount = -1
      existingWebhookInitialization.lastError = null
    }

    try {
      const svixApp = await this.svix.application.getOrCreate({
        name: organization.name,
        uid: organization.id,
      })
      existingWebhookInitialization.svixApplicationId = svixApp.id
      existingWebhookInitialization.retryCount = existingWebhookInitialization.retryCount + 1
      existingWebhookInitialization.lastError = null

      await this.webhookInitializationRepository.save(existingWebhookInitialization)

      this.logger.log(`Created Svix application for organization ${organization.id}: ${svixApp.id}`)
      return svixApp.id
    } catch (error) {
      existingWebhookInitialization.retryCount = existingWebhookInitialization.retryCount + 1
      existingWebhookInitialization.lastError = String(error)
      await this.webhookInitializationRepository.save(existingWebhookInitialization)
      this.logger.error(`Failed to create Svix application for organization ${organization.id}:`, error)
      throw error
    }
  }

  private static readonly ENDPOINT_FLAG_TTL_MS = 60_000

  /**
   * Refresh hasEndpoints by counting endpoints in Svix.
   * On error, returns the input row untouched. Callers distinguish a freshly-refreshed row
   * from a stale one via endpointsCheckedAt and fall back to attempting delivery rather than
   * treating a never-confirmed false as authoritative.
   */
  private async refreshEndpointFlag(init: WebhookInitialization): Promise<WebhookInitialization> {
    if (!this.svix) {
      return init
    }

    try {
      const result = await this.svix.endpoint.list(init.organizationId, { limit: 1 })
      return await this.webhookInitializationRepository.save({
        ...init,
        hasEndpoints: result.data.length > 0,
        endpointsCheckedAt: new Date(),
      })
    } catch (error) {
      this.logger.error(`Failed to refresh endpoint flag for organization ${init.organizationId}:`, error)
      return init
    }
  }

  /**
   * Public wrapper for the refresh used by the controller ping route.
   */
  async refreshEndpointFlagByOrg(organizationId: string): Promise<void> {
    const init = await this.getInitializationStatus(organizationId)
    if (!init) {
      throw new NotFoundException('Webhook initialization status not found')
    }
    await this.refreshEndpointFlag(init)
  }

  /**
   * Send a webhook message to all endpoints of an organization
   */
  async sendWebhook(organizationId: string, eventType: string, payload: any, eventId?: string): Promise<void> {
    if (!this.svix) {
      this.logger.debug('Svix not configured, skipping webhook delivery')
      return
    }

    try {
      // Check if webhooks are initialized for this organization
      let init = await this.getInitializationStatus(organizationId)

      if (!init) {
        this.logger.debug(`Skipping webhook ${eventType} for organization ${organizationId}: webhooks not initialized`)
        return
      }

      const isFresh = (checkedAt?: Date) =>
        checkedAt !== undefined &&
        checkedAt !== null &&
        Date.now() - checkedAt.getTime() <= WebhookService.ENDPOINT_FLAG_TTL_MS
      if (!isFresh(init.endpointsCheckedAt)) {
        init = await this.refreshEndpointFlag(init)
      }

      // Only treat hasEndpoints=false as authoritative when we have a recent confirmation;
      // if the refresh failed and the flag is stale, fall through to message.create rather
      // than silently dropping (a Svix outage should surface as a send failure, not a quiet skip).
      if (!init.hasEndpoints && isFresh(init.endpointsCheckedAt)) {
        this.logger.debug(`Skipping webhook ${eventType} for organization ${organizationId}: no endpoints`)
        return
      }

      // Send the webhook message
      await this.svix.message.create(organizationId, {
        eventType,
        payload,
        eventId,
      })

      this.logger.debug(`Sent webhook ${eventType} to organization ${organizationId}`)
    } catch (error) {
      this.logger.error(`Failed to send webhook ${eventType} to organization ${organizationId}:`, error)
      throw error
    }
  }

  /**
   * Get webhook delivery attempts for a message
   */
  async getMessageAttempts(organizationId: string, messageId: string): Promise<any[]> {
    if (!this.svix) {
      throw new ServiceUnavailableException('Webhook service is not configured')
    }

    try {
      const attempts = await this.svix.messageAttempt.listByMsg(organizationId, messageId)
      return attempts.data
    } catch (error) {
      this.logger.error(`Failed to get message attempts for message ${messageId}:`, error)
      throw error
    }
  }

  /**
   * Check if webhook service is enabled
   */
  isEnabled(): boolean {
    return this.svix !== null
  }

  /**
   * Get Svix Consumer App Portal access for an organization
   */
  async getAppPortalAccess(organizationId: string): Promise<{ token: string; url: string }> {
    if (!this.svix) {
      throw new ServiceUnavailableException('Webhook service is not configured')
    }
    try {
      const appPortalAccess = await this.svix.authentication.appPortalAccess(organizationId, {})
      this.logger.debug(`Generated app portal access for organization ${organizationId}`)
      return {
        token: appPortalAccess.token,
        url: appPortalAccess.url,
      }
    } catch (error) {
      this.logger.debug(`Failed to generate app portal access for organization ${organizationId}:`, error)
      if (error.code === 404) {
        throw new NotFoundException(`Organization ${organizationId} not found in Svix`)
      }
      throw error
    }
  }
}
