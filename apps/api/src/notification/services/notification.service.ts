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
import { NodeService } from '../../workspace/services/node.service'
import { ImageCreatedEvent } from '../../workspace/events/image-created.event'
import { ImageEvents } from '../../workspace/constants/image-events'
import { ImageDto } from '../../workspace/dto/image.dto'
import { ImageStateUpdatedEvent } from '../../workspace/events/image-state-updated.event'
import { ImageRemovedEvent } from '../../workspace/events/image-removed.event'
import { ImageEnabledToggledEvent } from '../../workspace/events/image-enabled-toggled.event'

@Injectable()
export class NotificationService {
  constructor(
    private readonly notificationGateway: NotificationGateway,
    private readonly nodeService: NodeService,
  ) {}

  @OnEvent(WorkspaceEvents.CREATED)
  async handleWorkspaceCreated(event: WorkspaceCreatedEvent) {
    const node = await this.nodeService.findOne(event.workspace.nodeId)
    const dto = WorkspaceDto.fromWorkspace(event.workspace, node.domain)
    this.notificationGateway.emitWorkspaceCreated(dto)
  }

  @OnEvent(WorkspaceEvents.STATE_UPDATED)
  async handleWorkspaceStateUpdated(event: WorkspaceStateUpdatedEvent) {
    const node = await this.nodeService.findOne(event.workspace.nodeId)
    const dto = WorkspaceDto.fromWorkspace(event.workspace, node.domain)
    this.notificationGateway.emitWorkspaceStateUpdated(dto, event.oldState, event.newState)
  }

  @OnEvent(ImageEvents.CREATED)
  async handleImageCreated(event: ImageCreatedEvent) {
    const dto = ImageDto.fromImage(event.image)
    this.notificationGateway.emitImageCreated(dto)
  }

  @OnEvent(ImageEvents.STATE_UPDATED)
  async handleImageStateUpdated(event: ImageStateUpdatedEvent) {
    const dto = ImageDto.fromImage(event.image)
    this.notificationGateway.emitImageStateUpdated(dto, event.oldState, event.newState)
  }

  @OnEvent(ImageEvents.ENABLED_TOGGLED)
  async handleImageEnabledToggled(event: ImageEnabledToggledEvent) {
    const dto = ImageDto.fromImage(event.image)
    this.notificationGateway.emitImageEnabledToggled(dto)
  }

  @OnEvent(ImageEvents.REMOVED)
  async handleImageRemoved(event: ImageRemovedEvent) {
    const dto = ImageDto.fromImage(event.image)
    this.notificationGateway.emitImageRemoved(dto)
  }
}
