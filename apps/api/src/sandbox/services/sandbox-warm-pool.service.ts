/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject, Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { FindOptionsWhere, In, MoreThan, Not, Repository } from 'typeorm'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { Sandbox } from '../entities/sandbox.entity'
import { SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/sandbox.constants'
import { WarmPool } from '../entities/warm-pool.entity'
import { EventEmitter2, OnEvent } from '@nestjs/event-emitter'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxOrganizationUpdatedEvent } from '../events/sandbox-organization-updated.event'
import { ConfigService } from '@nestjs/config'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { SandboxState } from '../enums/sandbox-state.enum'
import { Runner } from '../entities/runner.entity'
import { WarmPoolTopUpRequested } from '../events/warmpool-topup-requested.event'
import { WarmPoolEvents } from '../constants/warmpool-events.constants'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { isValidUuid } from '../../common/utils/uuid'

export type FetchWarmPoolSandboxParams = {
  snapshot: string
  target: string
  class: SandboxClass
  cpu: number
  mem: number
  disk: number
  osUser: string
  env: { [key: string]: string }
  organizationId: string
  state: string
}

@Injectable()
export class SandboxWarmPoolService {
  private readonly logger = new Logger(SandboxWarmPoolService.name)

  constructor(
    @InjectRepository(WarmPool)
    private readonly warmPoolRepository: Repository<WarmPool>,
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly configService: ConfigService,
    @Inject(EventEmitter2)
    private eventEmitter: EventEmitter2,
    @InjectRedis() private readonly redis: Redis,
  ) {}

  //  on init
  async onApplicationBootstrap() {
    //  await this.adHocBackupCheck()
  }

  async fetchWarmPoolSandbox(params: FetchWarmPoolSandboxParams): Promise<Sandbox | null> {
    //  validate snapshot
    const sandboxSnapshot = params.snapshot || this.configService.get<string>('DEFAULT_SNAPSHOT')

    const snapshotFilter: FindOptionsWhere<Snapshot>[] = [
      { organizationId: params.organizationId, name: sandboxSnapshot, state: SnapshotState.ACTIVE },
      { general: true, name: sandboxSnapshot, state: SnapshotState.ACTIVE },
    ]

    if (isValidUuid(sandboxSnapshot)) {
      snapshotFilter.push(
        { organizationId: params.organizationId, id: sandboxSnapshot, state: SnapshotState.ACTIVE },
        { general: true, id: sandboxSnapshot, state: SnapshotState.ACTIVE },
      )
    }

    const snapshot = await this.snapshotRepository.findOne({
      where: snapshotFilter,
    })
    if (!snapshot) {
      throw new BadRequestError(`Snapshot ${sandboxSnapshot} not found. Did you add it through the Daytona Dashboard?`)
    }

    //  check if sandbox is warm pool
    const warmPoolItem = await this.warmPoolRepository.findOne({
      where: {
        snapshot: snapshot.name,
        target: params.target,
        class: params.class,
        cpu: params.cpu,
        mem: params.mem,
        disk: params.disk,
        osUser: params.osUser,
        env: params.env,
        pool: MoreThan(0),
      },
    })
    if (warmPoolItem) {
      const unschedulableRunners = await this.runnerRepository.find({
        where: {
          region: params.target,
          unschedulable: true,
        },
      })

      const warmPoolSandboxes = await this.sandboxRepository.find({
        where: {
          runnerId: Not(In(unschedulableRunners.map((runner) => runner.id))),
          class: warmPoolItem.class,
          cpu: warmPoolItem.cpu,
          mem: warmPoolItem.mem,
          disk: warmPoolItem.disk,
          snapshot: snapshot.name, // Use snapshot.name instead of sandboxSnapshot
          osUser: warmPoolItem.osUser,
          env: warmPoolItem.env,
          organizationId: SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION,
          region: warmPoolItem.target,
          state: SandboxState.STARTED,
        },
        take: 10,
      })

      //  make sure we only release warm pool sandbox once
      let warmPoolSandbox: Sandbox | null = null
      for (const sandbox of warmPoolSandboxes) {
        const lockKey = `sandbox-warm-pool-${sandbox.id}`
        if (!(await this.redisLockProvider.lock(lockKey, 10))) {
          continue
        }

        warmPoolSandbox = sandbox
        break
      }

      return warmPoolSandbox
    }

    return null
  }

