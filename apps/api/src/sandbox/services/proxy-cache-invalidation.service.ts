/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { InjectRedis } from '@nestjs-modules/ioredis'
import { Injectable, Logger } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import Redis from 'ioredis'

import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxArchivedEvent } from '../events/sandbox-archived.event'

@Injectable()
export class ProxyCacheInvalidationService {
  private readonly logger = new Logger(ProxyCacheInvalidationService.name)
  private static readonly RUNNER_INFO_CACHE_PREFIX = 'proxy:sandbox-runner-info:'

  constructor(@InjectRedis() private readonly redis: Redis) {}

  @OnEvent(SandboxEvents.ARCHIVED)
  async handleSandboxArchived(event: SandboxArchivedEvent): Promise<void> {
    await this.invalidateRunnerCache(event.sandbox.id)
  }

  private async invalidateRunnerCache(sandboxId: string): Promise<void> {
    try {
      await this.redis.del(`${ProxyCacheInvalidationService.RUNNER_INFO_CACHE_PREFIX}${sandboxId}`)
      this.logger.debug(`Invalidated sandbox runner cache for ${sandboxId}`)
    } catch (error) {
      this.logger.warn(`Failed to invalidate runner cache for sandbox ${sandboxId}: ${error.message}`)
    }
  }
}
