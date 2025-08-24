/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { WebhookService } from './webhook.service'
import { SandboxEvents } from '../../sandbox/constants/sandbox-events.constants'
import { SnapshotEvents } from '../../sandbox/constants/snapshot-events'
import { VolumeEvents } from '../../sandbox/constants/volume-events'
import { AuditLogEvents } from '../../audit/constants/audit-log-events.constant'
import { UserEvents } from '../../user/constants/user-events.constant'
import { SandboxCreatedEvent } from '../../sandbox/events/sandbox-create.event'
import { SandboxStateUpdatedEvent } from '../../sandbox/events/sandbox-state-updated.event'
import { SandboxDesiredStateUpdatedEvent } from '../../sandbox/events/sandbox-desired-state-updated.event'
import { SnapshotCreatedEvent } from '../../sandbox/events/snapshot-created.event'
import { SnapshotStateUpdatedEvent } from '../../sandbox/events/snapshot-state-updated.event'
import { SnapshotRemovedEvent } from '../../sandbox/events/snapshot-removed.event'
import { VolumeCreatedEvent } from '../../sandbox/events/volume-created.event'
import { VolumeStateUpdatedEvent } from '../../sandbox/events/volume-state-updated.event'
import { VolumeLastUsedAtUpdatedEvent } from '../../sandbox/events/volume-last-used-at-updated.event'
import { AuditLogCreatedEvent } from '../../audit/events/audit-log-created.event'
import { AuditLogUpdatedEvent } from '../../audit/events/audit-log-updated.event'
import { WebhookEvents } from '../constants/webhook-events.constants'

@Injectable()
export class WebhookEventHandlerService {
  private readonly logger = new Logger(WebhookEventHandlerService.name)

  constructor(private readonly webhookService: WebhookService) {}

  @OnEvent(SandboxEvents.CREATED)
  async handleSandboxCreated(event: SandboxCreatedEvent) {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(event.sandbox.organizationId, WebhookEvents.SANDBOX_CREATED, {
        id: event.sandbox.id,
        organizationId: event.sandbox.organizationId,
        state: event.sandbox.state,
        class: event.sandbox.class,
        createdAt: event.sandbox.createdAt,
      })
    } catch (error) {
      this.logger.error(`Failed to send webhook for sandbox created: ${error.message}`)
    }
  }

  @OnEvent(SandboxEvents.STATE_UPDATED)
  async handleSandboxStateUpdated(event: SandboxStateUpdatedEvent) {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(event.sandbox.organizationId, WebhookEvents.SANDBOX_STATE_UPDATED, {
        id: event.sandbox.id,
        organizationId: event.sandbox.organizationId,
        oldState: event.oldState,
        newState: event.newState,
        updatedAt: event.sandbox.updatedAt,
      })
    } catch (error) {
      this.logger.error(`Failed to send webhook for sandbox state updated: ${error.message}`)
    }
  }

  @OnEvent(SnapshotEvents.CREATED)
  async handleSnapshotCreated(event: SnapshotCreatedEvent) {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(event.snapshot.organizationId, WebhookEvents.SNAPSHOT_CREATED, {
        id: event.snapshot.id,
        name: event.snapshot.name,
        organizationId: event.snapshot.organizationId,
        state: event.snapshot.state,
        createdAt: event.snapshot.createdAt,
      })
    } catch (error) {
      this.logger.error(`Failed to send webhook for snapshot created: ${error.message}`)
    }
  }

  @OnEvent(SnapshotEvents.STATE_UPDATED)
  async handleSnapshotStateUpdated(event: SnapshotStateUpdatedEvent) {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(event.snapshot.organizationId, WebhookEvents.SNAPSHOT_STATE_UPDATED, {
        id: event.snapshot.id,
        name: event.snapshot.name,
        organizationId: event.snapshot.organizationId,
        oldState: event.oldState,
        newState: event.newState,
        updatedAt: event.snapshot.updatedAt,
      })
    } catch (error) {
      this.logger.error(`Failed to send webhook for snapshot state updated: ${error.message}`)
    }
  }

  @OnEvent(SnapshotEvents.REMOVED)
  async handleSnapshotRemoved(event: SnapshotRemovedEvent) {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(event.snapshot.organizationId, WebhookEvents.SNAPSHOT_REMOVED, {
        id: event.snapshot.id,
        name: event.snapshot.name,
        organizationId: event.snapshot.organizationId,
        removedAt: new Date().toISOString(),
      })
    } catch (error) {
      this.logger.error(`Failed to send webhook for snapshot removed: ${error.message}`)
    }
  }

  @OnEvent(VolumeEvents.CREATED)
  async handleVolumeCreated(event: VolumeCreatedEvent) {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(event.volume.organizationId, WebhookEvents.VOLUME_CREATED, {
        id: event.volume.id,
        name: event.volume.name,
        organizationId: event.volume.organizationId,
        state: event.volume.state,
        createdAt: event.volume.createdAt,
      })
    } catch (error) {
      this.logger.error(`Failed to send webhook for volume created: ${error.message}`)
    }
  }

  @OnEvent(VolumeEvents.STATE_UPDATED)
  async handleVolumeStateUpdated(event: VolumeStateUpdatedEvent) {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(event.volume.organizationId, WebhookEvents.VOLUME_STATE_UPDATED, {
        id: event.volume.id,
        name: event.volume.name,
        organizationId: event.volume.organizationId,
        oldState: event.oldState,
        newState: event.newState,
        updatedAt: event.volume.updatedAt,
      })
    } catch (error) {
      this.logger.error(`Failed to send webhook for volume state updated: ${error.message}`)
    }
  }

  /**
   * Send a custom webhook event
   */
  async sendCustomWebhook(organizationId: string, eventType: string, payload: any, eventId?: string): Promise<void> {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(organizationId, eventType, payload, eventId)
    } catch (error) {
      this.logger.error(`Failed to send custom webhook: ${error.message}`)
    }
  }
}
