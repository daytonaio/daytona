/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { SandboxEvents } from '../../sandbox/constants/sandbox-events.constants'
import { SandboxCreatedEvent } from '../../sandbox/events/sandbox-create.event'
import { SandboxDesiredStateUpdatedEvent } from '../../sandbox/events/sandbox-desired-state-updated.event'
import { SandboxDestroyedEvent } from '../../sandbox/events/sandbox-destroyed.event'
import { SandboxPublicStatusUpdatedEvent } from '../../sandbox/events/sandbox-public-status-updated.event'
import { SandboxStartedEvent } from '../../sandbox/events/sandbox-started.event'
import { SandboxStateUpdatedEvent } from '../../sandbox/events/sandbox-state-updated.event'
import { SandboxStoppedEvent } from '../../sandbox/events/sandbox-stopped.event'
import { PostHog } from 'posthog-node'
import { OnAsyncEvent } from '../../common/decorators/on-async-event.decorator'
import { Organization } from '../../organization/entities/organization.entity'
import { OrganizationEvents } from '../../organization/constants/organization-events.constant'

@Injectable()
export class AnalyticsService {
  private readonly logger = new Logger(AnalyticsService.name)
  private readonly posthog?: PostHog

  constructor() {
    if (!process.env.POSTHOG_API_KEY) {
      return
    }

    if (!process.env.POSTHOG_HOST) {
      return
    }

    // Initialize PostHog client
    // Make sure to set POSTHOG_API_KEY in your environment variables
    this.posthog = new PostHog(process.env.POSTHOG_API_KEY, {
      host: process.env.POSTHOG_HOST,
    })
  }

  @OnEvent(SandboxEvents.CREATED)
  async handleSandboxCreatedEvent(event: SandboxCreatedEvent) {
    this.logger.debug(`Sandbox created: ${JSON.stringify(event)}`)
  }

  @OnEvent(SandboxEvents.STARTED)
  async handleSandboxStartedEvent(event: SandboxStartedEvent) {
    this.logger.debug(`Sandbox started: ${JSON.stringify(event)}`)
  }

  @OnEvent(SandboxEvents.STOPPED)
  async handleSandboxStoppedEvent(event: SandboxStoppedEvent) {
    this.logger.debug(`Sandbox stopped: ${JSON.stringify(event)}`)
  }

  @OnEvent(SandboxEvents.DESTROYED)
  async handleSandboxDestroyedEvent(event: SandboxDestroyedEvent) {
    this.logger.debug(`Sandbox destroyed: ${JSON.stringify(event)}`)
  }

  @OnEvent(SandboxEvents.PUBLIC_STATUS_UPDATED)
  async handleSandboxPublicStatusUpdatedEvent(event: SandboxPublicStatusUpdatedEvent) {
    this.logger.debug(`Sandbox public status updated: ${JSON.stringify(event)}`)
  }

  @OnEvent(SandboxEvents.DESIRED_STATE_UPDATED)
  async handleSandboxDesiredStateUpdatedEvent(event: SandboxDesiredStateUpdatedEvent) {
    this.logger.debug(`Sandbox desired state updated: ${JSON.stringify(event)}`)
  }

  @OnEvent(SandboxEvents.STATE_UPDATED)
  async handleSandboxStateUpdatedEvent(event: SandboxStateUpdatedEvent) {
    this.logger.debug(`Sandbox state updated: ${JSON.stringify(event)}`)
  }

  @OnAsyncEvent({
    event: OrganizationEvents.CREATED,
  })
  async handlePersonalOrganizationCreatedEvent(payload: Organization) {
    if (!payload.personal) {
      return
    }

    if (!this.posthog) {
      return
    }

    this.posthog.groupIdentify({
      groupType: 'organization',
      groupKey: payload.id,
      properties: {
        name: `Personal - ${payload.createdBy}`,
        created_at: payload.createdAt,
        created_by: payload.createdBy,
        personal: payload.personal,
      },
    })
  }
}
