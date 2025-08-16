/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { WebhookInitialization } from '../entities/webhook-initialization.entity'
import { WebhookService } from './webhook.service'
import { WebhookEndpointInitializerService } from './webhook-endpoint-initializer.service'
import { OrganizationService } from '../../organization/services/organization.service'

@Injectable()
export class WebhookInitializationManagerService {
  private readonly logger = new Logger(WebhookInitializationManagerService.name)

  constructor(
    @InjectRepository(WebhookInitialization)
    private readonly webhookInitializationRepository: Repository<WebhookInitialization>,
    private readonly webhookService: WebhookService,
    private readonly webhookEndpointInitializerService: WebhookEndpointInitializerService,
    private readonly organizationService: OrganizationService,
  ) {}

  /**
   * Check if webhooks are initialized for an organization
   */
  async isWebhookInitialized(organizationId: string): Promise<boolean> {
    const initialization = await this.webhookInitializationRepository.findOne({
      where: { organizationId },
    })

    return initialization?.endpointsCreated === true && initialization?.svixApplicationCreated === true
  }

  /**
   * Get webhook initialization status for an organization
   */
  async getInitializationStatus(organizationId: string): Promise<WebhookInitialization | null> {
    return this.webhookInitializationRepository.findOne({
      where: { organizationId },
    })
  }

  /**
   * Initialize webhooks for an organization (creates Svix app and endpoints)
   */
  async initializeWebhooks(organizationId: string): Promise<void> {
    if (!this.webhookService.isEnabled()) {
      this.logger.debug('Webhook service not enabled, skipping initialization')
      return
    }

    let initialization = await this.webhookInitializationRepository.findOne({
      where: { organizationId },
    })

    if (!initialization) {
      initialization = new WebhookInitialization()
      initialization.organizationId = organizationId
      initialization.endpointsCreated = false
      initialization.svixApplicationCreated = false
      initialization.endpointIds = []
      initialization.retryCount = 0
    }

    try {
      // Check if organization exists
      const organization = await this.organizationService.findOne(organizationId)
      if (!organization) {
        throw new Error(`Organization ${organizationId} not found`)
      }

      // Create Svix application if not exists
      if (!initialization.svixApplicationCreated) {
        try {
          const svixApp = await this.webhookService.createSvixApplication(organization)
          initialization.svixApplicationId = svixApp.id
          initialization.svixApplicationCreated = true
          this.logger.log(`Created Svix application for organization ${organizationId}: ${svixApp.id}`)
        } catch (error) {
          initialization.lastError = `Failed to create Svix application: ${error.message}`
          initialization.retryCount++
          await this.webhookInitializationRepository.save(initialization)
          throw error
        }
      }

      // Create webhook endpoints if not exists
      if (!initialization.endpointsCreated) {
        try {
          const endpointIds =
            await this.webhookEndpointInitializerService.createPreconfiguredEndpointsForOrganization(organizationId)
          initialization.endpointIds = endpointIds
          initialization.endpointsCreated = true
          initialization.lastError = null
          this.logger.log(`Created webhook endpoints for organization ${organizationId}`)
        } catch (error) {
          initialization.lastError = `Failed to create endpoints: ${error.message}`
          initialization.retryCount++
          await this.webhookInitializationRepository.save(initialization)
          throw error
        }
      }

      // Save successful initialization
      await this.webhookInitializationRepository.save(initialization)
      this.logger.log(`Successfully initialized webhooks for organization ${organizationId}`)
    } catch (error) {
      this.logger.error(`Failed to initialize webhooks for organization ${organizationId}:`, error)
      throw error
    }
  }

  /**
   * Initialize webhooks for all existing organizations
   */
  async initializeWebhooksForAllOrganizations(): Promise<void> {
    if (!this.webhookService.isEnabled()) {
      this.logger.debug('Webhook service not enabled, skipping bulk initialization')
      return
    }

    // Get all organizations - we'll need to implement this or use a different approach
    // For now, we'll skip bulk initialization until we have a way to get all organizations
    this.logger.warn('Bulk initialization not implemented - need to add findAll method to OrganizationService')
    return
  }

  /**
   * Retry failed initializations
   */
  async retryFailedInitializations(): Promise<void> {
    if (!this.webhookService.isEnabled()) {
      return
    }

    const failedInitializations = await this.webhookInitializationRepository.find({
      where: [{ endpointsCreated: false }, { svixApplicationCreated: false }],
    })

    this.logger.log(`Found ${failedInitializations.length} failed webhook initializations to retry`)

    for (const initialization of failedInitializations) {
      try {
        await this.initializeWebhooks(initialization.organizationId)
      } catch (error) {
        this.logger.error(`Retry failed for organization ${initialization.organizationId}:`, error)
      }
    }
  }

  /**
   * Update webhook endpoints for all organizations (for future updates)
   */
  async updateWebhookEndpointsForAllOrganizations(): Promise<void> {
    if (!this.webhookService.isEnabled()) {
      return
    }

    const initializedOrganizations = await this.webhookInitializationRepository.find({
      where: { endpointsCreated: true },
    })

    this.logger.log(`Updating webhook endpoints for ${initializedOrganizations.length} organizations`)

    for (const initialization of initializedOrganizations) {
      try {
        // Delete existing endpoints
        if (initialization.endpointIds) {
          for (const endpointId of initialization.endpointIds) {
            try {
              await this.webhookService.deleteEndpoint(initialization.organizationId, endpointId)
            } catch (error) {
              this.logger.warn(`Failed to delete endpoint ${endpointId}:`, error)
            }
          }
        }

        // Create new endpoints
        await this.webhookEndpointInitializerService.createPreconfiguredEndpointsForOrganization(
          initialization.organizationId,
        )

        // Update initialization record
        initialization.endpointsCreated = true
        initialization.lastError = null
        await this.webhookInitializationRepository.save(initialization)

        this.logger.log(`Updated webhook endpoints for organization ${initialization.organizationId}`)
      } catch (error) {
        this.logger.error(
          `Failed to update webhook endpoints for organization ${initialization.organizationId}:`,
          error,
        )
        initialization.lastError = `Update failed: ${error.message}`
        await this.webhookInitializationRepository.save(initialization)
      }
    }
  }

  /**
   * Get initialization statistics
   */
  async getInitializationStats(): Promise<{
    totalOrganizations: number
    initializedOrganizations: number
    failedInitializations: number
    pendingInitializations: number
  }> {
    // For now, we can't get total organizations count without implementing it in OrganizationService
    const initializedOrganizations = await this.webhookInitializationRepository.count({
      where: { endpointsCreated: true, svixApplicationCreated: true },
    })
    const failedInitializations = await this.webhookInitializationRepository.count({
      where: [{ endpointsCreated: false }, { svixApplicationCreated: false }],
    })

    return {
      totalOrganizations: 0, // TODO: Implement when OrganizationService.count() is available
      initializedOrganizations,
      failedInitializations,
      pendingInitializations: 0, // TODO: Calculate when totalOrganizations is available
    }
  }
}
