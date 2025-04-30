/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Inject, Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { In, MoreThan, Not, Repository } from 'typeorm'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { Workspace } from '../entities/workspace.entity'
import { WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION } from '../constants/workspace.constants'
import { WarmPool } from '../entities/warm-pool.entity'
import { EventEmitter2, OnEvent } from '@nestjs/event-emitter'
import { WorkspaceEvents } from '../constants/workspace-events.constants'
import { WorkspaceOrganizationUpdatedEvent } from '../events/workspace-organization-updated.event'
import { ConfigService } from '@nestjs/config'
import { Image } from '../entities/image.entity'
import { ImageState } from '../enums/image-state.enum'
import { NodeRegion } from '../enums/node-region.enum'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { Node } from '../entities/node.entity'
import { WarmPoolTopUpRequested } from '../events/warmpool-topup-requested.event'
import { WarmPoolEvents } from '../constants/warmpool-events.constants'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { WorkspaceDesiredState } from '../enums/workspace-desired-state.enum'

export type FetchWarmPoolWorkspaceParams = {
  image: string
  target: NodeRegion
  class: WorkspaceClass
  cpu: number
  mem: number
  disk: number
  osUser: string
  env: { [key: string]: string }
  organizationId: string
  state: string
}

@Injectable()
export class WorkspaceWarmPoolService {
  private readonly logger = new Logger(WorkspaceWarmPoolService.name)

  constructor(
    @InjectRepository(WarmPool)
    private readonly warmPoolRepository: Repository<WarmPool>,
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    @InjectRepository(Image)
    private readonly imageRepository: Repository<Image>,
    @InjectRepository(Node)
    private readonly nodeRepository: Repository<Node>,
    private readonly redisLockProvider: RedisLockProvider,
    private readonly configService: ConfigService,
    @Inject(EventEmitter2)
    private eventEmitter: EventEmitter2,
    @InjectRedis() private readonly redis: Redis,
  ) {}

  //  on init
  async onApplicationBootstrap() {
    //  await this.adHocSnapshotCheck()
  }

  async fetchWarmPoolWorkspace(params: FetchWarmPoolWorkspaceParams): Promise<Workspace | null> {
    //  validate image
    const workspaceImage = params.image || this.configService.get<string>('DEFAULT_IMAGE')
    const image = await this.imageRepository.findOne({
      where: [
        { organizationId: params.organizationId, name: workspaceImage, state: ImageState.ACTIVE },
        { general: true, name: workspaceImage, state: ImageState.ACTIVE },
      ],
    })
    if (!image) {
      throw new BadRequestError(`Image ${workspaceImage} not found or not accessible`)
    }

    //  check if workspace is warm pool
    const warmPoolItem = await this.warmPoolRepository.findOne({
      where: {
        image: image.name,
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
      const unschedulableNodes = await this.nodeRepository.find({
        where: {
          region: params.target,
          unschedulable: true,
        },
      })

      const warmPoolWorkspaces = await this.workspaceRepository.find({
        where: {
          nodeId: Not(In(unschedulableNodes.map((node) => node.id))),
          class: warmPoolItem.class,
          cpu: warmPoolItem.cpu,
          mem: warmPoolItem.mem,
          disk: warmPoolItem.disk,
          image: workspaceImage,
          osUser: warmPoolItem.osUser,
          env: warmPoolItem.env,
          organizationId: WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION,
          region: warmPoolItem.target,
          state: WorkspaceState.STARTED,
        },
        take: 10,
      })

      //  make sure we only release warm pool workspace once
      let warmPoolWorkspace: Workspace | null = null
      for (const workspace of warmPoolWorkspaces) {
        const lockKey = `workspace-warm-pool-${workspace.id}`
        if (await this.redis.get(lockKey)) {
          continue
        }
        await this.redis.setex(lockKey, 10, '1')

        warmPoolWorkspace = workspace
        break
      }

      return warmPoolWorkspace
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
        if (await this.redisLockProvider.lock(lockKey, 720)) {
          return
        }

        const workspaceCount = await this.workspaceRepository.count({
          where: {
            image: warmPoolItem.image,
            organizationId: WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION,
            class: warmPoolItem.class,
            osUser: warmPoolItem.osUser,
            env: warmPoolItem.env,
            region: warmPoolItem.target,
            cpu: warmPoolItem.cpu,
            gpu: warmPoolItem.gpu,
            mem: warmPoolItem.mem,
            disk: warmPoolItem.disk,
            desiredState: WorkspaceDesiredState.STARTED,
            state: Not(WorkspaceState.ERROR),
          },
        })

        const missingCount = warmPoolItem.pool - workspaceCount
        if (missingCount > 0) {
          this.logger.debug(`Creating ${missingCount} workspaces for warm pool id ${warmPoolItem.id}`)

          for (let i = 0; i < missingCount; i++) {
            this.eventEmitter.emit(WarmPoolEvents.TOPUP_REQUESTED, new WarmPoolTopUpRequested(warmPoolItem))
          }
        }

        await this.redisLockProvider.unlock(lockKey)
      }),
    )
  }

  @OnEvent(WorkspaceEvents.ORGANIZATION_UPDATED)
  async handleWorkspaceOrganizationUpdated(event: WorkspaceOrganizationUpdatedEvent) {
    if (event.newOrganizationId === WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION) {
      return
    }
    const warmPoolItem = await this.warmPoolRepository.findOne({
      where: {
        image: event.workspace.image,
        class: event.workspace.class,
        cpu: event.workspace.cpu,
        mem: event.workspace.mem,
        disk: event.workspace.disk,
        target: event.workspace.region,
        env: event.workspace.env,
        gpu: event.workspace.gpu,
        osUser: event.workspace.osUser,
      },
    })

    if (!warmPoolItem) {
      return
    }

    const workspaceCount = await this.workspaceRepository.count({
      where: {
        image: warmPoolItem.image,
        organizationId: WORKSPACE_WARM_POOL_UNASSIGNED_ORGANIZATION,
        class: warmPoolItem.class,
        osUser: warmPoolItem.osUser,
        env: warmPoolItem.env,
        region: warmPoolItem.target,
        cpu: warmPoolItem.cpu,
        gpu: warmPoolItem.gpu,
        mem: warmPoolItem.mem,
        disk: warmPoolItem.disk,
        desiredState: WorkspaceDesiredState.STARTED,
        state: Not(WorkspaceState.ERROR),
      },
    })

    if (warmPoolItem.pool <= workspaceCount) {
      return
    }

    if (warmPoolItem) {
      this.eventEmitter.emit(WarmPoolEvents.TOPUP_REQUESTED, new WarmPoolTopUpRequested(warmPoolItem))
    }
  }
}
