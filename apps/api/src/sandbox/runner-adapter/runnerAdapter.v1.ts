/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { RunnerAdapter, RunnerSandboxInfo } from './runnerAdapter'
import { Runner } from '../entities/runner.entity'
import { Sandbox } from '../entities/sandbox.entity'
import { BuildInfo } from '../entities/build-info.entity'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { BackupState } from '../enums/backup-state.enum'
import {
  SandboxState as RunnerSandboxState,
  BackupState as RunnerBackupState,
  RUNNER_V1_PACKAGE_NAME,
  SandboxServiceClient,
  SnapshotServiceClient,
  HealthServiceClient,
  SANDBOX_SERVICE_NAME,
  SNAPSHOT_SERVICE_NAME,
  HEALTH_SERVICE_NAME,
  HealthStatus,
} from '@daytonaio/runner-grpc-client'
import { ClientGrpc, ClientProxyFactory, Transport } from '@nestjs/microservices'
import { credentials, Metadata } from '@grpc/grpc-js'
import { join } from 'node:path'
import { firstValueFrom } from 'rxjs'

@Injectable()
export class RunnerAdapterV1 implements RunnerAdapter {
  private readonly logger = new Logger(RunnerAdapterV1.name)
  private runnerClient: ClientGrpc
  private sandboxServiceClient: SandboxServiceClient
  private snapshotServiceClient: SnapshotServiceClient
  private healthServiceClient: HealthServiceClient
  private runner: Runner

  private convertSandboxState(state: RunnerSandboxState): SandboxState {
    switch (state) {
      case RunnerSandboxState.SANDBOX_STATE_CREATING:
        return SandboxState.CREATING
      case RunnerSandboxState.SANDBOX_STATE_RESTORING:
        return SandboxState.RESTORING
      case RunnerSandboxState.SANDBOX_STATE_DESTROYED:
        return SandboxState.DESTROYED
      case RunnerSandboxState.SANDBOX_STATE_DESTROYING:
        return SandboxState.DESTROYING
      case RunnerSandboxState.SANDBOX_STATE_STARTED:
        return SandboxState.STARTED
      case RunnerSandboxState.SANDBOX_STATE_STOPPED:
        return SandboxState.STOPPED
      case RunnerSandboxState.SANDBOX_STATE_STARTING:
        return SandboxState.STARTING
      case RunnerSandboxState.SANDBOX_STATE_STOPPING:
        return SandboxState.STOPPING
      case RunnerSandboxState.SANDBOX_STATE_ERROR:
        return SandboxState.ERROR
      case RunnerSandboxState.SANDBOX_STATE_PULLING_SNAPSHOT:
        return SandboxState.PULLING_SNAPSHOT
      default:
        return SandboxState.UNKNOWN
    }
  }

  private convertBackupState(state: RunnerBackupState): BackupState {
    switch (state) {
      case RunnerBackupState.BACKUP_STATE_PENDING:
        return BackupState.PENDING
      case RunnerBackupState.BACKUP_STATE_IN_PROGRESS:
        return BackupState.IN_PROGRESS
      case RunnerBackupState.BACKUP_STATE_COMPLETED:
        return BackupState.COMPLETED
      case RunnerBackupState.BACKUP_STATE_FAILED:
        return BackupState.ERROR
      default:
        return BackupState.NONE
    }
  }

  private getMetadata(): Metadata {
    const md = new Metadata()
    md.add('authorization', `Bearer ${this.runner.apiKey}`)
    return md
  }

  public async init(runner: Runner): Promise<void> {
    this.runner = runner
    this.runnerClient = ClientProxyFactory.create({
      transport: Transport.GRPC,
      options: {
        credentials: credentials.createInsecure(),
        url: runner.apiUrl,
        package: RUNNER_V1_PACKAGE_NAME,
        protoPath: [join(__dirname, 'proto/runner/v1/runner.proto')],
        loader: {
          includeDirs: [join(__dirname, 'proto')],
        },
      },
    })
    this.sandboxServiceClient = this.runnerClient.getService<SandboxServiceClient>(SANDBOX_SERVICE_NAME)
    this.snapshotServiceClient = this.runnerClient.getService<SnapshotServiceClient>(SNAPSHOT_SERVICE_NAME)
    this.healthServiceClient = this.runnerClient.getService<HealthServiceClient>(HEALTH_SERVICE_NAME)
  }

