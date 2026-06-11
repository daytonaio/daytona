/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import axios, { AxiosError } from 'axios'
import axiosDebug from 'axios-debug-log'
import axiosRetry from 'axios-retry'

import { Injectable, Logger } from '@nestjs/common'
import {
  CreateSandboxSnapshotResult,
  RunnerAdapter,
  RunnerInfo,
  RunnerSandboxInfo,
  RunnerSnapshotInfo,
  StartSandboxResponse,
  SnapshotDigestResponse,
} from './runnerAdapter'
import { SnapshotStateError } from '../errors/snapshot-state-error'
import { Runner } from '../entities/runner.entity'
import {
  Configuration,
  SandboxApi,
  EnumsSandboxState,
  SnapshotsApi,
  EnumsBackupState,
  EnumsSnapshotFromSandboxState,
  DefaultApi,
  CreateSandboxDTO,
  BuildSnapshotRequestDTO,
  CreateBackupDTO,
  PullSnapshotRequestDTO,
  SnapshotFromSandboxStatusResponse,
  ToolboxApi,
  UpdateNetworkSettingsDTO,
  RecoverSandboxDTO,
} from '@daytona/runner-api-client'
import { Sandbox } from '../entities/sandbox.entity'
import { BuildInfo } from '../entities/build-info.entity'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { BackupState } from '../enums/backup-state.enum'
import { RunnerApiError } from '../errors/runner-api-error'
import { TypedConfigService } from '../../config/typed-config.service'

const isDebugEnabled = process.env.DEBUG === 'true'

// Network error codes that should trigger a retry
const RETRYABLE_NETWORK_ERROR_CODES = ['ECONNRESET', 'ETIMEDOUT']

// Per-request timeout for snapshot capture status polls; keeps one hung
// connection from eating the whole poll budget (the instance-wide axios
// timeout is 15 minutes).
const SNAPSHOT_STATUS_POLL_TIMEOUT_MS = 30 * 1_000

@Injectable()
export class RunnerAdapterV0 implements RunnerAdapter {
  private readonly logger = new Logger(RunnerAdapterV0.name)
  private sandboxApiClient: SandboxApi
  private snapshotApiClient: SnapshotsApi
  private runnerApiClient: DefaultApi
  private toolboxApiClient: ToolboxApi

  constructor(private readonly configService: TypedConfigService) {}

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
    if (!runner.apiUrl) {
      throw new Error('Runner API URL is required')
    }

    const axiosInstance = axios.create({
      baseURL: runner.apiUrl,
      headers: {
        Authorization: `Bearer ${runner.apiKey}`,
      },
      timeout: 15 * 60 * 1000, // 15 minutes
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
        const statusCode = error.response?.data?.statusCode || error.response?.status || error.status
        const code = error.response?.data?.code || (error as any).code || (error as any).cause?.code || ''

        throw new RunnerApiError(String(errorMessage), statusCode, code)
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
      serviceHealth: response.data.serviceHealth,
      metrics: response.data.metrics,
      appVersion: response.data.appVersion,
    }
  }

  async sandboxInfo(sandboxId: string): Promise<RunnerSandboxInfo> {
    const sandboxInfo = await this.sandboxApiClient.info(sandboxId)
    return {
      state: this.convertSandboxState(sandboxInfo.data.state),
      backupState: this.convertBackupState(sandboxInfo.data.backupState),
      backupSnapshot: sandboxInfo.data.backupSnapshot,
      backupErrorReason: sandboxInfo.data.backupError,
      recoverable: sandboxInfo.data.recoverable,
      daemonVersion: sandboxInfo.data.daemonVersion,
    }
  }

  async createSandbox(
    sandbox: Sandbox,
    snapshotRef: string,
    registry?: DockerRegistry,
    entrypoint?: string[],
    metadata?: { [key: string]: string },
    otelEndpoint?: string,
    skipStart?: boolean,
  ): Promise<StartSandboxResponse | undefined> {
    const createSandboxDto: CreateSandboxDTO = {
      id: sandbox.id,
      name: sandbox.name,
      userId: sandbox.organizationId,
      snapshot: snapshotRef,
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
      skipStart: skipStart,
      organizationId: sandbox.organizationId,
      regionId: sandbox.region,
      linkedSandboxId: sandbox.linkedSandboxId ?? undefined,
      sandboxClass: sandbox.sandboxClass,
    }

    const response = await this.sandboxApiClient.create(createSandboxDto)

    if (!response?.data?.daemonVersion) {
      return undefined
    }

    return {
      daemonVersion: response.data.daemonVersion,
    }
  }

