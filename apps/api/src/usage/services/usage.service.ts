/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { SandboxUsagePeriod } from '../entities/sandbox-usage-period.entity'
import { OnEvent } from '@nestjs/event-emitter'
import { SandboxStateUpdatedEvent } from '../../sandbox/events/sandbox-state-updated.event'
import { SandboxState } from '../../sandbox/enums/sandbox-state.enum'
import { SandboxEvents } from './../../sandbox/constants/sandbox-events.constants'

@Injectable()
export class UsageService {
  constructor(
    @InjectRepository(SandboxUsagePeriod)
    private sandboxUsagePeriodRepository: Repository<SandboxUsagePeriod>,
  ) {}

  @OnEvent(SandboxEvents.STATE_UPDATED)
  async handleSandboxStateUpdate(event: SandboxStateUpdatedEvent) {
    switch (event.newState) {
      case SandboxState.STARTED: {
        await this.closeUsagePeriod(event.sandbox.id)
        await this.createUsagePeriod(event)
        break
      }
      case SandboxState.STOPPED:
        await this.closeUsagePeriod(event.sandbox.id)
        await this.createUsagePeriod(event, true)
        break
      case SandboxState.ERROR:
      case SandboxState.ARCHIVED:
      case SandboxState.DESTROYED: {
        await this.closeUsagePeriod(event.sandbox.id)
        break
      }
    }
  }

  private async createUsagePeriod(event: SandboxStateUpdatedEvent, diskOnly = false) {
    const usagePeriod = new SandboxUsagePeriod()
    usagePeriod.sandboxId = event.sandbox.id
    usagePeriod.startAt = new Date()
    usagePeriod.endAt = null
    if (!diskOnly) {
      usagePeriod.cpu = event.sandbox.cpu
      usagePeriod.gpu = event.sandbox.gpu
      usagePeriod.mem = event.sandbox.mem
    } else {
      usagePeriod.cpu = 0
      usagePeriod.gpu = 0
      usagePeriod.mem = 0
    }
    usagePeriod.disk = event.sandbox.disk
    usagePeriod.organizationId = event.sandbox.organizationId
    usagePeriod.region = event.sandbox.region

    await this.sandboxUsagePeriodRepository.save(usagePeriod)
  }

  private async closeUsagePeriod(sandboxId: string) {
    const lastUsagePeriod = await this.sandboxUsagePeriodRepository.findOne({
      where: {
        sandboxId,
        endAt: null,
      },
      order: {
        startAt: 'DESC',
      },
    })

    if (lastUsagePeriod) {
      lastUsagePeriod.endAt = new Date()
      await this.sandboxUsagePeriodRepository.save(lastUsagePeriod)
    }
  }
}
