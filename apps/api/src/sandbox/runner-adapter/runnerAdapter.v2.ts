/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { RunnerAdapter, RunnerInfo, RunnerSandboxInfo } from './runnerAdapter'
import { Runner } from '../entities/runner.entity'
import { Sandbox } from '../entities/sandbox.entity'
import { BuildInfo } from '../entities/build-info.entity'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { BackupState } from '../enums/backup-state.enum'
import { v2 } from '@daytonaio/runner-grpc-client'
import { ClientGrpc, ClientProxyFactory, Transport } from '@nestjs/microservices'
import { ChannelCredentials, credentials, Metadata, status, ServiceError } from '@grpc/grpc-js'
import { join } from 'node:path'
import { catchError, firstValueFrom, of } from 'rxjs'
import { inspect } from 'node:util'

@Injectable()
export class RunnerAdapterV2 implements RunnerAdapter {
  private readonly logger = new Logger(RunnerAdapterV2.name)
  private runnerClient: ClientGrpc
  private sandboxServiceClient: v2.SandboxServiceClient
  private snapshotServiceClient: v2.SnapshotServiceClient
  private statusServiceClient: v2.StatusServiceClient
  private runner: Runner

  private convertSandboxState(state: v2.SandboxState): SandboxState {
    switch (state) {
      case v2.SandboxState.SANDBOX_STATE_CREATING:
        return SandboxState.CREATING
      case v2.SandboxState.SANDBOX_STATE_RESTORING:
        return SandboxState.RESTORING
      case v2.SandboxState.SANDBOX_STATE_DESTROYED:
        return SandboxState.DESTROYED
      case v2.SandboxState.SANDBOX_STATE_DESTROYING:
        return SandboxState.DESTROYING
      case v2.SandboxState.SANDBOX_STATE_STARTED:
        return SandboxState.STARTED
      case v2.SandboxState.SANDBOX_STATE_STOPPED:
        return SandboxState.STOPPED
      case v2.SandboxState.SANDBOX_STATE_STARTING:
        return SandboxState.STARTING
      case v2.SandboxState.SANDBOX_STATE_STOPPING:
        return SandboxState.STOPPING
      case v2.SandboxState.SANDBOX_STATE_ERROR:
        return SandboxState.ERROR
      default:
        return SandboxState.UNKNOWN
    }
  }

  private getMetadata(): Metadata {
    const md = new Metadata()
    md.add('authorization', `Bearer ${this.runner.apiKey}`)
    return md
  }

  public async init(runner: Runner): Promise<void> {
    // Get grpc security scheme from runner url
    // api url must be in format: scheme://url where scheme is either grpc or grpcs
    // if scheme is grpcs, we need to use credentials.createSsl()
    // if scheme is grpc, we need to use credentials.createInsecure()
    const [scheme, url] = runner.apiUrl.split('://')
    let creds: ChannelCredentials
    switch (scheme) {
      case 'grpc':
        creds = credentials.createInsecure()
        break
      case 'grpcs':
        creds = credentials.createSsl()
        break
      default:
        throw new Error(`Invalid runner apiUrl: ${runner.apiUrl}`)
    }

    this.runner = runner
    this.runnerClient = ClientProxyFactory.create({
      transport: Transport.GRPC,
      options: {
        credentials: creds,
        url: url,
        package: v2.RUNNER_V2ALPHA_PACKAGE_NAME,
        protoPath: [join(__dirname, 'proto/runner/v2/runner.proto')],
        loader: {
          includeDirs: [join(__dirname, 'proto')],
          longs: Number, // Convert Long values to JavaScript numbers
        },
      },
    })
    this.sandboxServiceClient = this.runnerClient.getService<v2.SandboxServiceClient>(v2.SANDBOX_SERVICE_NAME)
    this.snapshotServiceClient = this.runnerClient.getService<v2.SnapshotServiceClient>(v2.SNAPSHOT_SERVICE_NAME)
    this.statusServiceClient = this.runnerClient.getService<v2.StatusServiceClient>(v2.STATUS_SERVICE_NAME)
  }

  async healthCheck(): Promise<void> {
    const response = await firstValueFrom(this.statusServiceClient.health({}, this.getMetadata()))
    if (response.status !== v2.HealthStatus.HEALTH_STATUS_HEALTHY) {
      throw new Error('Runner is not healthy')
    }
  }

