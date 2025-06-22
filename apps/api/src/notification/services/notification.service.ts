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
import { RunnerService } from '../../sandbox/services/runner.service'
import { SnapshotCreatedEvent } from '../../sandbox/events/snapshot-created.event'
import { SnapshotEvents } from '../../sandbox/constants/snapshot-events'
import { SnapshotDto } from '../../sandbox/dto/snapshot.dto'
import { SnapshotStateUpdatedEvent } from '../../sandbox/events/snapshot-state-updated.event'
import { SnapshotRemovedEvent } from '../../sandbox/events/snapshot-removed.event'
import { SnapshotEnabledToggledEvent } from '../../sandbox/events/snapshot-enabled-toggled.event'
import { VolumeEvents } from '../../sandbox/constants/volume-events'
import { VolumeCreatedEvent } from '../../sandbox/events/volume-created.event'
import { VolumeDto } from '../../sandbox/dto/volume.dto'
import { VolumeStateUpdatedEvent } from '../../sandbox/events/volume-state-updated.event'
import { VolumeLastUsedAtUpdatedEvent } from '../../sandbox/events/volume-last-used-at-updated.event'
import { SandboxDesiredStateUpdatedEvent } from '../../sandbox/events/sandbox-desired-state-updated.event'
import { AuditLogEvents } from '../../audit/constants/audit-log-events.constant'
import { AuditLogDto } from '../../audit/dto/audit-log.dto'
import { AuditLogCreatedEvent } from '../../audit/events/audit-log-created.event'
import { AuditLogUpdatedEvent } from '../../audit/events/audit-log-updated.event'

@Injectable()
export class NotificationService {
  constructor(
    private readonly notificationGateway: NotificationGateway,
    private readonly runnerService: RunnerService,
  ) {}

  @OnEvent(SandboxEvents.CREATED)
  async handleSandboxCreated(event: SandboxCreatedEvent) {
    const runner = await this.runnerService.findOne(event.sandbox.runnerId)
    const dto = SandboxDto.fromSandbox(event.sandbox, runner.domain)
    this.notificationGateway.emitSandboxCreated(dto)
  }

  @OnEvent(SandboxEvents.STATE_UPDATED)
  async handleSandboxStateUpdated(event: SandboxStateUpdatedEvent) {
    const runner = await this.runnerService.findOne(event.sandbox.runnerId)
    const dto = SandboxDto.fromSandbox(event.sandbox, runner.domain)
    this.notificationGateway.emitSandboxStateUpdated(dto, event.oldState, event.newState)
  }

  @OnEvent(SandboxEvents.DESIRED_STATE_UPDATED)
  async handleSandboxDesiredStateUpdated(event: SandboxDesiredStateUpdatedEvent) {
    const runner = await this.runnerService.findOne(event.sandbox.runnerId)
    const dto = SandboxDto.fromSandbox(event.sandbox, runner.domain)
    this.notificationGateway.emitSandboxDesiredStateUpdated(dto, event.oldDesiredState, event.newDesiredState)
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

  @OnEvent(AuditLogEvents.CREATED)
  async handleAuditLogCreated(event: AuditLogCreatedEvent) {
    const dto = AuditLogDto.fromAuditLog(event.auditLog)
    this.notificationGateway.emitAuditLogCreated(dto)
  }

  @OnEvent(AuditLogEvents.UPDATED)
  async handleAuditLogUpdated(event: AuditLogUpdatedEvent) {
    const dto = AuditLogDto.fromAuditLog(event.auditLog)
    this.notificationGateway.emitAuditLogUpdated(dto)
  }
}
