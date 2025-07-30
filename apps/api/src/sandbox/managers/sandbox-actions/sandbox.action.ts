/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { Sandbox } from '../../entities/sandbox.entity'
import { Repository } from 'typeorm'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { ToolboxService } from '../../services/toolbox.service'

export const SYNC_AGAIN = 'sync-again'
export const DONT_SYNC_AGAIN = 'dont-sync-again'
export type SyncState = typeof SYNC_AGAIN | typeof DONT_SYNC_AGAIN

@Injectable()
export abstract class SandboxAction {
  constructor(
    protected readonly runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    protected readonly sandboxRepository: Repository<Sandbox>,
    protected readonly toolboxService: ToolboxService,
  ) {}

  abstract run(sandbox: Sandbox): Promise<SyncState>

  protected async updateSandboxState(
    sandboxId: string,
    state: SandboxState,
    runnerId?: string | null | undefined,
    errorReason?: string,
    daemonVersion?: string,
  ) {
    const sandbox = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })
    if (sandbox.state === state && sandbox.runnerId === runnerId && sandbox.errorReason === errorReason) {
      return
    }

    sandbox.state = state

    if (runnerId !== undefined) {
      sandbox.runnerId = runnerId
    }

    if (errorReason !== undefined) {
      sandbox.errorReason = errorReason
    }

    if (daemonVersion !== undefined) {
      sandbox.daemonVersion = daemonVersion
    }

    await this.sandboxRepository.save(sandbox)
  }
}