  async startSandbox(
    sandboxId: string,
    authToken: string,
    metadata?: { [key: string]: string },
  ): Promise<StartSandboxResponse | undefined> {
    const response = await this.sandboxApiClient.start(sandboxId, authToken, metadata)

    if (!response?.data?.daemonVersion) {
      return undefined
    }

    return {
      daemonVersion: response.data.daemonVersion,
    }
  }

  async stopSandbox(sandboxId: string, force?: boolean): Promise<void> {
    await this.sandboxApiClient.stop(sandboxId, { force })
  }

  async destroySandbox(sandboxId: string): Promise<void> {
    await this.sandboxApiClient.destroy(sandboxId)
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
    newTag?: string,
  ): Promise<void> {
    const request: PullSnapshotRequestDTO = {
      snapshot: snapshotName,
      newTag,
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

  async snapshotExists(snapshotName: string): Promise<boolean> {
    const response = await this.snapshotApiClient.snapshotExists(snapshotName)
    return response.data.exists
  }

  async getSnapshotInfo(snapshotName: string): Promise<RunnerSnapshotInfo> {
    try {
      const response = await this.snapshotApiClient.getSnapshotInfo(snapshotName)

      return {
        name: response.data.name || '',
        sizeGB: response.data.sizeGB,
        entrypoint: response.data.entrypoint,
        cmd: response.data.cmd,
        hash: response.data.hash,
      }
    } catch (err) {
      if (err instanceof RunnerApiError && err.statusCode === 422) {
        throw new SnapshotStateError(err.message)
      }
      throw err
    }
  }

  async inspectSnapshotInRegistry(snapshotName: string, registry?: DockerRegistry): Promise<SnapshotDigestResponse> {
    const response = await this.snapshotApiClient.inspectSnapshotInRegistry({
      snapshot: snapshotName,
      registry: registry
        ? {
            project: registry.project,
            url: registry.url.replace(/^(https?:\/\/)/, ''),
            username: registry.username,
            password: registry.password,
          }
        : undefined,
    })

    return {
      hash: response.data.hash,
      sizeGB: response.data.sizeGB,
    }
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

  async forkSandbox(_sourceSandboxId: string, _newSandboxId: string): Promise<void> {
    throw new Error('forkSandbox is not supported for V0 runners')
  }

  async createSnapshotFromSandbox(
    sandboxId: string,
    snapshotName: string,
    organizationId: string,
    registry?: DockerRegistry,
    _includeMemory?: boolean,
  ): Promise<CreateSandboxSnapshotResult> {
    if (!registry) {
      throw new Error('registry is required to snapshot a Docker sandbox')
    }

    const response = await this.sandboxApiClient.snapshotFromSandbox(sandboxId, {
      name: snapshotName,
      organizationId,
      async: true,
      registry: {
        project: registry.project,
        url: registry.url.replace(/^(https?:\/\/)/, ''),
        username: registry.username,
        password: registry.password,
      },
    })

    if (response.status === 200) {
      // Old runner (deploy window): the capture ran synchronously and the
      // payload is in the response. Still bounded by the instance timeout.
      const data = response.data
      if (!data?.name || !data?.hash) {
        throw new Error('runner returned invalid snapshot-from-sandbox response')
      }

      return {
        ref: data.name,
        hash: data.hash,
        sizeGB: data.sizeGB,
        entrypoint: data.entrypoint,
        cmd: data.cmd,
      }
    }

    // 202 Accepted: the capture runs in the background on the runner; poll the
    // status endpoint until it reaches a terminal state.
    return this.pollSnapshotFromSandbox(sandboxId, snapshotName)
  }

  private async pollSnapshotFromSandbox(sandboxId: string, snapshotName: string): Promise<CreateSandboxSnapshotResult> {
    const pollTimeoutMin = this.configService.getOrThrow('sandboxSnapshottingTimeoutMin')
    const pollTimeoutMs = pollTimeoutMin * 60 * 1_000
    const pollIntervalMs = 5 * 1_000 // 5 seconds
    const deadline = Date.now() + pollTimeoutMs

    while (Date.now() < deadline) {
      await new Promise((resolve) => setTimeout(resolve, pollIntervalMs))

      // Clamp the per-request timeout to the remaining poll budget so a
      // request fired near the deadline cannot overshoot it (floored at 1s).
      const timeout = Math.max(1_000, Math.min(SNAPSHOT_STATUS_POLL_TIMEOUT_MS, deadline - Date.now()))

      let status: SnapshotFromSandboxStatusResponse
      try {
        const response = await this.sandboxApiClient.snapshotFromSandboxStatus(sandboxId, { timeout })
        status = response.data
      } catch (err) {
        const reason = err instanceof Error ? err.message : String(err)
        const statusCode =
          err instanceof RunnerApiError ? err.statusCode : axios.isAxiosError(err) ? err.response?.status : undefined

        if (statusCode !== undefined && statusCode >= 400 && statusCode < 500) {
          // Permanent client errors will not heal with more polling. A 404 may
          // also be a route miss: the runner restarted into an older version
          // that lacks the status endpoint, indistinguishable from a missing
          // capture record here.
          if (statusCode === 404) {
            throw new Error(
              `runner no longer exposes the snapshot capture status (runner restarted or downgraded?): ${reason}`,
            )
          }
          throw new Error(`snapshot capture status poll failed with HTTP ${statusCode}: ${reason}`)
        }

        // Transient poll failures (network blips, per-request timeouts, runner
        // 5xx) are tolerated; the poll budget bounds them.
        this.logger.warn(`Failed to poll snapshot capture status for sandbox ${sandboxId}: ${reason}`)
        continue
      }

      if (status?.name && status.name !== snapshotName) {
        throw new Error('snapshot capture state on runner does not match this capture (superseded or stale)')
      }

      switch (status?.state) {
        case EnumsSnapshotFromSandboxState.SnapshotFromSandboxStateCompleted: {
          const snapshot = status.snapshot
          if (!snapshot?.name || !snapshot?.hash) {
            throw new Error('runner returned invalid snapshot-from-sandbox capture result')
          }

          return {
            ref: snapshot.name,
            hash: snapshot.hash,
            sizeGB: snapshot.sizeGB,
            entrypoint: snapshot.entrypoint,
            cmd: snapshot.cmd,
          }
        }
        case EnumsSnapshotFromSandboxState.SnapshotFromSandboxStateFailed:
          throw new Error(status.error || 'snapshot capture failed on runner')
        case EnumsSnapshotFromSandboxState.SnapshotFromSandboxStateNone:
          // The runner writes IN_PROGRESS before answering 202, so NONE means
          // the in-memory capture record is gone (runner restarted).
          throw new Error('runner is no longer tracking the snapshot capture (runner restarted?)')
        default:
          continue
      }
    }

    throw new Error(`Timed out waiting for snapshot capture after ${pollTimeoutMin} minutes`)
  }

  // skipStart is a v2-only signal (carried in the job payload); v0's sync API has no equivalent.
  async recoverSandbox(sandbox: Sandbox, registry?: DockerRegistry, _skipStart?: boolean): Promise<void> {
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
      registry: registry
        ? {
            project: registry.project,
            url: registry.url.replace(/^(https?:\/\/)/, ''),
            username: registry.username,
            password: registry.password,
          }
        : undefined,
    }
    await this.sandboxApiClient.recover(sandbox.id, recoverSandboxDTO)
  }

  async resizeSandbox(
    sandboxId: string,
    cpu?: number,
    memory?: number,
    disk?: number,
    registry?: DockerRegistry,
  ): Promise<void> {
    await this.sandboxApiClient.resize(sandboxId, {
      cpu,
      memory,
      disk,
      registry: registry
        ? {
            project: registry.project,
            url: registry.url.replace(/^(https?:\/\/)/, ''),
            username: registry.username,
            password: registry.password,
          }
        : undefined,
    })
  }
}
