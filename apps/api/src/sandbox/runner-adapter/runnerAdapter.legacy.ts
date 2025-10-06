/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import axios from 'axios'
import axiosDebug from 'axios-debug-log'
import axiosRetry from 'axios-retry'

import { Logger } from '@nestjs/common'
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
  UpdateNetworkSettingsDTO,
} from '@daytonaio/runner-api-client'
import { Sandbox } from '../entities/sandbox.entity'
import { BuildInfo } from '../entities/build-info.entity'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { BackupState } from '../enums/backup-state.enum'

const isDebugEnabled = process.env.DEBUG === 'true'

export class RunnerAdapterLegacy implements RunnerAdapter {
  private readonly logger = new Logger(RunnerAdapterLegacy.name)
  private sandboxApiClient: SandboxApi
  private snapshotApiClient: SnapshotsApi
  private runnerApiClient: DefaultApi
  private toolboxApiClient: ToolboxApi

  constructor(runner: Runner) {
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
      onRetry: (retryCount, _, requestConfig) => {
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

  // This is required by the RunnerAdapter interface, but we don't need to do anything here
  // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
  public async init(_: Runner): Promise<void> {}

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
      case EnumsBackupState.BackupStateFailed:
        return BackupState.ERROR
      default:
        return BackupState.NONE
    }
  }

  async healthCheck(signal?: AbortSignal): Promise<void> {
    const response = await this.runnerApiClient.healthCheck({ signal })
    if (response.data.status !== 'ok') {
      throw new Error('Runner is not healthy')
    }
  }

  async runnerInfo(signal?: AbortSignal): Promise<RunnerInfo> {
    const response = await this.runnerApiClient.runnerInfo({ signal })
    return {
      metrics: response.data.metrics,
    }
  }

  async sandboxInfo(sandboxId: string): Promise<RunnerSandboxInfo> {
    const sandboxInfo = await this.sandboxApiClient.info(sandboxId)
    return {
      state: sandboxInfo.data.state ? this.convertSandboxState(sandboxInfo.data.state) : SandboxState.UNKNOWN,
      backupState: sandboxInfo.data.backupState
        ? this.convertBackupState(sandboxInfo.data.backupState)
        : BackupState.NONE,
      backupErrorReason: sandboxInfo.data.backupError,
    }
  }

  async createSandbox(
    sandbox: Sandbox,
    registry?: DockerRegistry,
    entrypoint?: string[],
    metadata?: { [key: string]: string },
  ): Promise<void> {
    if (!sandbox.snapshot) {
      throw new Error('Snapshot is required')
    }

    const createSandboxDto: CreateSandboxDTO = {
      id: sandbox.id,
      userId: sandbox.organizationId,
      snapshot: sandbox.snapshot,
      osUser: sandbox.osUser,
      cpuQuota: sandbox.cpu,
      gpuQuota: sandbox.gpu,
      memoryQuota: sandbox.mem,
      storageQuota: sandbox.disk,
      env: sandbox.env,
      registry: registry
        ? {
            project: registry.project,
            url: registry.url,
            username: registry.username,
            password: registry.password,
          }
        : undefined,
      entrypoint: entrypoint,
      volumes: sandbox.volumes?.map((volume) => ({
        volumeId: volume.volumeId,
        mountPath: volume.mountPath,
      })),
      networkBlockAll: sandbox.networkBlockAll,
      networkAllowList: sandbox.networkAllowList ?? undefined,
      metadata: metadata,
    }

    await this.sandboxApiClient.create(createSandboxDto)
  }

  async startSandbox(sandboxId: string, metadata?: { [key: string]: string }): Promise<void> {
    await this.sandboxApiClient.start(sandboxId, metadata)
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

  async createBackup(sandbox: Sandbox, backupSnapshotName: string, registry: DockerRegistry): Promise<void> {
    const request: CreateBackupDTO = {
      snapshot: backupSnapshotName,
      registry: {
        project: registry.project,
        url: registry.url,
        username: registry.username,
        password: registry.password,
      },
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
      // TODO: verify this
      dockerfile: buildInfo.dockerfileContent ?? '',
      organizationId: organizationId ?? '',
      context: buildInfo.contextHashes ?? [],
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
    return !!response.data.exists
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

  async updateNetworkSettings(
    sandboxId: string,
    networkBlockAll?: boolean,
    networkAllowList?: string,
    networkLimitEgress?: boolean,
  ): Promise<void> {
    const updateNetworkSettingsDto: UpdateNetworkSettingsDTO = {
      networkBlockAll: networkBlockAll,
      networkAllowList: networkAllowList,
      networkLimitEgress: networkLimitEgress,
    }

    await this.sandboxApiClient.updateNetworkSettings(sandboxId, updateNetworkSettingsDto)
  }
}
