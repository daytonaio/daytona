/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { Runner } from '../entities/runner.entity'
import { RunnerAdapterV2 } from './runnerAdapter.v2'
import { ModuleRef } from '@nestjs/core'
import { RunnerAdapterV1 } from './runnerAdapter.v1'
import { BuildInfo } from '../entities/build-info.entity'

export enum RunnerSandboxState {
  CREATING = 'creating',
  RESTORING = 'restoring',
  DESTROYED = 'destroyed',
  DESTROYING = 'destroying',
  STARTED = 'started',
  STOPPED = 'stopped',
  STARTING = 'starting',
  STOPPING = 'stopping',
  ERROR = 'error',
  BUILD_FAILED = 'build_failed',
  PENDING_BUILD = 'pending_build',
  BUILDING_SNAPSHOT = 'building_snapshot',
  UNKNOWN = 'unknown',
  PULLING_SNAPSHOT = 'pulling_snapshot',
  ARCHIVING = 'archiving',
  ARCHIVED = 'archived',
}

export interface RunnerSandboxInfo {
  state: RunnerSandboxState
  backupState?: string
}

export interface RunnerAdapter {
  init(runner: Runner): Promise<void>
  info(sandboxId: string): Promise<RunnerSandboxInfo>
  create(sandboxId: string): Promise<void>
  createBackup(sandboxId: string, backupSnapshotName: string): Promise<void>
  start(sandboxId: string): Promise<void>
  stop(sandboxId: string): Promise<void>
  destroy(sandboxId: string): Promise<void>
  removeDestroyed(sandboxId: string): Promise<void>
  snapshot(sandboxId: string, snapshotName: string): Promise<void>
  buildSnapshot(buildInfo: BuildInfo, organizationId?: string): Promise<void>
  pullSnapshot(snapshotName: string): Promise<void>
  snapshotExists(snapshotName: string): Promise<boolean>
}

@Injectable()
export class RunnerAdapterFactory {
  private readonly logger = new Logger(RunnerAdapterFactory.name)
  private moduleRef: ModuleRef

  async create(runner: Runner): Promise<RunnerAdapter> {
    switch (runner.version) {
      case '1': {
        const adapter = await this.moduleRef.create(RunnerAdapterV1)
        await adapter.init(runner)
        return adapter
      }
      case '2': {
        const adapter = await this.moduleRef.create(RunnerAdapterV2)
        await adapter.init(runner)
        return adapter
      }
      default:
        throw new Error(`Unsupported runner version: ${runner.version}`)
    }
  }
}