  async healthCheck(): Promise<void> {
    const response = await firstValueFrom(this.healthServiceClient.healthCheck({}, this.getMetadata()))
    if (response.status !== HealthStatus.HEALTH_STATUS_HEALTHY) {
      throw new Error('Runner is not healthy')
    }
  }

  async buildSnapshot(buildInfo: BuildInfo, organizationId?: string, registry?: DockerRegistry): Promise<void> {
    await firstValueFrom(
      this.snapshotServiceClient.buildSnapshot(
        {
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
        },
        this.getMetadata(),
      ),
    )
  }

  async create(sandbox: Sandbox, registry: DockerRegistry, entrypoint?: string[]): Promise<void> {
    await firstValueFrom(
      this.sandboxServiceClient.createSandbox(
        {
          id: sandbox.id,
          snapshot: sandbox.snapshot,
          osUser: sandbox.osUser,
          userId: sandbox.organizationId,
          storageQuota: sandbox.disk,
          memoryQuota: sandbox.mem,
          cpuQuota: sandbox.cpu,
          gpuQuota: sandbox.gpu,
          env: sandbox.env,
          registry: {
            project: registry.name,
            url: registry.url,
            username: registry.username,
            password: registry.password,
          },
          entrypoint: entrypoint || [],
          volumes: sandbox.volumes,
        },
        this.getMetadata(),
      ),
    )
  }

  async createBackup(sandbox: Sandbox, backupSnapshotName: string, registry: DockerRegistry): Promise<void> {
    await firstValueFrom(
      this.sandboxServiceClient.createBackup(
        {
          sandboxId: sandbox.id,
          registry: {
            project: registry.name,
            url: registry.url,
            username: registry.username,
            password: registry.password,
          },
          snapshot: backupSnapshotName,
        },
        this.getMetadata(),
      ),
    )
  }

  async info(sandboxId: string): Promise<RunnerSandboxInfo> {
    const sandboxInfo = await firstValueFrom(
      this.sandboxServiceClient.getSandboxInfo(
        {
          sandboxId,
        },
        this.getMetadata(),
      ),
    )
    return {
      state: this.convertSandboxState(sandboxInfo.state),
      backupState: this.convertBackupState(sandboxInfo.backupState),
    }
  }

  async start(sandboxId: string): Promise<void> {
    await firstValueFrom(
      this.sandboxServiceClient.startSandbox(
        {
          sandboxId,
        },
        this.getMetadata(),
      ),
    )
  }

  async stop(sandboxId: string): Promise<void> {
    await firstValueFrom(
      this.sandboxServiceClient.stopSandbox(
        {
          sandboxId,
        },
        this.getMetadata(),
      ),
    )
  }

  async destroy(sandboxId: string): Promise<void> {
    await firstValueFrom(
      this.sandboxServiceClient.destroySandbox(
        {
          sandboxId,
        },
        this.getMetadata(),
      ),
    )
  }

  async removeDestroyed(sandboxId: string): Promise<void> {
    await firstValueFrom(
      this.sandboxServiceClient.removeDestroyedSandbox(
        {
          sandboxId,
        },
        this.getMetadata(),
      ),
    )
  }

  async removeSnapshot(snapshotName: string, force: boolean): Promise<void> {
    await firstValueFrom(
      this.snapshotServiceClient.removeSnapshot(
        {
          snapshot: snapshotName,
          force,
        },
        this.getMetadata(),
      ),
    )
  }

  async getSnapshotLogs(snapshotRef: string, follow: boolean): Promise<string> {
    const logs = await firstValueFrom(
      this.snapshotServiceClient.getSnapshotLogs(
        {
          snapshotRef,
          follow,
        },
        this.getMetadata(),
      ),
    )
    return 'TODO'
  }

  async snapshotExists(snapshotName: string): Promise<boolean> {
    const snapshot = await firstValueFrom(
      this.snapshotServiceClient.snapshotExists(
        {
          snapshot: snapshotName,
          includeLatest: true,
        },
        this.getMetadata(),
      ),
    )
    return snapshot.exists
  }

  async pullSnapshot(snapshotName: string, registry: DockerRegistry): Promise<void> {
    await firstValueFrom(
      this.snapshotServiceClient.pullSnapshot(
        {
          snapshot: snapshotName,
          registry: {
            project: registry.name,
            url: registry.url,
            username: registry.username,
            password: registry.password,
          },
        },
        this.getMetadata(),
      ),
    )
  }
}
