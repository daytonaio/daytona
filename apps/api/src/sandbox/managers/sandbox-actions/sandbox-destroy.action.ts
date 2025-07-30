/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable } from '@nestjs/common'
import { Sandbox } from '../../entities/sandbox.entity'
import { SandboxState } from '../../enums/sandbox-state.enum'
import { DONT_SYNC_AGAIN, SandboxAction, SyncState, SYNC_AGAIN } from './sandbox.action'
import { RunnerState } from '../../enums/runner-state.enum'
import { ToolboxService } from '../../services/toolbox.service'
import { RunnerService } from '../../services/runner.service'
import { RunnerAdapterFactory } from '../../runner-adapter/runnerAdapter'
import { Repository } from 'typeorm'
import { InjectRepository } from '@nestjs/typeorm'

@Injectable()
export class SandboxDestroyAction extends SandboxAction {
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
    if (sandbox.state === SandboxState.ARCHIVED) {
      await this.updateSandboxState(sandbox.id, SandboxState.DESTROYED)
      return DONT_SYNC_AGAIN
    }

    const runner = await this.runnerService.findOne(sandbox.runnerId)
    if (runner.state !== RunnerState.READY) {
      return DONT_SYNC_AGAIN
    }

    const runnerAdapter = await this.runnerAdapterFactory.create(runner)

    switch (sandbox.state) {
      case SandboxState.DESTROYED:
        return DONT_SYNC_AGAIN
      case SandboxState.DESTROYING: {
        // check if sandbox is destroyed
        try {
          const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
          if (sandboxInfo.state === SandboxState.DESTROYED || sandboxInfo.state === SandboxState.ERROR) {
            await runnerAdapter.removeDestroyedSandbox(sandbox.id)
          }
        } catch (e) {
          //  if the sandbox is not found on runner, it is already destroyed
          if (!e.response || e.response.status !== 404) {
            throw e
          }
        }

        await this.updateSandboxState(sandbox.id, SandboxState.DESTROYED)
        return SYNC_AGAIN
      }
      default: {
        // destroy sandbox
        try {
          const sandboxInfo = await runnerAdapter.sandboxInfo(sandbox.id)
          if (sandboxInfo?.state === SandboxState.DESTROYED) {
            await this.updateSandboxState(sandbox.id, SandboxState.DESTROYING)
            return SYNC_AGAIN
          }
          await runnerAdapter.destroySandbox(sandbox.id)
        } catch (e) {
          //  if the sandbox is not found on runner, it is already destroyed
          if (e.response.status !== 404) {
            throw e
          }
        }
        await this.updateSandboxState(sandbox.id, SandboxState.DESTROYING)
        return SYNC_AGAIN
      }
    }
  }
}
