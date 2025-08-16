/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { WebhookService } from './webhook.service'
import { WebhookEvents } from '../constants/webhook-events.constant'
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

  @OnEvent(SandboxEvents.DESIRED_STATE_UPDATED)
  async handleSandboxDesiredStateUpdated(event: SandboxDesiredStateUpdatedEvent) {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(event.sandbox.organizationId, WebhookEvents.SANDBOX_DESIRED_STATE_UPDATED, {
        id: event.sandbox.id,
        organizationId: event.sandbox.organizationId,
        oldDesiredState: event.oldDesiredState,
        newDesiredState: event.newDesiredState,
        updatedAt: event.sandbox.updatedAt,
      })
    } catch (error) {
      this.logger.error(`Failed to send webhook for sandbox desired state updated: ${error.message}`)
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

  @OnEvent(VolumeEvents.LAST_USED_AT_UPDATED)
  async handleVolumeLastUsedAtUpdated(event: VolumeLastUsedAtUpdatedEvent) {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(event.volume.organizationId, WebhookEvents.VOLUME_LAST_USED_AT_UPDATED, {
        id: event.volume.id,
        name: event.volume.name,
        organizationId: event.volume.organizationId,
        lastUsedAt: event.volume.lastUsedAt,
        updatedAt: event.volume.updatedAt,
      })
    } catch (error) {
      this.logger.error(`Failed to send webhook for volume last used at updated: ${error.message}`)
    }
  }

  @OnEvent(UserEvents.CREATED)
  async handleUserCreated() {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      // Note: This would need to be sent to all organizations the user belongs to
      // For now, we'll skip this as it's more complex to implement
      this.logger.debug('Skipping user created webhook - would need to send to all user organizations')
    } catch (error) {
      this.logger.error(`Failed to send webhook for user created: ${error.message}`)
    }
  }

  @OnEvent(AuditLogEvents.CREATED)
  async handleAuditLogCreated(event: AuditLogCreatedEvent) {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(event.auditLog.organizationId, WebhookEvents.AUDIT_LOG_CREATED, {
        id: event.auditLog.id,
        organizationId: event.auditLog.organizationId,
        action: event.auditLog.action,
        targetType: event.auditLog.targetType,
        targetId: event.auditLog.targetId,
        createdAt: event.auditLog.createdAt,
      })
    } catch (error) {
      this.logger.error(`Failed to send webhook for audit log created: ${error.message}`)
    }
  }

  @OnEvent(AuditLogEvents.UPDATED)
  async handleAuditLogUpdated(event: AuditLogUpdatedEvent) {
    if (!this.webhookService.isEnabled()) {
      return
    }

    try {
      await this.webhookService.sendWebhook(event.auditLog.organizationId, WebhookEvents.AUDIT_LOG_UPDATED, {
        id: event.auditLog.id,
        organizationId: event.auditLog.organizationId,
        action: event.auditLog.action,
        targetType: event.auditLog.targetType,
        targetId: event.auditLog.targetId,
        updatedAt: event.auditLog.createdAt,
      })
    } catch (error) {
      this.logger.error(`Failed to send webhook for audit log updated: ${error.message}`)
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