  async runnerInfo(): Promise<RunnerInfo> {
    const response = await firstValueFrom(this.statusServiceClient.stats({}, this.getMetadata()))

    return {
      metrics: {
        currentCpuUsagePercentage: response.metrics.currentCpuUsagePercentage || 0,
        currentMemoryUsagePercentage: response.metrics.currentMemoryUsagePercentage || 0,
        currentDiskUsagePercentage: response.metrics.currentDiskUsagePercentage || 0,
        currentAllocatedCpu: response.metrics.currentAllocatedCpu || 0,
        currentAllocatedMemoryGiB: response.metrics.currentAllocatedMemoryGiB || 0,
        currentAllocatedDiskGiB: response.metrics.currentAllocatedDiskGiB || 0,
        currentSnapshotCount: response.metrics.currentSnapshotCount || 0,
      },
    }
  }

  async sandboxInfo(sandboxId: string): Promise<RunnerSandboxInfo> {
    const sandboxInfo = await firstValueFrom(
      this.sandboxServiceClient.getSandbox(
        {
          id: sandboxId,
        },
        this.getMetadata(),
      ),
    )
    return {
      state: this.convertSandboxState(sandboxInfo.sandbox.state),
      backupState: BackupState.NONE,
    }
  }

  async createSandbox(sandbox: Sandbox): Promise<void> {
    const request: v2.CreateSandboxRequest = {
      id: sandbox.id,
      snapshotId: sandbox.snapshot,
      volumes: sandbox.volumes,
    }

    await firstValueFrom(this.sandboxServiceClient.createSandbox(request, this.getMetadata()))
  }

  async startSandbox(sandboxId: string): Promise<void> {
    await firstValueFrom(
      this.sandboxServiceClient.startSandbox(
        {
          id: sandboxId,
        },
        this.getMetadata(),
      ),
    )
  }

  async stopSandbox(sandboxId: string): Promise<void> {
    await firstValueFrom(
      this.sandboxServiceClient.stopSandbox(
        {
          id: sandboxId,
        },
        this.getMetadata(),
      ),
    )
  }

  async destroySandbox(sandboxId: string): Promise<void> {
    await firstValueFrom(
      this.sandboxServiceClient
        .destroySandbox(
          {
            id: sandboxId,
          },
          this.getMetadata(),
        )
        .pipe(
          catchError((error: ServiceError) => {
            if (error.code === status.NOT_FOUND) {
              return of()
            }
            throw error
          }),
        ),
    )
  }

  async removeDestroyedSandbox(sandboxId: string): Promise<void> {
    void Promise.resolve()
  }

  async createBackup(sandbox: Sandbox, backupSnapshotName: string, registry?: DockerRegistry): Promise<void> {
    throw new Error('Create backup is not supported in v2 runner')
  }

  async removeSnapshot(snapshotName: string): Promise<void> {
    await firstValueFrom(
      this.snapshotServiceClient.removeSnapshot(
        {
          id: snapshotName,
        },
        this.getMetadata(),
      ),
    )
  }

  async buildSnapshot(
    buildInfo: BuildInfo,
    organizationId?: string,
    registry?: DockerRegistry,
    pushToInternalRegistry?: boolean,
  ): Promise<void> {
    throw new Error('Build snapshot is not supported in v2 runner')
  }

  async pullSnapshot(snapshotName: string, registry?: DockerRegistry): Promise<void> {
    const request: v2.PullSnapshotRequest = {
      id: snapshotName,
      registry,
    }

    if (registry) {
      request.registry = {
        project: registry.name,
        url: registry.url,
        username: registry.username,
        password: registry.password,
      }
    }

    await firstValueFrom(this.snapshotServiceClient.pullSnapshot(request, this.getMetadata()))
  }

  async snapshotExists(snapshotName: string): Promise<boolean> {
    try {
      await firstValueFrom(
        this.snapshotServiceClient.getSnapshot(
          {
            id: snapshotName,
          },
          this.getMetadata(),
        ),
      )
      return true
    } catch (error) {
      return false
    }
  }

  async getSnapshotLogs(snapshotRef: string, follow: boolean): Promise<string> {
    const logs = await firstValueFrom(
      this.snapshotServiceClient.getSnapshotLogs(
        {
          id: snapshotRef,
        },
        this.getMetadata(),
      ),
    )
    return logs.content
  }

  async getSandboxDaemonVersion(sandboxId: string): Promise<string> {
    return 'unknown'
  }

  async updateNetworkSettings(sandboxId: string, networkBlockAll?: boolean, networkAllowList?: string): Promise<void> {
    throw new Error('Update network settings is not supported in v2 runner')
  }

  async snapshotSandbox(sandboxId: string): Promise<string> {
    const snapshot = await firstValueFrom(
      this.sandboxServiceClient.snapshotSandbox(
        {
          id: sandboxId,
          snapshotType: v2.SnapshotType.SNAPSHOT_TYPE_FULL,
        },
        this.getMetadata(),
      ),
    )
    return snapshot.snapshot.id
  }
}
