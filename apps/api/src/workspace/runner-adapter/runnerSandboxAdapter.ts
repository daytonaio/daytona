/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { Node } from '../entities/node.entity'
import { RunnerSandboxAdapterV2 } from './runnerSandboxAdapter.v2'
import { ModuleRef } from '@nestjs/core'
import { Workspace } from '../entities/workspace.entity'
import { RunnerSandboxAdapterV1 } from './runnerSandboxAdapter.v1'

export const COMPLETE_SYNC_TASK = 'complete'
export const RERUN_SYNC_TASK = 'rerun'
export type SyncTaskStatus = typeof COMPLETE_SYNC_TASK | typeof RERUN_SYNC_TASK

export interface RunnerSandboxAdapter {
  init(node: Node): Promise<void>
  syncInstanceState(workspace: Workspace): Promise<SyncTaskStatus>
}

@Injectable()
export class RunnerSandboxAdapterFactory {
  private readonly logger = new Logger(RunnerSandboxAdapterFactory.name)
  private moduleRef: ModuleRef

  async create(node: Node): Promise<RunnerSandboxAdapter> {
    switch (node.version) {
      case '1': {
        const adapter = await this.moduleRef.create(RunnerSandboxAdapterV1)
        await adapter.init(node)
        return adapter
      }
      case '2': {
        const adapter = await this.moduleRef.create(RunnerSandboxAdapterV2)
        await adapter.init(node)
        return adapter
      }
      default:
        throw new Error(`Unsupported runner version: ${node.version}`)
    }
  }
}
