/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { WebhookInitialization } from '../entities/webhook-initialization.entity'

@Injectable()
export class WebhookInitializationCheckerService {
  private readonly logger = new Logger(WebhookInitializationCheckerService.name)

  constructor(
    @InjectRepository(WebhookInitialization)
    private readonly webhookInitializationRepository: Repository<WebhookInitialization>,
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
}
