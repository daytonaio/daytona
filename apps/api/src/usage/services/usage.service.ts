/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { IsNull, LessThan, Not, Repository } from 'typeorm'
import { SandboxUsagePeriod } from '../entities/sandbox-usage-period.entity'
import { OnEvent } from '@nestjs/event-emitter'
import { SandboxStateUpdatedEvent } from '../../sandbox/events/sandbox-state-updated.event'
import { SandboxState } from '../../sandbox/enums/sandbox-state.enum'
import { SandboxEvents } from './../../sandbox/constants/sandbox-events.constants'
import { Cron, CronExpression } from '@nestjs/schedule'
import { RedisLockProvider } from '../../sandbox/common/redis-lock.provider'
import { SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../../sandbox/constants/sandbox.constants'
import { Sandbox } from '../../sandbox/entities/sandbox.entity'

@Injectable()
export class UsageService {
  private readonly logger = new Logger(UsageService.name)

  constructor(
    @InjectRepository(SandboxUsagePeriod)
    private sandboxUsagePeriodRepository: Repository<SandboxUsagePeriod>,
    private readonly redisLockProvider: RedisLockProvider,
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
  ) {}

  @OnEvent(SandboxEvents.STATE_UPDATED)
  async handleSandboxStateUpdate(event: SandboxStateUpdatedEvent) {
    await this.waitForLock(event.sandbox.id)

    try {
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
        case SandboxState.BUILD_FAILED:
        case SandboxState.ARCHIVED:
        case SandboxState.DESTROYED: {
          await this.closeUsagePeriod(event.sandbox.id)
          break
        }
      }
    } finally {
      this.releaseLock(event.sandbox.id).catch((error) => {
        this.logger.error(`Error releasing lock for sandbox ${event.sandbox.id}`, error)
      })
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

  @Cron(CronExpression.EVERY_MINUTE, { name: 'close-and-reopen-usage-periods' })
  async closeAndReopenUsagePeriods() {
    if (!(await this.redisLockProvider.lock('close-and-reopen-usage-periods', 60))) {
      return
    }

    const usagePeriods = await this.sandboxUsagePeriodRepository.find({
      where: {
        endAt: IsNull(),
        // 1 day ago
        startAt: LessThan(new Date(Date.now() - 1000 * 60 * 60 * 24)),
        organizationId: Not(SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION),
      },
      order: {
        startAt: 'ASC',
      },
      take: 100,
    })

    for (const usagePeriod of usagePeriods) {
      if (!(await this.aquireLock(usagePeriod.sandboxId))) {
        continue
      }

      // validate that the usage period should remain active just in case
      try {
        const sandbox = await this.sandboxRepository.findOne({
          where: {
            id: usagePeriod.sandboxId,
          },
        })

        await this.sandboxUsagePeriodRepository.manager.transaction(async (transactionalEntityManager) => {
          // Close usage period
          const closeTime = new Date()
          usagePeriod.endAt = closeTime
          await transactionalEntityManager.save(usagePeriod)

          if (sandbox && (sandbox.state === SandboxState.STARTED || sandbox.state === SandboxState.STOPPED)) {
            // Create new usage period
            const newUsagePeriod = SandboxUsagePeriod.fromUsagePeriod(usagePeriod)
            newUsagePeriod.startAt = closeTime
            newUsagePeriod.endAt = null
            await transactionalEntityManager.save(newUsagePeriod)
          }
        })
      } catch (error) {
        this.logger.error(`Error closing and reopening usage period ${usagePeriod.sandboxId}`, error)
      } finally {
        await this.releaseLock(usagePeriod.sandboxId)
      }
    }

    await this.redisLockProvider.unlock('close-and-reopen-usage-periods')
  }

  private async waitForLock(sandboxId: string) {
    while (!(await this.aquireLock(sandboxId))) {
      await new Promise((resolve) => setTimeout(resolve, 500))
    }
  }

  private async aquireLock(sandboxId: string): Promise<boolean> {
    return await this.redisLockProvider.lock(`usage-period-${sandboxId}`, 60)
  }

  private async releaseLock(sandboxId: string) {
    await this.redisLockProvider.unlock(`usage-period-${sandboxId}`)
  }
}
