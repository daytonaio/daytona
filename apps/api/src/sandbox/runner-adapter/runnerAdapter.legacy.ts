/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import axios, { AxiosError } from 'axios'
import axiosDebug from 'axios-debug-log'
import axiosRetry from 'axios-retry'

import { Injectable, Logger } from '@nestjs/common'
import { RunnerAdapter, RunnerInfo, RunnerSandboxInfo, RunnerSnapshotInfo } from './runnerAdapter'
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
  RecoverSandboxDTO,
  IsRecoverableDTO,
} from '@daytonaio/runner-api-client'
import { Sandbox } from '../entities/sandbox.entity'
import { BuildInfo } from '../entities/build-info.entity'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { BackupState } from '../enums/backup-state.enum'

const isDebugEnabled = process.env.DEBUG === 'true'

// Network error codes that should trigger a retry
const RETRYABLE_NETWORK_ERROR_CODES = ['ECONNRESET', 'ETIMEDOUT']

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
      case EnumsBackupState.BackupStateFailed:
        return BackupState.ERROR
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

    const retryErrorMap = new WeakMap<AxiosError, string>()

    // Configure axios-retry to handle network errors
    axiosRetry(axiosInstance, {
      retries: 3,
      retryDelay: axiosRetry.exponentialDelay,
      retryCondition: (error) => {
        // Check if error code or message matches any retryable error
        const matchedErrorCode = RETRYABLE_NETWORK_ERROR_CODES.find(
          (code) =>
            (error as any).code === code || error.message?.includes(code) || (error as any).cause?.code === code,
        )

        if (matchedErrorCode) {
          retryErrorMap.set(error, matchedErrorCode)
          return true
        }

        return false
      },
      onRetry: (retryCount, error, requestConfig) => {
        this.logger.warn(
          `Retrying request due to ${retryErrorMap.get(error)} (attempt ${retryCount}): ${requestConfig.method?.toUpperCase()} ${requestConfig.url}`,
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
      state: this.convertSandboxState(sandboxInfo.data.state),
      backupState: this.convertBackupState(sandboxInfo.data.backupState),
      backupErrorReason: sandboxInfo.data.backupError,
    }
  }

  async createSandbox(
    sandbox: Sandbox,
    registry?: DockerRegistry,
    entrypoint?: string[],
    metadata?: { [key: string]: string },
    otelEndpoint?: string,
  ): Promise<void> {
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
            url: registry.url.replace(/^(https?:\/\/)/, ''),
            username: registry.username,
            password: registry.password,
          }
        : undefined,
      entrypoint: entrypoint,
      volumes: sandbox.volumes?.map((volume) => ({
        volumeId: volume.volumeId,
        mountPath: volume.mountPath,
        subpath: volume.subpath,
      })),
      networkBlockAll: sandbox.networkBlockAll,
      networkAllowList: sandbox.networkAllowList,
      metadata: metadata,
      authToken: sandbox.authToken,
      otelEndpoint,
    }

    await this.sandboxApiClient.create(createSandboxDto)
  }

  async startSandbox(sandboxId: string, authToken: string, metadata?: { [key: string]: string }): Promise<void> {
    await this.sandboxApiClient.start(sandboxId, authToken, metadata)
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
        url: registry.url.replace(/^(https?:\/\/)/, ''),
        username: registry.username,
        password: registry.password,
      }
    }

    await this.sandboxApiClient.createBackup(sandbox.id, request)
  }

  async buildSnapshot(
    buildInfo: BuildInfo,
    organizationId?: string,
    sourceRegistries?: DockerRegistry[],
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

    if (sourceRegistries) {
      request.sourceRegistries = sourceRegistries.map((sourceRegistry) => ({
        project: sourceRegistry.project,
        url: sourceRegistry.url.replace(/^(https?:\/\/)/, ''),
        username: sourceRegistry.username,
        password: sourceRegistry.password,
      }))
    }

    if (registry) {
      request.registry = {
        project: registry.project,
        url: registry.url.replace(/^(https?:\/\/)/, ''),
        username: registry.username,
        password: registry.password,
      }
    }

    await this.snapshotApiClient.buildSnapshot(request)
  }

  async removeSnapshot(snapshotName: string): Promise<void> {
    await this.snapshotApiClient.removeSnapshot(snapshotName)
  }

  async pullSnapshot(
    snapshotName: string,
    registry?: DockerRegistry,
    destinationRegistry?: DockerRegistry,
    destinationRef?: string,
  ): Promise<void> {
    const request: PullSnapshotRequestDTO = {
      snapshot: snapshotName,
    }

    if (registry) {
      request.registry = {
        project: registry.project,
        url: registry.url.replace(/^(https?:\/\/)/, ''),
        username: registry.username,
        password: registry.password,
      }
    }

    if (destinationRegistry) {
      request.destinationRegistry = {
        project: destinationRegistry.project,
        url: destinationRegistry.url.replace(/^(https?:\/\/)/, ''),
        username: destinationRegistry.username,
        password: destinationRegistry.password,
      }
    }

    if (destinationRef) {
      request.destinationRef = destinationRef
    }

    await this.snapshotApiClient.pullSnapshot(request)
  }

  async tagImage(sourceImage: string, targetImage: string): Promise<void> {
    await this.snapshotApiClient.tagImage({ sourceImage, targetImage })
  }

  async snapshotExists(snapshotName: string): Promise<boolean> {
    const response = await this.snapshotApiClient.snapshotExists(snapshotName)
    return response.data.exists
  }

  async getSnapshotInfo(snapshotName: string): Promise<RunnerSnapshotInfo> {
    const response = await this.snapshotApiClient.getSnapshotInfo(snapshotName)
    return {
      name: response.data.name || '',
      sizeGB: response.data.sizeGB,
      entrypoint: response.data.entrypoint,
      cmd: response.data.cmd,
      hash: response.data.hash,
    }
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

  async recover(sandbox: Sandbox): Promise<void> {
    const recoverSandboxDTO: RecoverSandboxDTO = {
      userId: sandbox.organizationId,
      snapshot: sandbox.snapshot,
      osUser: sandbox.osUser,
      cpuQuota: sandbox.cpu,
      gpuQuota: sandbox.gpu,
      memoryQuota: sandbox.mem,
      storageQuota: sandbox.disk,
      env: sandbox.env,
      volumes: sandbox.volumes?.map((volume) => ({
        volumeId: volume.volumeId,
        mountPath: volume.mountPath,
        subpath: volume.subpath,
      })),
      networkBlockAll: sandbox.networkBlockAll,
      networkAllowList: sandbox.networkAllowList,
      errorReason: sandbox.errorReason,
      backupErrorReason: sandbox.backupErrorReason,
    }
    await this.sandboxApiClient.recover(sandbox.id, recoverSandboxDTO)
  }

  async isRecoverable(sandboxId: string, errorReason: string): Promise<boolean> {
    const isRecoverableDTO: IsRecoverableDTO = { errorReason }

    const response = await this.sandboxApiClient.isRecoverable(sandboxId, isRecoverableDTO)
    return response.data.recoverable
  }
}
