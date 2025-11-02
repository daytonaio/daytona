/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { Sandbox } from '../../entities/sandbox.entity'
import { Repository } from 'typeorm'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { ToolboxService } from '../../services/toolbox.service'
import { BackupState } from '../../enums/backup-state.enum'

export const SYNC_AGAIN = 'sync-again'
export const DONT_SYNC_AGAIN = 'dont-sync-again'
export type SyncState = typeof SYNC_AGAIN | typeof DONT_SYNC_AGAIN

@Injectable()
export abstract class SandboxAction {
  protected readonly logger = new Logger(SandboxAction.name)

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
    const sandbox = await this.sandboxRepository.findOneBy({
      id: sandboxId,
      pending: true,
    })
    if (!sandbox) {
      //  this should never happen
      //  if it does, we need to log the error and return
      //  this indicates a concurrency error and should be investigated
      //  we don't to throw the error, just log it and return to avoid setting the error state
      //  on the otherwise ready sandbox
      const err = new Error(`sandbox ${sandboxId} is not in a pending state`)
      this.logger.error(err)
      return
    }

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

    if (sandbox.state === SandboxState.ERROR && !sandbox.errorReason) {
      sandbox.errorReason = 'Sandbox is in error state during update'
    }

    if (daemonVersion !== undefined) {
      sandbox.daemonVersion = daemonVersion
    }

    if (sandbox.state == SandboxState.DESTROYED) {
      sandbox.backupState = BackupState.NONE
    }

    await this.sandboxRepository.save(sandbox)
  }
}
