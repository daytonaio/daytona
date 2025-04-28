/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceEvents } from '../../workspace/constants/workspace-events.constants'
import { WorkspaceCreatedEvent } from '../../workspace/events/workspace-create.event'
import { WorkspaceDesiredStateUpdatedEvent } from '../../workspace/events/workspace-desired-state-updated.event'
import { WorkspaceDestroyedEvent } from '../../workspace/events/workspace-destroyed.event'
import { WorkspacePublicStatusUpdatedEvent } from '../../workspace/events/workspace-public-status-updated.event'
import { WorkspaceResizedEvent } from '../../workspace/events/workspace-resized.event'
import { WorkspaceStartedEvent } from '../../workspace/events/workspace-started.event'
import { WorkspaceStateUpdatedEvent } from '../../workspace/events/workspace-state-updated.event'
import { WorkspaceStoppedEvent } from '../../workspace/events/workspace-stopped.event'
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

  @OnEvent(WorkspaceEvents.CREATED)
  async handleWorkspaceCreatedEvent(event: WorkspaceCreatedEvent) {
    this.logger.debug(`Workspace created: ${JSON.stringify(event)}`)
  }

  @OnEvent(WorkspaceEvents.STARTED)
  async handleWorkspaceStartedEvent(event: WorkspaceStartedEvent) {
    this.logger.debug(`Workspace started: ${JSON.stringify(event)}`)
  }

  @OnEvent(WorkspaceEvents.STOPPED)
  async handleWorkspaceStoppedEvent(event: WorkspaceStoppedEvent) {
    this.logger.debug(`Workspace stopped: ${JSON.stringify(event)}`)
  }

  @OnEvent(WorkspaceEvents.DESTROYED)
  async handleWorkspaceDestroyedEvent(event: WorkspaceDestroyedEvent) {
    this.logger.debug(`Workspace destroyed: ${JSON.stringify(event)}`)
  }

  @OnEvent(WorkspaceEvents.RESIZED)
  async handleWorkspaceResizedEvent(event: WorkspaceResizedEvent) {
    this.logger.debug(`Workspace resized: ${JSON.stringify(event)}`)
  }

  @OnEvent(WorkspaceEvents.PUBLIC_STATUS_UPDATED)
  async handleWorkspacePublicStatusUpdatedEvent(event: WorkspacePublicStatusUpdatedEvent) {
    this.logger.debug(`Workspace public status updated: ${JSON.stringify(event)}`)
  }

  @OnEvent(WorkspaceEvents.DESIRED_STATE_UPDATED)
  async handleWorkspaceDesiredStateUpdatedEvent(event: WorkspaceDesiredStateUpdatedEvent) {
    this.logger.debug(`Workspace desired state updated: ${JSON.stringify(event)}`)
  }

  @OnEvent(WorkspaceEvents.STATE_UPDATED)
  async handleWorkspaceStateUpdatedEvent(event: WorkspaceStateUpdatedEvent) {
    this.logger.debug(`Workspace state updated: ${JSON.stringify(event)}`)
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
