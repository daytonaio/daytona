/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { Runner } from '../entities/runner.entity'
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

export interface RunnerMetrics {
  currentAllocatedCpu?: number
  currentAllocatedDiskGiB?: number
  currentAllocatedMemoryGiB?: number
  currentCpuUsagePercentage?: number
  currentDiskUsagePercentage?: number
  currentMemoryUsagePercentage?: number
  currentSnapshotCount?: number
}

export interface RunnerInfo {
  metrics?: RunnerMetrics
}

export interface RunnerAdapter {
  init(runner: Runner): Promise<void>

  healthCheck(): Promise<void>

  runnerInfo(): Promise<RunnerInfo>

  sandboxInfo(sandboxId: string): Promise<RunnerSandboxInfo>
  createSandbox(sandbox: Sandbox, registry?: DockerRegistry, entrypoint?: string[]): Promise<void>
  startSandbox(sandboxId: string): Promise<void>
  stopSandbox(sandboxId: string): Promise<void>
  destroySandbox(sandboxId: string): Promise<void>
  removeDestroyedSandbox(sandboxId: string): Promise<void>
  createBackup(sandbox: Sandbox, backupSnapshotName: string, registry?: DockerRegistry): Promise<void>

  removeSnapshot(snapshotName: string): Promise<void>
  buildSnapshot(
    buildInfo: BuildInfo,
    organizationId?: string,
    registry?: DockerRegistry,
    pushToInternalRegistry?: boolean,
  ): Promise<void>
  pullSnapshot(snapshotName: string, registry?: DockerRegistry): Promise<void>
  snapshotExists(snapshotName: string): Promise<boolean>
  getSnapshotLogs(snapshotRef: string, follow: boolean): Promise<string>

  getSandboxDaemonVersion(sandboxId: string): Promise<string>
}

@Injectable()
export class RunnerAdapterFactory {
  private readonly logger = new Logger(RunnerAdapterFactory.name)

  constructor(private moduleRef: ModuleRef) {}

  async create(runner: Runner): Promise<RunnerAdapter> {
    switch (runner.version) {
      case '0': {
        const adapter = await this.moduleRef.create(RunnerAdapterLegacy)
        await adapter.init(runner)
        return adapter
      }
      default:
        throw new Error(`Unsupported runner version: ${runner.version}`)
    }
  }
}
