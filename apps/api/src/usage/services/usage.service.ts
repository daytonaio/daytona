/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { IsNull, LessThan, Not, Repository } from 'typeorm'
import { WorkspaceUsagePeriod } from '../entities/workspace-usage-period.entity'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceStateUpdatedEvent } from '../../workspace/events/workspace-state-updated.event'
import { WorkspaceState } from '../../workspace/enums/workspace-state.enum'
import { WorkspaceEvents } from './../../workspace/constants/workspace-events.constants'
import { Cron, CronExpression } from '@nestjs/schedule'
import { RedisLockProvider } from '../../workspace/common/redis-lock.provider'
import { WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../../workspace/constants/workspace.constants'
@Injectable()
export class UsageService {
  private readonly logger = new Logger(UsageService.name)

  constructor(
    @InjectRepository(WorkspaceUsagePeriod)
    private workspaceUsagePeriodRepository: Repository<WorkspaceUsagePeriod>,
    private readonly redisLockProvider: RedisLockProvider,
  ) {}

  @OnEvent(WorkspaceEvents.STATE_UPDATED)
  async handleWorkspaceStateUpdate(event: WorkspaceStateUpdatedEvent) {
    await this.waitForLock(event.workspace.id)

    try {
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
    } finally {
      this.releaseLock(event.workspace.id).catch((error) => {
        this.logger.error(`Error releasing lock for workspace ${event.workspace.id}`, error)
      })
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

  @Cron(CronExpression.EVERY_MINUTE, { name: 'close-and-reopen-usage-periods' })
  async closeAndReopenUsagePeriods() {
    if (!(await this.redisLockProvider.lock('close-and-reopen-usage-periods', 60))) {
      return
    }

    const usagePeriods = await this.workspaceUsagePeriodRepository.find({
      where: {
        endAt: IsNull(),
        // 1 day ago
        startAt: LessThan(new Date(Date.now() - 1000 * 60 * 60 * 24)),
        organizationId: Not(WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION),
      },
      order: {
        startAt: 'ASC',
      },
      take: 100,
    })

    for (const usagePeriod of usagePeriods) {
      if (!(await this.aquireLock(usagePeriod.workspaceId))) {
        continue
      }

      try {
        await this.workspaceUsagePeriodRepository.manager.transaction(async (transactionalEntityManager) => {
          // Close usage period
          const closeTime = new Date()
          usagePeriod.endAt = closeTime
          await transactionalEntityManager.save(usagePeriod)

          // Create new usage period
          const newUsagePeriod = WorkspaceUsagePeriod.fromUsagePeriod(usagePeriod)
          newUsagePeriod.startAt = closeTime
          newUsagePeriod.endAt = null
          await transactionalEntityManager.save(newUsagePeriod)
        })
      } catch (error) {
        this.logger.error(`Error closing and reopening usage period ${usagePeriod.workspaceId}`, error)
      } finally {
        await this.releaseLock(usagePeriod.workspaceId)
      }
    }

    await this.redisLockProvider.unlock('close-and-reopen-usage-periods')
  }

  private async waitForLock(workspaceId: string) {
    while (!(await this.aquireLock(workspaceId))) {
      await new Promise((resolve) => setTimeout(resolve, 500))
    }
  }

  private async aquireLock(workspaceId: string): Promise<boolean> {
    return await this.redisLockProvider.lock(`usage-period-${workspaceId}`, 60)
  }

  private async releaseLock(workspaceId: string) {
    await this.redisLockProvider.unlock(`usage-period-${workspaceId}`)
  }
}
