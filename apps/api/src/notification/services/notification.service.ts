/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { NotificationGateway } from '../gateways/notification.gateway'
import { SandboxEvents } from '../../sandbox/constants/sandbox-events.constants'
import { SandboxDto } from '../../sandbox/dto/sandbox.dto'
import { SandboxCreatedEvent } from '../../sandbox/events/sandbox-create.event'
import { SandboxStateUpdatedEvent } from '../../sandbox/events/sandbox-state-updated.event'
import { SnapshotCreatedEvent } from '../../sandbox/events/snapshot-created.event'
import { SnapshotEvents } from '../../sandbox/constants/snapshot-events'
import { SnapshotDto } from '../../sandbox/dto/snapshot.dto'
import { SnapshotStateUpdatedEvent } from '../../sandbox/events/snapshot-state-updated.event'
import { SnapshotRemovedEvent } from '../../sandbox/events/snapshot-removed.event'
import { VolumeEvents } from '../../sandbox/constants/volume-events'
import { VolumeCreatedEvent } from '../../sandbox/events/volume-created.event'
import { VolumeDto } from '../../sandbox/dto/volume.dto'
import { VolumeStateUpdatedEvent } from '../../sandbox/events/volume-state-updated.event'
import { VolumeLastUsedAtUpdatedEvent } from '../../sandbox/events/volume-last-used-at-updated.event'
import { SandboxDesiredStateUpdatedEvent } from '../../sandbox/events/sandbox-desired-state-updated.event'

@Injectable()
export class NotificationService {
  constructor(private readonly notificationGateway: NotificationGateway) {}

  @OnEvent(SandboxEvents.CREATED)
  async handleSandboxCreated(event: SandboxCreatedEvent) {
    this.notificationGateway.emitSandboxCreated(new SandboxDto(event.sandbox))
  }

  @OnEvent(SandboxEvents.STATE_UPDATED)
  async handleSandboxStateUpdated(event: SandboxStateUpdatedEvent) {
    this.notificationGateway.emitSandboxStateUpdated(new SandboxDto(event.sandbox), event.oldState, event.newState)
  }

  @OnEvent(SandboxEvents.DESIRED_STATE_UPDATED)
  async handleSandboxDesiredStateUpdated(event: SandboxDesiredStateUpdatedEvent) {
    this.notificationGateway.emitSandboxDesiredStateUpdated(
      new SandboxDto(event.sandbox),
      event.oldDesiredState,
      event.newDesiredState,
    )
  }

  @OnEvent(SnapshotEvents.CREATED)
  async handleSnapshotCreated(event: SnapshotCreatedEvent) {
    const dto = new SnapshotDto(event.snapshot)
    this.notificationGateway.emitSnapshotCreated(dto)
  }

  @OnEvent(SnapshotEvents.STATE_UPDATED)
  async handleSnapshotStateUpdated(event: SnapshotStateUpdatedEvent) {
    const dto = new SnapshotDto(event.snapshot)
    this.notificationGateway.emitSnapshotStateUpdated(dto, event.oldState, event.newState)
  }

  @OnEvent(SnapshotEvents.REMOVED)
  async handleSnapshotRemoved(event: SnapshotRemovedEvent) {
    const dto = new SnapshotDto(event.snapshot)
    this.notificationGateway.emitSnapshotRemoved(dto)
  }

  @OnEvent(VolumeEvents.CREATED)
  async handleVolumeCreated(event: VolumeCreatedEvent) {
    const dto = new VolumeDto(event.volume)
    this.notificationGateway.emitVolumeCreated(dto)
  }

  @OnEvent(VolumeEvents.STATE_UPDATED)
  async handleVolumeStateUpdated(event: VolumeStateUpdatedEvent) {
    const dto = new VolumeDto(event.volume)
    this.notificationGateway.emitVolumeStateUpdated(dto, event.oldState, event.newState)
  }

  @OnEvent(VolumeEvents.LAST_USED_AT_UPDATED)
  async handleVolumeLastUsedAtUpdated(event: VolumeLastUsedAtUpdatedEvent) {
    const dto = new VolumeDto(event.volume)
    this.notificationGateway.emitVolumeLastUsedAtUpdated(dto)
  }
}
