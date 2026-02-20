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
import { SandboxCreatedEvent } from '../../sandbox/events/sandbox-create.event'
import { SandboxStateUpdatedEvent } from '../../sandbox/events/sandbox-state-updated.event'
import { SnapshotCreatedEvent } from '../../sandbox/events/snapshot-created.event'
import { SnapshotStateUpdatedEvent } from '../../sandbox/events/snapshot-state-updated.event'
import { SnapshotRemovedEvent } from '../../sandbox/events/snapshot-removed.event'
import { VolumeCreatedEvent } from '../../sandbox/events/volume-created.event'
import { VolumeStateUpdatedEvent } from '../../sandbox/events/volume-state-updated.event'
import { WebhookEvent } from '../constants/webhook-events.constants'
import {
  SandboxCreatedWebhookDto,
  SandboxStateUpdatedWebhookDto,
  SnapshotCreatedWebhookDto,
  SnapshotStateUpdatedWebhookDto,
  SnapshotRemovedWebhookDto,
  VolumeCreatedWebhookDto,
  VolumeStateUpdatedWebhookDto,
} from '../dto/webhook-event-payloads.dto'

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
      const payload = SandboxCreatedWebhookDto.fromEvent(event, WebhookEvent.SANDBOX_CREATED)
      await this.webhookService.sendWebhook(event.sandbox.organizationId, WebhookEvent.SANDBOX_CREATED, payload)
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
      const payload = SandboxStateUpdatedWebhookDto.fromEvent(event, WebhookEvent.SANDBOX_STATE_UPDATED)
      await this.webhookService.sendWebhook(event.sandbox.organizationId, WebhookEvent.SANDBOX_STATE_UPDATED, payload)
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
      const payload = SnapshotCreatedWebhookDto.fromEvent(event, WebhookEvent.SNAPSHOT_CREATED)
      await this.webhookService.sendWebhook(event.snapshot.organizationId, WebhookEvent.SNAPSHOT_CREATED, payload)
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
      const payload = SnapshotStateUpdatedWebhookDto.fromEvent(event, WebhookEvent.SNAPSHOT_STATE_UPDATED)
      await this.webhookService.sendWebhook(event.snapshot.organizationId, WebhookEvent.SNAPSHOT_STATE_UPDATED, payload)
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
      const payload = SnapshotRemovedWebhookDto.fromEvent(event, WebhookEvent.SNAPSHOT_REMOVED)
      await this.webhookService.sendWebhook(event.snapshot.organizationId, WebhookEvent.SNAPSHOT_REMOVED, payload)
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
      const payload = VolumeCreatedWebhookDto.fromEvent(event, WebhookEvent.VOLUME_CREATED)
      await this.webhookService.sendWebhook(event.volume.organizationId, WebhookEvent.VOLUME_CREATED, payload)
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
      const payload = VolumeStateUpdatedWebhookDto.fromEvent(event, WebhookEvent.VOLUME_STATE_UPDATED)
      await this.webhookService.sendWebhook(event.volume.organizationId, WebhookEvent.VOLUME_STATE_UPDATED, payload)
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
