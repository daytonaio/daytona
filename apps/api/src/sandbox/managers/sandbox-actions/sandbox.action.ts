/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { Sandbox } from '../../entities/sandbox.entity'
import { LockCode, RedisLockProvider } from '../../common/redis-lock.provider'

export const SYNC_AGAIN = 'sync-again'
export const DONT_SYNC_AGAIN = 'dont-sync-again'
export type SyncState = typeof SYNC_AGAIN | typeof DONT_SYNC_AGAIN

@Injectable()
export abstract class SandboxAction {
  protected readonly logger = new Logger(SandboxAction.name)

  constructor(
    protected readonly runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    protected readonly redisLockProvider: RedisLockProvider,
  ) {}

  abstract run(sandbox: Sandbox, lockCode: LockCode): Promise<SyncState>
}
