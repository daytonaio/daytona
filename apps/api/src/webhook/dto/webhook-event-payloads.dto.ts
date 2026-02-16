/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ApiProperty, ApiSchema } from '@nestjs/swagger'
import { WebhookEvent } from '../constants/webhook-events.constants'
import { SandboxState } from '../../sandbox/enums/sandbox-state.enum'
import { SandboxClass } from '../../sandbox/enums/sandbox-class.enum'
import { SnapshotState } from '../../sandbox/enums/snapshot-state.enum'
import { VolumeState } from '../../sandbox/enums/volume-state.enum'
import { SandboxCreatedEvent } from '../../sandbox/events/sandbox-create.event'
import { SandboxStateUpdatedEvent } from '../../sandbox/events/sandbox-state-updated.event'
import { SnapshotCreatedEvent } from '../../sandbox/events/snapshot-created.event'
import { SnapshotStateUpdatedEvent } from '../../sandbox/events/snapshot-state-updated.event'
import { SnapshotRemovedEvent } from '../../sandbox/events/snapshot-removed.event'
import { VolumeCreatedEvent } from '../../sandbox/events/volume-created.event'
import { VolumeStateUpdatedEvent } from '../../sandbox/events/volume-state-updated.event'

export abstract class BaseWebhookEventDto {
  @ApiProperty({
    description: 'Event type identifier',
    enum: WebhookEvent,
    enumName: 'WebhookEvent',
    example: 'sandbox.created',
  })
  event: string

  @ApiProperty({
    description: 'Timestamp when the event occurred',
    example: '2025-12-19T10:30:00.000Z',
    format: 'date-time',
  })
  timestamp: string
}

@ApiSchema({ name: 'SandboxCreatedWebhook' })
export class SandboxCreatedWebhookDto extends BaseWebhookEventDto {
  @ApiProperty({
    description: 'Sandbox ID',
    example: 'sandbox123',
  })
  id: string

  @ApiProperty({
    description: 'Organization ID',
    example: 'org123',
  })
  organizationId: string

  @ApiProperty({
    description: 'Sandbox state',
    enum: SandboxState,
    enumName: 'SandboxState',
  })
  state: SandboxState

  @ApiProperty({
    description: 'Sandbox class',
    enum: SandboxClass,
    enumName: 'SandboxClass',
  })
  class: SandboxClass

  @ApiProperty({
    description: 'When the sandbox was created',
    example: '2025-12-19T10:30:00.000Z',
    format: 'date-time',
  })
  createdAt: string

  static fromEvent(event: SandboxCreatedEvent, eventType: string): SandboxCreatedWebhookDto {
    return {
      event: eventType,
      timestamp: new Date().toISOString(),
      id: event.sandbox.id,
      organizationId: event.sandbox.organizationId,
      state: event.sandbox.state,
      class: event.sandbox.class,
      createdAt: event.sandbox.createdAt.toISOString(),
    }
  }
}

@ApiSchema({ name: 'SandboxStateUpdatedWebhook' })
export class SandboxStateUpdatedWebhookDto extends BaseWebhookEventDto {
  @ApiProperty({
    description: 'Sandbox ID',
    example: 'sandbox123',
  })
  id: string

  @ApiProperty({
    description: 'Organization ID',
    example: 'org123',
  })
  organizationId: string

  @ApiProperty({
    description: 'Previous state',
    enum: SandboxState,
    enumName: 'SandboxState',
  })
  oldState: SandboxState

  @ApiProperty({
    description: 'New state',
    enum: SandboxState,
    enumName: 'SandboxState',
  })
  newState: SandboxState

  @ApiProperty({
    description: 'When the sandbox was last updated',
    example: '2025-12-19T10:30:00.000Z',
    format: 'date-time',
  })
  updatedAt: string

  static fromEvent(event: SandboxStateUpdatedEvent, eventType: string): SandboxStateUpdatedWebhookDto {
    return {
      event: eventType,
      timestamp: new Date().toISOString(),
      id: event.sandbox.id,
      organizationId: event.sandbox.organizationId,
      oldState: event.oldState,
      newState: event.newState,
      updatedAt: event.sandbox.updatedAt.toISOString(),
    }
  }
}

@ApiSchema({ name: 'SnapshotCreatedWebhook' })
export class SnapshotCreatedWebhookDto extends BaseWebhookEventDto {
  @ApiProperty({
    description: 'Snapshot ID',
    example: 'snapshot123',
  })
  id: string

  @ApiProperty({
    description: 'Snapshot name',
    example: 'my-snapshot',
  })
  name: string

  @ApiProperty({
    description: 'Organization ID',
    example: 'org123',
  })
  organizationId: string

  @ApiProperty({
    description: 'Snapshot state',
    enum: SnapshotState,
    enumName: 'SnapshotState',
  })
  state: SnapshotState

  @ApiProperty({
    description: 'When the snapshot was created',
    example: '2025-12-19T10:30:00.000Z',
    format: 'date-time',
  })
  createdAt: string

  static fromEvent(event: SnapshotCreatedEvent, eventType: string): SnapshotCreatedWebhookDto {
    return {
      event: eventType,
      timestamp: new Date().toISOString(),
      id: event.snapshot.id,
      name: event.snapshot.name,
      organizationId: event.snapshot.organizationId,
      state: event.snapshot.state,
      createdAt: event.snapshot.createdAt.toISOString(),
    }
  }
}

