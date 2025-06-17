/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import axios from 'axios'
import axiosDebug from 'axios-debug-log'
import axiosRetry from 'axios-retry'

import { Injectable, Logger } from '@nestjs/common'
import { RunnerAdapter, RunnerSandboxInfo } from './runnerAdapter'
import { Runner } from '../entities/runner.entity'
import {
  Configuration,
  SandboxApi,
  EnumsSandboxState,
  SnapshotsApi,
  EnumsBackupState,
  DefaultApi,
} from '@daytonaio/runner-api-client'
import { Sandbox } from '../entities/sandbox.entity'
import { BuildInfo } from '../entities/build-info.entity'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { BackupState } from '../enums/backup-state.enum'

const isDebugEnabled = process.env.DEBUG === 'true'

@Injectable()
export class RunnerAdapterLegacy implements RunnerAdapter {
  private readonly logger = new Logger(RunnerAdapterLegacy.name)
  private apiClientSandbox: SandboxApi
  private apiClientSnapshot: SnapshotsApi
  private apiClient: DefaultApi

  private convertSandboxState(state: EnumsSandboxState): SandboxState {
    switch (state) {
      case EnumsSandboxState.SandboxStateCreating:
        return SandboxState.CREATING
      case EnumsSandboxState.SandboxStateRestoring:
        return SandboxState.RESTORING
      case EnumsSandboxState.SandboxStateDestroyed:
        return SandboxState.DESTROYED
      case EnumsSandboxState.SandboxStateDestroying:
        return SandboxState.DESTROYING
      case EnumsSandboxState.SandboxStateStarted:
        return SandboxState.STARTED
      case EnumsSandboxState.SandboxStateStopped:
        return SandboxState.STOPPED
      case EnumsSandboxState.SandboxStateStarting:
        return SandboxState.STARTING
      case EnumsSandboxState.SandboxStateStopping:
        return SandboxState.STOPPING
      case EnumsSandboxState.SandboxStateError:
        return SandboxState.ERROR
      case EnumsSandboxState.SandboxStatePullingSnapshot:
        return SandboxState.PULLING_SNAPSHOT
      default:
        return SandboxState.UNKNOWN
    }
  }

  private convertBackupState(state: EnumsBackupState): BackupState {
    switch (state) {
      case EnumsBackupState.BackupStatePending:
        return BackupState.PENDING
      case EnumsBackupState.BackupStateInProgress:
        return BackupState.IN_PROGRESS
      case EnumsBackupState.BackupStateCompleted:
        return BackupState.COMPLETED
      default:
        return BackupState.NONE
    }
  }

  public async init(runner: Runner): Promise<void> {
    const axiosInstance = axios.create({
      baseURL: runner.apiUrl,
      headers: {
        Authorization: `Bearer ${runner.apiKey}`,
      },
      timeout: 1 * 60 * 60 * 1000, // 1 hour
    })

    // Configure axios-retry to handle ECONNRESET errors
    axiosRetry(axiosInstance, {
      retries: 3,
      retryDelay: axiosRetry.exponentialDelay,
      retryCondition: (error) => {
        // Retry on ECONNRESET errors
        return (
          (error as any).code === 'ECONNRESET' ||
          error.message?.includes('ECONNRESET') ||
          (error as any).cause?.code === 'ECONNRESET'
        )
      },
      onRetry: (retryCount, error, requestConfig) => {
        this.logger.warn(
          `Retrying request due to ECONNRESET (attempt ${retryCount}): ${requestConfig.method?.toUpperCase()} ${requestConfig.url}`,
        )
      },
    })

    axiosInstance.interceptors.response.use(
      (response) => {
        return response
      },
      (error) => {
        const errorMessage = error.response?.data?.message || error.response?.data || error.message || String(error)

        throw new Error(String(errorMessage))
      },
    )

    if (isDebugEnabled) {
      axiosDebug.addLogger(axiosInstance)
    }

    this.apiClientSandbox = new SandboxApi(new Configuration(), '', axiosInstance)
    this.apiClientSnapshot = new SnapshotsApi(new Configuration(), '', axiosInstance)
  }

  async healthCheck(): Promise<void> {
    const response = await this.apiClient.healthCheck()
    if (response.data.status !== 'ok') {
      throw new Error('Runner is not healthy')
    }
  }

  async buildSnapshot(buildInfo: BuildInfo, organizationId?: string, registry?: DockerRegistry): Promise<void> {
    await this.apiClientSnapshot.buildSnapshot({
      snapshot: buildInfo.snapshotRef,
      registry: {
        project: registry.name,
        url: registry.url,
        username: registry.username,
        password: registry.password,
      },
      dockerfile: buildInfo.dockerfileContent,
      organizationId: organizationId,
      context: buildInfo.contextHashes,
    })
  }

  async create(sandbox: Sandbox, registry: DockerRegistry): Promise<void> {
    await this.apiClientSandbox.create({
      id: sandbox.id,
      snapshot: sandbox.snapshot,
      osUser: sandbox.osUser,
      userId: sandbox.organizationId,
      storageQuota: sandbox.disk,
      memoryQuota: sandbox.mem,
      cpuQuota: sandbox.cpu,
      env: sandbox.env,
      registry: {
        url: registry.url,
        username: registry.username,
        password: registry.password,
      },
    })
  }

  async createBackup(sandbox: Sandbox, backupSnapshotName: string, registry: DockerRegistry): Promise<void> {
    await this.apiClientSandbox.createBackup(sandbox.id, {
      registry: {
        url: registry.url,
        username: registry.username,
        password: registry.password,
      },
      snapshot: backupSnapshotName,
    })
  }

  async info(sandboxId: string): Promise<RunnerSandboxInfo> {
    const sandboxInfo = await this.apiClientSandbox.info(sandboxId)
    return {
      state: this.convertSandboxState(sandboxInfo.data.state),
      backupState: this.convertBackupState(sandboxInfo.data.backupState),
    }
  }

  async start(sandboxId: string): Promise<void> {
    await this.apiClientSandbox.start(sandboxId)
  }

  async stop(sandboxId: string): Promise<void> {
    await this.apiClientSandbox.stop(sandboxId)
  }

  async destroy(sandboxId: string): Promise<void> {
    await this.apiClientSandbox.destroy(sandboxId)
  }

  async removeDestroyed(sandboxId: string): Promise<void> {
    await this.apiClientSandbox.removeDestroyed(sandboxId)
  }

  async removeSnapshot(snapshotName: string, force: boolean): Promise<void> {
    await this.apiClientSnapshot.removeSnapshot(snapshotName)
  }

  async getSnapshotLogs(snapshotRef: string, follow: boolean): Promise<string> {
    const response = await this.apiClientSnapshot.getBuildLogs(snapshotRef, follow)
    return response.data
  }

  async snapshotExists(snapshotName: string): Promise<boolean> {
    const response = await this.apiClientSnapshot.snapshotExists(snapshotName)
    return response.data.exists
  }

  async pullSnapshot(snapshotName: string, registry: DockerRegistry): Promise<void> {
    await this.apiClientSnapshot.pullSnapshot({
      snapshot: snapshotName,
      registry: {
        url: registry.url,
        username: registry.username,
        password: registry.password,
      },
    })
  }
}
