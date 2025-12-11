/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { Runner } from '../entities/runner.entity'
import { ModuleRef } from '@nestjs/core'
import { RunnerAdapterV0 } from './runnerAdapter.v0'
import { RunnerAdapterV2 } from './runnerAdapter.v2'
import { BuildInfo } from '../entities/build-info.entity'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { Sandbox } from '../entities/sandbox.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { BackupState } from '../enums/backup-state.enum'

export interface RunnerSandboxInfo {
  state: SandboxState
  daemonVersion?: string
  backupState?: BackupState
  backupErrorReason?: string
}

export interface RunnerSnapshotInfo {
  name: string
  sizeGB: number
  entrypoint: string[]
  cmd: string[]
  hash: string
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
  appVersion?: string
}

export interface StartSandboxResponse {
  daemonVersion: string
}

export interface RunnerAdapter {
  init(runner: Runner): Promise<void>

  healthCheck(signal?: AbortSignal): Promise<void>

  runnerInfo(signal?: AbortSignal): Promise<RunnerInfo>

  sandboxInfo(sandboxId: string): Promise<RunnerSandboxInfo>
  createSandbox(
    sandbox: Sandbox,
    registry?: DockerRegistry,
    entrypoint?: string[],
    metadata?: { [key: string]: string },
  ): Promise<StartSandboxResponse | undefined>
  startSandbox(sandboxId: string, metadata?: { [key: string]: string }): Promise<StartSandboxResponse | undefined>
  stopSandbox(sandboxId: string): Promise<void>
  destroySandbox(sandboxId: string): Promise<void>
  removeDestroyedSandbox(sandboxId: string): Promise<void>
  createBackup(sandbox: Sandbox, backupSnapshotName: string, registry?: DockerRegistry): Promise<void>

  removeSnapshot(snapshotName: string): Promise<void>
  buildSnapshot(
    buildInfo: BuildInfo,
    organizationId?: string,
    sourceRegistries?: DockerRegistry[],
    registry?: DockerRegistry,
    pushToInternalRegistry?: boolean,
  ): Promise<void>
  pullSnapshot(
    snapshotName: string,
    registry?: DockerRegistry,
    destinationRegistry?: DockerRegistry,
    destinationRef?: string,
  ): Promise<void>
  tagImage(sourceImage: string, targetImage: string): Promise<void>
  snapshotExists(snapshotRef: string): Promise<boolean>
  getSnapshotInfo(snapshotName: string): Promise<RunnerSnapshotInfo>

  updateNetworkSettings(
    sandboxId: string,
    networkBlockAll?: boolean,
    networkAllowList?: string,
    networkLimitEgress?: boolean,
  ): Promise<void>
}

@Injectable()
export class RunnerAdapterFactory {
  private readonly logger = new Logger(RunnerAdapterFactory.name)

  constructor(private moduleRef: ModuleRef) {}

  async create(runner: Runner): Promise<RunnerAdapter> {
    switch (runner.apiVersion) {
      case '0': {
        const adapter = await this.moduleRef.create(RunnerAdapterV0)
        await adapter.init(runner)
        return adapter
      }
      case '2': {
        const adapter = await this.moduleRef.create(RunnerAdapterV2)
        await adapter.init(runner)
        return adapter
      }
      default:
        throw new Error(`Unsupported runner version: ${runner.apiVersion}`)
    }
  }
}
