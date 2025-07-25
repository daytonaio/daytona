/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SyncState, SYNC_AGAIN } from './sandbox.action'
import { BackupState } from '../../enums/backup-state.enum'
import { RunnerState } from '../../enums/runner-state.enum'
import { ToolboxService } from '../../services/toolbox.service'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { Repository } from 'typeorm'
import { InjectRepository } from '@nestjs/typeorm'

@Injectable()
export class SandboxStopAction extends SandboxAction {
  constructor(
    protected runnerService: RunnerService,
    protected runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    protected sandboxRepository: Repository<Sandbox>,
    protected toolboxService: ToolboxService,
  ) {
    super(runnerService, runnerAdapterFactory, sandboxRepository, toolboxService)
  }

  async run(sandbox: Sandbox): Promise<SyncState> {
    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      return DONT_SYNC_AGAIN
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    switch (sandbox.state) {
      case SandboxState.STARTED: {
        // stop sandbox
        await runnerAdapter.stopSandbox(sandbox.id)
        await this.updateSandboxState(sandbox.id, SandboxState.STOPPING)
        //  sync states again immediately for sandbox
        return SYNC_AGAIN
      }
      case SandboxState.STOPPING: {
        // check if sandbox is stopped
        const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
        switch (sandboxInfo.state) {
          case SandboxState.STOPPED: {
            const sandboxToUpdate = await this.sandboxRepository.findOneByOrFail({
              id: sandbox.id,
            })
            sandboxToUpdate.state = SandboxState.STOPPED
            sandboxToUpdate.backupState = BackupState.NONE
            await this.sandboxRepository.save(sandboxToUpdate)
            return SYNC_AGAIN
          }
          case SandboxState.ERROR: {
            await this.updateSandboxState(
              sandbox.id,
              SandboxState.ERROR,
              undefined,
              'Sandbox is in error state on runner',
            )
            return DONT_SYNC_AGAIN
          }
        }
        return SYNC_AGAIN
      }
      case SandboxState.ERROR: {
        const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
        if (sandboxInfo.state === SandboxState.STOPPED) {
          await this.updateSandboxState(sandbox.id, SandboxState.STOPPED)
        }
      }
    }

    return DONT_SYNC_AGAIN
  }
}
