/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { NotificationGateway } from '../gateways/notification.gateway'
import { SandboxEvents } from '../../sandbox/constants/sandbox-events.constants'
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
import { RunnerEvents } from '../../sandbox/constants/runner-events'
import { RunnerDto } from '../../sandbox/dto/runner.dto'
import { RunnerCreatedEvent } from '../../sandbox/events/runner-created.event'
import { RunnerStateUpdatedEvent } from '../../sandbox/events/runner-state-updated.event'
import { RunnerUnschedulableUpdatedEvent } from '../../sandbox/events/runner-unschedulable-updated.event'
import { RegionService } from '../../region/services/region.service'
import { SandboxService } from '../../sandbox/services/sandbox.service'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { SANDBOX_EVENT_CHANNEL } from '../../common/constants/constants'

@Injectable()
export class NotificationService {
  constructor(
    private readonly notificationGateway: NotificationGateway,
    private readonly regionService: RegionService,
    private readonly sandboxService: SandboxService,
    @InjectRedis() private readonly redis: Redis,
  ) {}

  @OnEvent(SandboxEvents.CREATED)
  async handleSandboxCreated(event: SandboxCreatedEvent) {
    const dto = await this.sandboxService.toSandboxDto(event.sandbox)
    this.notificationGateway.emitSandboxCreated(dto)
  }

  @OnEvent(SandboxEvents.STATE_UPDATED)
  async handleSandboxStateUpdated(event: SandboxStateUpdatedEvent) {
    const dto = await this.sandboxService.toSandboxDto(event.sandbox)
    this.notificationGateway.emitSandboxStateUpdated(dto, event.oldState, event.newState)
    this.redis.publish(SANDBOX_EVENT_CHANNEL, JSON.stringify(event))
  }

  @OnEvent(SandboxEvents.DESIRED_STATE_UPDATED)
  async handleSandboxDesiredStateUpdated(event: SandboxDesiredStateUpdatedEvent) {
    const dto = await this.sandboxService.toSandboxDto(event.sandbox)
    this.notificationGateway.emitSandboxDesiredStateUpdated(dto, event.oldDesiredState, event.newDesiredState)
    this.redis.publish(SANDBOX_EVENT_CHANNEL, JSON.stringify(event))
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

  @OnEvent(RunnerEvents.CREATED)
  async handleRunnerCreated(event: RunnerCreatedEvent) {
    const dto = RunnerDto.fromRunner(event.runner)
    const organizationId = await this.regionService.getOrganizationId(event.runner.region)
    if (organizationId !== undefined) {
      this.notificationGateway.emitRunnerCreated(dto, organizationId)
    }
  }

  @OnEvent(RunnerEvents.STATE_UPDATED)
  async handleRunnerStateUpdated(event: RunnerStateUpdatedEvent) {
    const dto = RunnerDto.fromRunner(event.runner)
    const organizationId = await this.regionService.getOrganizationId(event.runner.region)
    if (organizationId !== undefined) {
      this.notificationGateway.emitRunnerStateUpdated(dto, organizationId, event.oldState, event.newState)
    }
  }

  @OnEvent(RunnerEvents.UNSCHEDULABLE_UPDATED)
  async handleRunnerUnschedulableUpdated(event: RunnerUnschedulableUpdatedEvent) {
    const dto = RunnerDto.fromRunner(event.runner)
    const organizationId = await this.regionService.getOrganizationId(event.runner.region)
    if (organizationId !== undefined) {
      this.notificationGateway.emitRunnerUnschedulableUpdated(dto, organizationId)
    }
  }
}