  //  todo: make frequency configurable or more efficient
  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'warm-pool-check' })
  async warmPoolCheck(): Promise<void> {
    const warmPoolItems = await this.warmPoolRepository.find()

    await Promise.all(
      warmPoolItems.map(async (warmPoolItem) => {
        const lockKey = `warm-pool-lock-${warmPoolItem.id}`
        if (!(await this.redisLockProvider.lock(lockKey, 720))) {
          return
        }

        const sandboxCount = await this.sandboxRepository.count({
          where: {
            snapshot: warmPoolItem.snapshot,
            organizationId: SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION,
            class: warmPoolItem.class,
            osUser: warmPoolItem.osUser,
            env: warmPoolItem.env,
            region: warmPoolItem.target,
            cpu: warmPoolItem.cpu,
            gpu: warmPoolItem.gpu,
            mem: warmPoolItem.mem,
            disk: warmPoolItem.disk,
            desiredState: SandboxDesiredState.STARTED,
            state: Not(In([SandboxState.ERROR, SandboxState.BUILD_FAILED])),
          },
        })

        const missingCount = warmPoolItem.pool - sandboxCount
        if (missingCount > 0) {
          const promises = []
          this.logger.debug(`Creating ${missingCount} sandboxes for warm pool id ${warmPoolItem.id}`)

          for (let i = 0; i < missingCount; i++) {
            promises.push(
              this.eventEmitter.emitAsync(WarmPoolEvents.TOPUP_REQUESTED, new WarmPoolTopUpRequested(warmPoolItem)),
            )
          }

          // Wait for all promises to settle before releasing the lock. Otherwise, another worker could start creating sandboxes
          await Promise.allSettled(promises)
        }

        await this.redisLockProvider.unlock(lockKey)
      }),
    )
  }

  @OnEvent(SandboxEvents.ORGANIZATION_UPDATED)
  async handleSandboxOrganizationUpdated(event: SandboxOrganizationUpdatedEvent) {
    if (event.newOrganizationId === SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION) {
      return
    }
    const warmPoolItem = await this.warmPoolRepository.findOne({
      where: {
        snapshot: event.sandbox.snapshot,
        class: event.sandbox.class,
        cpu: event.sandbox.cpu,
        mem: event.sandbox.mem,
        disk: event.sandbox.disk,
        target: event.sandbox.region,
        env: event.sandbox.env,
        gpu: event.sandbox.gpu,
        osUser: event.sandbox.osUser,
      },
    })

    if (!warmPoolItem) {
      return
    }

    const sandboxCount = await this.sandboxRepository.count({
      where: {
        snapshot: warmPoolItem.snapshot,
        organizationId: SANDBOX_WARM_POOL_UNASSIGNED_ORGANIZATION,
        class: warmPoolItem.class,
        osUser: warmPoolItem.osUser,
        env: warmPoolItem.env,
        region: warmPoolItem.target,
        cpu: warmPoolItem.cpu,
        gpu: warmPoolItem.gpu,
        mem: warmPoolItem.mem,
        disk: warmPoolItem.disk,
        desiredState: SandboxDesiredState.STARTED,
        state: Not(In([SandboxState.ERROR, SandboxState.BUILD_FAILED])),
      },
    })

    if (warmPoolItem.pool <= sandboxCount) {
      return
    }

    if (warmPoolItem) {
      this.eventEmitter.emit(WarmPoolEvents.TOPUP_REQUESTED, new WarmPoolTopUpRequested(warmPoolItem))
    }
  }
}
