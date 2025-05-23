/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { NotificationGateway } from '../gateways/notification.gateway'
import { WorkspaceEvents } from '../../workspace/constants/workspace-events.constants'
import { WorkspaceDto } from '../../workspace/dto/workspace.dto'
import { WorkspaceCreatedEvent } from '../../workspace/events/workspace-create.event'
import { WorkspaceStateUpdatedEvent } from '../../workspace/events/workspace-state-updated.event'
import { RunnerService } from '../../workspace/services/runner.service'
import { SnapshotCreatedEvent } from '../../workspace/events/snapshot-created.event'
import { SnapshotEvents } from '../../workspace/constants/snapshot-events'
import { SnapshotDto } from '../../workspace/dto/snapshot.dto'
import { SnapshotStateUpdatedEvent } from '../../workspace/events/snapshot-state-updated.event'
import { SnapshotRemovedEvent } from '../../workspace/events/snapshot-removed.event'
import { SnapshotEnabledToggledEvent } from '../../workspace/events/snapshot-enabled-toggled.event'
import { VolumeEvents } from '../../workspace/constants/volume-events'
import { VolumeCreatedEvent } from '../../workspace/events/volume-created.event'
import { VolumeDto } from '../../workspace/dto/volume.dto'
import { VolumeStateUpdatedEvent } from '../../workspace/events/volume-state-updated.event'
import { VolumeLastUsedAtUpdatedEvent } from '../../workspace/events/volume-last-used-at-updated.event'

@Injectable()
export class NotificationService {
  constructor(
    private readonly notificationGateway: NotificationGateway,
    private readonly runnerService: RunnerService,
  ) {}

  @OnEvent(WorkspaceEvents.CREATED)
  async handleWorkspaceCreated(event: WorkspaceCreatedEvent) {
    const runner = await this.runnerService.findOne(event.workspace.runnerId)
    const dto = WorkspaceDto.fromWorkspace(event.workspace, runner.domain)
    this.notificationGateway.emitWorkspaceCreated(dto)
  }

  @OnEvent(WorkspaceEvents.STATE_UPDATED)
  async handleWorkspaceStateUpdated(event: WorkspaceStateUpdatedEvent) {
    const runner = await this.runnerService.findOne(event.workspace.runnerId)
    const dto = WorkspaceDto.fromWorkspace(event.workspace, runner.domain)
    this.notificationGateway.emitWorkspaceStateUpdated(dto, event.oldState, event.newState)
  }

  @OnEvent(SnapshotEvents.CREATED)
  async handleSnapshotCreated(event: SnapshotCreatedEvent) {
    const dto = SnapshotDto.fromSnapshot(event.snapshot)
    this.notificationGateway.emitSnapshotCreated(dto)
  }

  @OnEvent(SnapshotEvents.STATE_UPDATED)
  async handleSnapshotStateUpdated(event: SnapshotStateUpdatedEvent) {
    const dto = SnapshotDto.fromSnapshot(event.snapshot)
    this.notificationGateway.emitSnapshotStateUpdated(dto, event.oldState, event.newState)
  }

  @OnEvent(SnapshotEvents.ENABLED_TOGGLED)
  async handleSnapshotEnabledToggled(event: SnapshotEnabledToggledEvent) {
    const dto = SnapshotDto.fromSnapshot(event.snapshot)
    this.notificationGateway.emitSnapshotEnabledToggled(dto)
  }

  @OnEvent(SnapshotEvents.REMOVED)
  async handleSnapshotRemoved(event: SnapshotRemovedEvent) {
    const dto = SnapshotDto.fromSnapshot(event.snapshot)
    this.notificationGateway.emitSnapshotRemoved(dto)
  }

  @OnEvent(VolumeEvents.CREATED)
  async handleVolumeCreated(event: VolumeCreatedEvent) {
    const dto = VolumeDto.fromVolume(event.volume)
    this.notificationGateway.emitVolumeCreated(dto)
  }

  @OnEvent(VolumeEvents.STATE_UPDATED)
  async handleVolumeStateUpdated(event: VolumeStateUpdatedEvent) {
    const dto = VolumeDto.fromVolume(event.volume)
    this.notificationGateway.emitVolumeStateUpdated(dto, event.oldState, event.newState)
  }

  @OnEvent(VolumeEvents.LAST_USED_AT_UPDATED)
  async handleVolumeLastUsedAtUpdated(event: VolumeLastUsedAtUpdatedEvent) {
    const dto = VolumeDto.fromVolume(event.volume)
    this.notificationGateway.emitVolumeLastUsedAtUpdated(dto)
  }
}
