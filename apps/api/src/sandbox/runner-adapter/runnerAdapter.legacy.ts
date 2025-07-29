/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import axios from 'axios'
import axiosDebug from 'axios-debug-log'
import axiosRetry from 'axios-retry'

import { Injectable, Logger } from '@nestjs/common'
import { RunnerAdapter, RunnerInfo, RunnerSandboxInfo } from './runnerAdapter'
import { Runner } from '../entities/runner.entity'
import {
  Configuration,
  SandboxApi,
  EnumsSandboxState,
  SnapshotsApi,
  EnumsBackupState,
  DefaultApi,
  CreateSandboxDTO,
  BuildSnapshotRequestDTO,
  CreateBackupDTO,
  PullSnapshotRequestDTO,
  ToolboxApi,
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
  private sandboxApiClient: SandboxApi
  private snapshotApiClient: SnapshotsApi
  private runnerApiClient: DefaultApi
  private toolboxApiClient: ToolboxApi

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

    this.sandboxApiClient = new SandboxApi(new Configuration(), '', axiosInstance)
    this.snapshotApiClient = new SnapshotsApi(new Configuration(), '', axiosInstance)
    this.runnerApiClient = new DefaultApi(new Configuration(), '', axiosInstance)
    this.toolboxApiClient = new ToolboxApi(new Configuration(), '', axiosInstance)
  }

  async healthCheck(): Promise<void> {
    const response = await this.runnerApiClient.healthCheck()
    if (response.data.status !== 'ok') {
      throw new Error('Runner is not healthy')
    }
  }

  async runnerInfo(): Promise<RunnerInfo> {
    const response = await this.runnerApiClient.runnerInfo()
    return {
      metrics: response.data.metrics,
    }
  }

  async sandboxInfo(sandboxId: string): Promise<RunnerSandboxInfo> {
    const sandboxInfo = await this.sandboxApiClient.info(sandboxId)
    return {
      state: this.convertSandboxState(sandboxInfo.data.state),
      backupState: this.convertBackupState(sandboxInfo.data.backupState),
    }
  }

  async createSandbox(sandbox: Sandbox, registry?: DockerRegistry, entrypoint?: string[]): Promise<void> {
    const request: CreateSandboxDTO = {
      id: sandbox.id,
      snapshot: sandbox.snapshot,
      osUser: sandbox.osUser,
      userId: sandbox.organizationId,
      storageQuota: sandbox.disk,
      memoryQuota: sandbox.mem,
      cpuQuota: sandbox.cpu,
      env: sandbox.env,
      volumes: sandbox.volumes,
      entrypoint: entrypoint,
    }

    if (registry) {
      request.registry = {
        project: registry.project,
        url: registry.url,
        username: registry.username,
        password: registry.password,
      }
    }

    await this.sandboxApiClient.create(request)
  }

  async startSandbox(sandboxId: string): Promise<void> {
    await this.sandboxApiClient.start(sandboxId)
  }

  async stopSandbox(sandboxId: string): Promise<void> {
    await this.sandboxApiClient.stop(sandboxId)
  }

  async destroySandbox(sandboxId: string): Promise<void> {
    await this.sandboxApiClient.destroy(sandboxId)
  }

  async removeDestroyedSandbox(sandboxId: string): Promise<void> {
    await this.sandboxApiClient.removeDestroyed(sandboxId)
  }

  async createBackup(sandbox: Sandbox, backupSnapshotName: string, registry?: DockerRegistry): Promise<void> {
    const request: CreateBackupDTO = {
      snapshot: backupSnapshotName,
      registry: undefined,
    }

    if (registry) {
      request.registry = {
        project: registry.project,
        url: registry.url,
        username: registry.username,
        password: registry.password,
      }
    }

    await this.sandboxApiClient.createBackup(sandbox.id, request)
  }

  async buildSnapshot(
    buildInfo: BuildInfo,
    organizationId?: string,
    registry?: DockerRegistry,
    pushToInternalRegistry?: boolean,
  ): Promise<void> {
    const request: BuildSnapshotRequestDTO = {
      snapshot: buildInfo.snapshotRef,
      dockerfile: buildInfo.dockerfileContent,
      organizationId: organizationId,
      context: buildInfo.contextHashes,
      pushToInternalRegistry: pushToInternalRegistry,
    }

    if (registry) {
      request.registry = {
        project: registry.project,
        url: registry.url,
        username: registry.username,
        password: registry.password,
      }
    }

    await this.snapshotApiClient.buildSnapshot(request)
  }

  async removeSnapshot(snapshotName: string): Promise<void> {
    await this.snapshotApiClient.removeSnapshot(snapshotName)
  }

  async pullSnapshot(snapshotName: string, registry?: DockerRegistry): Promise<void> {
    const request: PullSnapshotRequestDTO = {
      snapshot: snapshotName,
    }

    if (registry) {
      request.registry = {
        project: registry.project,
        url: registry.url,
        username: registry.username,
        password: registry.password,
      }
    }

    await this.snapshotApiClient.pullSnapshot(request)
  }

  async snapshotExists(snapshotName: string): Promise<boolean> {
    const response = await this.snapshotApiClient.snapshotExists(snapshotName)
    return response.data.exists
  }

  async getSnapshotLogs(snapshotRef: string, follow: boolean): Promise<string> {
    const response = await this.snapshotApiClient.getBuildLogs(snapshotRef, follow)
    return response.data
  }

  async getSandboxDaemonVersion(sandboxId: string): Promise<string> {
    const getVersionResponse = await this.toolboxApiClient.sandboxesSandboxIdToolboxPathGet(sandboxId, 'version')
    if (!getVersionResponse.data || !(getVersionResponse.data as any).version) {
      throw new Error('Failed to get sandbox daemon version')
    }

    return (getVersionResponse.data as any).version
  }
}
