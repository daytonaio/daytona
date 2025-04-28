/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { WorkspaceUsagePeriod } from '../entities/workspace-usage-period.entity'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceStateUpdatedEvent } from '../../workspace/events/workspace-state-updated.event'
import { WorkspaceState } from '../../workspace/enums/workspace-state.enum'
import { WorkspaceEvents } from './../../workspace/constants/workspace-events.constants'

@Injectable()
export class UsageService {
  constructor(
    @InjectRepository(WorkspaceUsagePeriod)
    private workspaceUsagePeriodRepository: Repository<WorkspaceUsagePeriod>,
  ) {}

  @OnEvent(WorkspaceEvents.STATE_UPDATED)
  async handleWorkspaceStateUpdate(event: WorkspaceStateUpdatedEvent) {
    switch (event.newState) {
      case WorkspaceState.STARTED: {
        await this.closeUsagePeriod(event.workspace.id)
        await this.createUsagePeriod(event)
        break
      }
      case WorkspaceState.STOPPED:
        await this.closeUsagePeriod(event.workspace.id)
        await this.createUsagePeriod(event, true)
        break
      case WorkspaceState.ERROR:
      case WorkspaceState.ARCHIVED:
      case WorkspaceState.DESTROYED: {
        await this.closeUsagePeriod(event.workspace.id)
        break
      }
    }
  }

  private async createUsagePeriod(event: WorkspaceStateUpdatedEvent, diskOnly = false) {
    const usagePeriod = new WorkspaceUsagePeriod()
    usagePeriod.workspaceId = event.workspace.id
    usagePeriod.startAt = new Date()
    usagePeriod.endAt = null
    if (!diskOnly) {
      usagePeriod.cpu = event.workspace.cpu
      usagePeriod.gpu = event.workspace.gpu
      usagePeriod.mem = event.workspace.mem
    } else {
      usagePeriod.cpu = 0
      usagePeriod.gpu = 0
      usagePeriod.mem = 0
    }
    usagePeriod.disk = event.workspace.disk
    usagePeriod.organizationId = event.workspace.organizationId
    usagePeriod.region = event.workspace.region

    await this.workspaceUsagePeriodRepository.save(usagePeriod)
  }

  private async closeUsagePeriod(workspaceId: string) {
    const lastUsagePeriod = await this.workspaceUsagePeriodRepository.findOne({
      where: {
        workspaceId,
        endAt: null,
      },
      order: {
        startAt: 'DESC',
      },
    })

    if (lastUsagePeriod) {
      lastUsagePeriod.endAt = new Date()
      await this.workspaceUsagePeriodRepository.save(lastUsagePeriod)
    }
  }
}
