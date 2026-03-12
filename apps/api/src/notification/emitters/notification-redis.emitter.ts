/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, OnModuleInit } from '@nestjs/common'
import { Emitter } from '@socket.io/redis-emitter'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'
import { NotificationEmitter } from '../gateways/notification-emitter.abstract'
import { SandboxDto } from '../../sandbox/dto/sandbox.dto'
import { SandboxState } from '../../sandbox/enums/sandbox-state.enum'
import { SandboxDesiredState } from '../../sandbox/enums/sandbox-desired-state.enum'
import { SandboxEvents } from '../../sandbox/constants/sandbox-events.constants'
import { SnapshotDto } from '../../sandbox/dto/snapshot.dto'
import { SnapshotState } from '../../sandbox/enums/snapshot-state.enum'
import { SnapshotEvents } from '../../sandbox/constants/snapshot-events'
import { VolumeDto } from '../../sandbox/dto/volume.dto'
import { VolumeState } from '../../sandbox/enums/volume-state.enum'
import { VolumeEvents } from '../../sandbox/constants/volume-events'
import { RunnerDto } from '../../sandbox/dto/runner.dto'
import { RunnerState } from '../../sandbox/enums/runner-state.enum'
import { RunnerEvents } from '../../sandbox/constants/runner-events'

@Injectable()
export class NotificationRedisEmitter extends NotificationEmitter implements OnModuleInit {
  private readonly logger = new Logger(NotificationRedisEmitter.name)
  private emitter: Emitter

  constructor(@InjectRedis() private readonly redis: Redis) {
    super()
  }

  onModuleInit() {
    this.emitter = new Emitter(this.redis.duplicate())
    this.logger.debug('Socket.io Redis emitter initialized (publish-only)')
  }

  emitSandboxCreated(sandbox: SandboxDto) {
    this.emitter.to(sandbox.organizationId).emit(SandboxEvents.CREATED, sandbox)
  }

  emitSandboxStateUpdated(sandbox: SandboxDto, oldState: SandboxState, newState: SandboxState) {
    this.emitter.to(sandbox.organizationId).emit(SandboxEvents.STATE_UPDATED, { sandbox, oldState, newState })
  }

  emitSandboxDesiredStateUpdated(
    sandbox: SandboxDto,
    oldDesiredState: SandboxDesiredState,
    newDesiredState: SandboxDesiredState,
  ) {
    this.emitter
      .to(sandbox.organizationId)
      .emit(SandboxEvents.DESIRED_STATE_UPDATED, { sandbox, oldDesiredState, newDesiredState })
  }

  emitSnapshotCreated(snapshot: SnapshotDto) {
    this.emitter.to(snapshot.organizationId).emit(SnapshotEvents.CREATED, snapshot)
  }

  emitSnapshotStateUpdated(snapshot: SnapshotDto, oldState: SnapshotState, newState: SnapshotState) {
    this.emitter
      .to(snapshot.organizationId)
      .emit(SnapshotEvents.STATE_UPDATED, { snapshot: snapshot, oldState, newState })
  }

  emitSnapshotRemoved(snapshot: SnapshotDto) {
    this.emitter.to(snapshot.organizationId).emit(SnapshotEvents.REMOVED, snapshot)
  }

  emitVolumeCreated(volume: VolumeDto) {
    this.emitter.to(volume.organizationId).emit(VolumeEvents.CREATED, volume)
  }

  emitVolumeStateUpdated(volume: VolumeDto, oldState: VolumeState, newState: VolumeState) {
    this.emitter.to(volume.organizationId).emit(VolumeEvents.STATE_UPDATED, { volume, oldState, newState })
  }

  emitVolumeLastUsedAtUpdated(volume: VolumeDto) {
    this.emitter.to(volume.organizationId).emit(VolumeEvents.LAST_USED_AT_UPDATED, volume)
  }

  emitRunnerCreated(runner: RunnerDto, organizationId: string | null) {
    if (!organizationId) {
      return
    }
    this.emitter.to(organizationId).emit(RunnerEvents.CREATED, runner)
  }

  emitRunnerStateUpdated(
    runner: RunnerDto,
    organizationId: string | null,
    oldState: RunnerState,
    newState: RunnerState,
  ) {
    if (!organizationId) {
      return
    }
    this.emitter.to(organizationId).emit(RunnerEvents.STATE_UPDATED, { runner, oldState, newState })
  }

  emitRunnerUnschedulableUpdated(runner: RunnerDto, organizationId: string | null) {
    if (!organizationId) {
      return
    }
    this.emitter.to(organizationId).emit(RunnerEvents.UNSCHEDULABLE_UPDATED, runner)
  }
}
