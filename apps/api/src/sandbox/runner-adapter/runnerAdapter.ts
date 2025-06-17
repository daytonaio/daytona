/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { Runner } from '../entities/runner.entity'
import { RunnerAdapterV1 } from './runnerAdapter.v1'
import { ModuleRef } from '@nestjs/core'
import { RunnerAdapterLegacy } from './runnerAdapter.legacy'
import { BuildInfo } from '../entities/build-info.entity'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { BackupState } from '../enums/backup-state.enum'

export interface RunnerSandboxInfo {
  state: SandboxState
  backupState?: BackupState
}

export interface RunnerAdapter {
  init(runner: Runner): Promise<void>

  healthCheck(): Promise<void>

  info(sandboxId: string): Promise<RunnerSandboxInfo>
  create(sandbox: Sandbox, registry: DockerRegistry): Promise<void>
  createBackup(sandbox: Sandbox, backupSnapshotName: string, registry: DockerRegistry): Promise<void>
  start(sandboxId: string): Promise<void>
  stop(sandboxId: string): Promise<void>
  destroy(sandboxId: string): Promise<void>
  removeDestroyed(sandboxId: string): Promise<void>

  removeSnapshot(snapshotName: string, force: boolean): Promise<void>
  buildSnapshot(buildInfo: BuildInfo, organizationId?: string, registry?: DockerRegistry): Promise<void>
  pullSnapshot(snapshotName: string, registry: DockerRegistry): Promise<void>
  snapshotExists(snapshotName: string): Promise<boolean>
  getSnapshotLogs(snapshotRef: string, follow: boolean): Promise<string>
}

@Injectable()
export class RunnerAdapterFactory {
  private readonly logger = new Logger(RunnerAdapterFactory.name)
  private moduleRef: ModuleRef

  async create(runner: Runner): Promise<RunnerAdapter> {
    switch (runner.version) {
      case '0': {
        const adapter = await this.moduleRef.create(RunnerAdapterLegacy)
        await adapter.init(runner)
        return adapter
      }
      case '1': {
        const adapter = await this.moduleRef.create(RunnerAdapterV1)
        await adapter.init(runner)
        return adapter
      }
      default:
        throw new Error(`Unsupported runner version: ${runner.version}`)
    }
  }
}