@ApiSchema({ name: 'SnapshotStateUpdatedWebhook' })
export class SnapshotStateUpdatedWebhookDto extends BaseWebhookEventDto {
  @ApiProperty({
    description: 'Snapshot ID',
    example: 'snapshot123',
  })
  id: string

  @ApiProperty({
    description: 'Snapshot name',
    example: 'my-snapshot',
  })
  name: string

  @ApiProperty({
    description: 'Organization ID',
    example: 'org123',
  })
  organizationId: string

  @ApiProperty({
    description: 'Previous state',
    enum: SnapshotState,
    enumName: 'SnapshotState',
  })
  oldState: SnapshotState

  @ApiProperty({
    description: 'New state',
    enum: SnapshotState,
    enumName: 'SnapshotState',
  })
  newState: SnapshotState

  @ApiProperty({
    description: 'When the snapshot was last updated',
    example: '2025-12-19T10:30:00.000Z',
    format: 'date-time',
  })
  updatedAt: string

  static fromEvent(event: SnapshotStateUpdatedEvent, eventType: string): SnapshotStateUpdatedWebhookDto {
    return {
      event: eventType,
      timestamp: new Date().toISOString(),
      id: event.snapshot.id,
      name: event.snapshot.name,
      organizationId: event.snapshot.organizationId,
      oldState: event.oldState,
      newState: event.newState,
      updatedAt: event.snapshot.updatedAt.toISOString(),
    }
  }
}

@ApiSchema({ name: 'SnapshotRemovedWebhook' })
export class SnapshotRemovedWebhookDto extends BaseWebhookEventDto {
  @ApiProperty({
    description: 'Snapshot ID',
    example: 'snapshot123',
  })
  id: string

  @ApiProperty({
    description: 'Snapshot name',
    example: 'my-snapshot',
  })
  name: string

  @ApiProperty({
    description: 'Organization ID',
    example: 'org123',
  })
  organizationId: string

  @ApiProperty({
    description: 'When the snapshot was removed',
    example: '2025-12-19T10:30:00.000Z',
    format: 'date-time',
  })
  removedAt: string

  static fromEvent(event: SnapshotRemovedEvent, eventType: string): SnapshotRemovedWebhookDto {
    return {
      event: eventType,
      timestamp: new Date().toISOString(),
      id: event.snapshot.id,
      name: event.snapshot.name,
      organizationId: event.snapshot.organizationId,
      removedAt: new Date().toISOString(),
    }
  }
}

@ApiSchema({ name: 'VolumeCreatedWebhook' })
export class VolumeCreatedWebhookDto extends BaseWebhookEventDto {
  @ApiProperty({
    description: 'Volume ID',
    example: 'vol-12345678',
  })
  id: string

  @ApiProperty({
    description: 'Volume name',
    example: 'my-volume',
  })
  name: string

  @ApiProperty({
    description: 'Organization ID',
    example: 'org123',
  })
  organizationId: string

  @ApiProperty({
    description: 'Volume state',
    enum: VolumeState,
    enumName: 'VolumeState',
  })
  state: VolumeState

  @ApiProperty({
    description: 'When the volume was created',
    example: '2025-12-19T10:30:00.000Z',
    format: 'date-time',
  })
  createdAt: string

  static fromEvent(event: VolumeCreatedEvent, eventType: string): VolumeCreatedWebhookDto {
    return {
      event: eventType,
      timestamp: new Date().toISOString(),
      id: event.volume.id,
      name: event.volume.name,
      organizationId: event.volume.organizationId,
      state: event.volume.state,
      createdAt: event.volume.createdAt.toISOString(),
    }
  }
}

@ApiSchema({ name: 'VolumeStateUpdatedWebhook' })
export class VolumeStateUpdatedWebhookDto extends BaseWebhookEventDto {
  @ApiProperty({
    description: 'Volume ID',
    example: 'vol-12345678',
  })
  id: string

  @ApiProperty({
    description: 'Volume name',
    example: 'my-volume',
  })
  name: string

  @ApiProperty({
    description: 'Organization ID',
    example: 'org123',
  })
  organizationId: string

  @ApiProperty({
    description: 'Previous state',
    enum: VolumeState,
    enumName: 'VolumeState',
  })
  oldState: VolumeState

  @ApiProperty({
    description: 'New state',
    enum: VolumeState,
    enumName: 'VolumeState',
  })
  newState: VolumeState

  @ApiProperty({
    description: 'When the volume was last updated',
    example: '2025-12-19T10:30:00.000Z',
    format: 'date-time',
  })
  updatedAt: string

  static fromEvent(event: VolumeStateUpdatedEvent, eventType: string): VolumeStateUpdatedWebhookDto {
    return {
      event: eventType,
      timestamp: new Date().toISOString(),
      id: event.volume.id,
      name: event.volume.name,
      organizationId: event.volume.organizationId,
      oldState: event.oldState,
      newState: event.newState,
      updatedAt: event.volume.updatedAt.toISOString(),
    }
  }
}
