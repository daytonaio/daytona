/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnModuleInit } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { TypedConfigService } from '../../config/typed-config.service'
import { Svix } from 'svix'
import { OrganizationEvents } from '../../organization/constants/organization-events.constant'
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
      throw new Error('Svix not configured')
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
      const isInitialized = await this.getInitializationStatus(organizationId)

      if (!isInitialized) {
        this.logger.log(`Webhooks not initialized for organization ${organizationId}, creating Svix application now...`)
        // For now, we'll just log that initialization is needed
        // The actual initialization should be done through the API or event handler
        this.logger.warn(
          `Organization ${organizationId} needs webhook initialization. Please use the initialization API endpoint.`,
        )
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
      throw new Error('Svix not configured')
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
   * Get Svix Consumer App Portal access URL for an organization
   */
  async getAppPortalAccessUrl(organizationId: string): Promise<string> {
    if (!this.svix) {
      throw new Error('Svix not configured')
    }
    try {
      const dashboard = await this.svix.authentication.appPortalAccess(organizationId, {})
      this.logger.debug(`Generated app portal access URL for organization ${organizationId}`)
      return dashboard.url
    } catch (error) {
      this.logger.error(`Failed to generate app portal access URL for organization ${organizationId}:`, error)
      throw error
    }
  }
}
