/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, IsNull, Not } from 'typeorm'
import { create, toJson } from '@bufbuild/protobuf'
import {
  CreateSandboxPayloadSchema,
  CreateSandboxResultSchema,
  StartSandboxPayloadSchema,
  StartSandboxResultSchema,
  StopSandboxPayloadSchema,
  ResizeSandboxPayloadSchema,
  UpdateNetworkSettingsPayloadSchema,
  PullSnapshotPayloadSchema,
  SnapshotInfoResponseSchema,
  InspectSnapshotInRegistryPayloadSchema,
  SnapshotDigestResponseSchema,
  ForkSandboxPayloadSchema,
  ForkSandboxResultSchema,
  SnapshotSandboxPayloadSchema,
  SnapshotSandboxResultSchema,
  RegistrySchema,
  VolumeSchema,
} from '@daytona/runner-proto'
import {
  RunnerAdapter,
  RunnerInfo,
  RunnerSandboxInfo,
  RunnerSnapshotInfo,
  StartSandboxResponse,
  SnapshotDigestResponse,
} from './runnerAdapter'
import { Runner } from '../entities/runner.entity'
import { Sandbox } from '../entities/sandbox.entity'
import { Job } from '../entities/job.entity'
import { BuildInfo } from '../entities/build-info.entity'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { JobType } from '../enums/job-type.enum'
import { JobStatus } from '../enums/job-status.enum'
import { ResourceType } from '../enums/resource-type.enum'
import { JobService } from '../services/job.service'
import { SandboxRepository } from '../repositories/sandbox.repository'
import { SnapshotStateError } from '../errors/snapshot-state-error'

@Injectable()
export class RunnerAdapterV3 implements RunnerAdapter {
  private readonly logger = new Logger(RunnerAdapterV3.name)
  private runner: Runner

  constructor(
    private readonly sandboxRepository: SandboxRepository,
    @InjectRepository(Job)
    private readonly jobRepository: Repository<Job>,
    private readonly jobService: JobService,
  ) {}

  async init(runner: Runner): Promise<void> {
    this.runner = runner
  }

  async healthCheck(_signal?: AbortSignal): Promise<void> {
    throw new Error('healthCheck is not supported for V3 runners')
  }

  async runnerInfo(_signal?: AbortSignal): Promise<RunnerInfo> {
    throw new Error('runnerInfo is not supported for V3 runners')
  }

  async sandboxInfo(sandboxId: string): Promise<RunnerSandboxInfo> {
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      throw new Error(`Sandbox ${sandboxId} not found`)
    }

    const incompleteJob = await this.jobRepository.findOne({
      where: {
        resourceType: ResourceType.SANDBOX,
        resourceId: sandboxId,
        completedAt: IsNull(),
      },
      order: { createdAt: 'DESC' },
    })

    let state = sandbox.state
    let daemonVersion: string | undefined = undefined

    if (incompleteJob) {
      state = this.inferStateFromJob(incompleteJob, sandbox)
      daemonVersion = incompleteJob.getResultMetadata()?.daemonVersion
    } else {
      const latestJob = await this.jobRepository.findOne({
        where: {
          resourceType: ResourceType.SANDBOX,
          resourceId: sandboxId,
        },
        order: { createdAt: 'DESC' },
      })
      if (latestJob) {
        state = this.inferStateFromJob(latestJob, sandbox)
        daemonVersion = latestJob.getResultMetadata()?.daemonVersion
      }
    }

    return {
      state,
      backupState: sandbox.backupState,
      backupErrorReason: sandbox.backupErrorReason,
      daemonVersion,
    }
  }

  private inferStateFromJob(job: Job, sandbox: Sandbox): SandboxState {
    switch (job.type) {
      case JobType.CREATE_SANDBOX:
        return job.status === JobStatus.COMPLETED ? SandboxState.STARTED : SandboxState.CREATING
      case JobType.START_SANDBOX:
        return job.status === JobStatus.COMPLETED ? SandboxState.STARTED : SandboxState.STARTING
      case JobType.STOP_SANDBOX:
        return job.status === JobStatus.COMPLETED ? SandboxState.STOPPED : SandboxState.STOPPING
      case JobType.DESTROY_SANDBOX:
        return job.status === JobStatus.COMPLETED ? SandboxState.DESTROYED : SandboxState.DESTROYING
      default:
        return sandbox.state
    }
  }

  private toRegistryProto(registry: DockerRegistry) {
    return create(RegistrySchema, {
      url: registry.url.replace(/^(https?:\/\/)/, ''),
      username: registry.username ?? undefined,
      password: registry.password ?? undefined,
      project: registry.project ?? undefined,
    })
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
    const payload = toJson(
      CreateSandboxPayloadSchema,
      create(CreateSandboxPayloadSchema, {
        id: sandbox.id,
        userId: sandbox.organizationId,
        snapshot: snapshotRef,
        osUser: sandbox.osUser,
        cpuQuota: BigInt(sandbox.cpu),
        gpuQuota: BigInt(sandbox.gpu),
        memoryQuota: BigInt(sandbox.mem),
        storageQuota: BigInt(sandbox.disk),
        env: sandbox.env ?? {},
        registry: registry ? this.toRegistryProto(registry) : undefined,
        entrypoint: entrypoint ?? [],
        volumes:
          sandbox.volumes?.map((volume) =>
            create(VolumeSchema, {
              volumeId: volume.volumeId,
              mountPath: volume.mountPath,
              subpath: volume.subpath,
            }),
          ) ?? [],
        networkBlockAll: sandbox.networkBlockAll,
        networkAllowList: sandbox.networkAllowList,
        metadata: metadata ?? {},
        authToken: sandbox.authToken,
        otelEndpoint: otelEndpoint,
        skipStart: skipStart,
        organizationId: sandbox.organizationId,
        regionId: sandbox.region,
      }),
    )

    await this.jobService.createJob(
      null,
      JobType.CREATE_SANDBOX,
      this.runner.id,
      ResourceType.SANDBOX,
      sandbox.id,
      payload as Record<string, any>,
      CreateSandboxPayloadSchema.typeName,
      CreateSandboxResultSchema.typeName,
    )

    this.logger.debug(`Created CREATE_SANDBOX job for sandbox ${sandbox.id} on runner ${this.runner.id}`)
    return undefined
  }

  async startSandbox(
    sandboxId: string,
    authToken: string,
    metadata?: { [key: string]: string },
  ): Promise<StartSandboxResponse | undefined> {
    const payload = toJson(
      StartSandboxPayloadSchema,
      create(StartSandboxPayloadSchema, {
        authToken,
        metadata: metadata ?? {},
      }),
    )

    await this.jobService.createJob(
      null,
      JobType.START_SANDBOX,
      this.runner.id,
      ResourceType.SANDBOX,
      sandboxId,
      payload as Record<string, any>,
      StartSandboxPayloadSchema.typeName,
      StartSandboxResultSchema.typeName,
    )

    this.logger.debug(`Created START_SANDBOX job for sandbox ${sandboxId} on runner ${this.runner.id}`)
    return undefined
  }

  async stopSandbox(sandboxId: string, force?: boolean): Promise<void> {
    const payload = toJson(StopSandboxPayloadSchema, create(StopSandboxPayloadSchema, { force }))

    await this.jobService.createJob(
      null,
      JobType.STOP_SANDBOX,
      this.runner.id,
      ResourceType.SANDBOX,
      sandboxId,
      payload as Record<string, any>,
      StopSandboxPayloadSchema.typeName,
    )

    this.logger.debug(`Created STOP_SANDBOX job for sandbox ${sandboxId} on runner ${this.runner.id}`)
  }

  async destroySandbox(sandboxId: string): Promise<void> {
    await this.jobService.createJob(null, JobType.DESTROY_SANDBOX, this.runner.id, ResourceType.SANDBOX, sandboxId)

    this.logger.debug(`Created DESTROY_SANDBOX job for sandbox ${sandboxId} on runner ${this.runner.id}`)
  }

  async recoverSandbox(_sandbox: Sandbox): Promise<void> {
    throw new Error('recoverSandbox is not supported for V3 runners')
  }

  async createBackup(_sandbox: Sandbox, _backupSnapshotName: string, _registry?: DockerRegistry): Promise<void> {
    throw new Error('createBackup is not supported for V3 runners')
  }

  async buildSnapshot(
    _buildInfo: BuildInfo,
    _organizationId?: string,
    _sourceRegistries?: DockerRegistry[],
    _registry?: DockerRegistry,
    _pushToInternalRegistry?: boolean,
  ): Promise<void> {
    throw new Error('buildSnapshot is not supported for V3 runners')
  }

  async pullSnapshot(
    snapshotName: string,
    registry?: DockerRegistry,
    destinationRegistry?: DockerRegistry,
    destinationRef?: string,
    newTag?: string,
  ): Promise<void> {
    const payload = toJson(
      PullSnapshotPayloadSchema,
      create(PullSnapshotPayloadSchema, {
        snapshot: snapshotName,
        registry: registry ? this.toRegistryProto(registry) : undefined,
        destinationRegistry: destinationRegistry ? this.toRegistryProto(destinationRegistry) : undefined,
        destinationRef,
        newTag,
      }),
    )

    await this.jobService.createJob(
      null,
      JobType.PULL_SNAPSHOT,
      this.runner.id,
      ResourceType.SNAPSHOT,
      destinationRef || snapshotName,
      payload as Record<string, any>,
      PullSnapshotPayloadSchema.typeName,
      SnapshotInfoResponseSchema.typeName,
    )

    this.logger.debug(`Created PULL_SNAPSHOT job for ${snapshotName} on runner ${this.runner.id}`)
  }

  async removeSnapshot(snapshotName: string): Promise<void> {
    await this.jobService.createJob(null, JobType.REMOVE_SNAPSHOT, this.runner.id, ResourceType.SNAPSHOT, snapshotName)

    this.logger.debug(`Created REMOVE_SNAPSHOT job for ${snapshotName} on runner ${this.runner.id}`)
  }

  async snapshotExists(snapshotRef: string): Promise<boolean> {
    const latestJob = await this.jobRepository.findOne({
      where: [
        {
          runnerId: this.runner.id,
          resourceType: ResourceType.SNAPSHOT,
          resourceId: snapshotRef,
          type: Not(JobType.INSPECT_SNAPSHOT_IN_REGISTRY),
        },
      ],
      order: { createdAt: 'DESC' },
    })

    if (!latestJob) {
      return false
    }

    if (latestJob.type === JobType.REMOVE_SNAPSHOT) {
      return false
    }

    if (latestJob.type === JobType.PULL_SNAPSHOT || latestJob.type === JobType.BUILD_SNAPSHOT) {
      return latestJob.status === JobStatus.COMPLETED
    }

    return false
  }

  async getSnapshotInfo(snapshotRef: string): Promise<RunnerSnapshotInfo> {
    const latestJob = await this.jobRepository.findOne({
      where: [
        {
          runnerId: this.runner.id,
          resourceType: ResourceType.SNAPSHOT,
          resourceId: snapshotRef,
          type: Not(JobType.INSPECT_SNAPSHOT_IN_REGISTRY),
        },
      ],
      order: { createdAt: 'DESC' },
    })

    if (!latestJob) {
      throw new Error(`Snapshot ${snapshotRef} not found on runner ${this.runner.id}`)
    }

    const metadata = latestJob.getResultMetadata()

    switch (latestJob.status) {
      case JobStatus.COMPLETED:
        if (latestJob.type === JobType.PULL_SNAPSHOT || latestJob.type === JobType.BUILD_SNAPSHOT) {
          const missingFields: string[] = []
          if (metadata?.sizeGB === undefined) {
            missingFields.push('sizeGB')
          }
          if (metadata?.entrypoint === undefined) {
            missingFields.push('entrypoint')
          }
          if (metadata?.cmd === undefined) {
            missingFields.push('cmd')
          }
          if (metadata?.hash === undefined) {
            missingFields.push('hash')
          }
          if (missingFields.length > 0) {
            throw new Error(
              `Snapshot ${snapshotRef} on runner ${this.runner.id} has incomplete result metadata: missing ${missingFields.join(', ')}`,
            )
          }
          return {
            name: latestJob.resourceId,
            sizeGB: metadata.sizeGB,
            entrypoint: metadata.entrypoint,
            cmd: metadata.cmd,
            hash: metadata.hash,
          }
        }
        throw new Error(
          `Snapshot ${snapshotRef} is in an unknown state (${latestJob.status}) on runner ${this.runner.id}`,
        )
      case JobStatus.FAILED:
        throw new SnapshotStateError(
          latestJob.errorMessage || `Snapshot ${snapshotRef} failed on runner ${this.runner.id}`,
        )
      default:
        throw new Error(
          `Snapshot ${snapshotRef} is in an unknown state (${latestJob.status}) on runner ${this.runner.id}`,
        )
    }
  }

  async inspectSnapshotInRegistry(snapshotName: string, registry?: DockerRegistry): Promise<SnapshotDigestResponse> {
    const payload = toJson(
      InspectSnapshotInRegistryPayloadSchema,
      create(InspectSnapshotInRegistryPayloadSchema, {
        snapshot: snapshotName,
        registry: registry ? this.toRegistryProto(registry) : undefined,
      }),
    )

    const job = await this.jobService.createJob(
      null,
      JobType.INSPECT_SNAPSHOT_IN_REGISTRY,
      this.runner.id,
      ResourceType.SNAPSHOT,
      snapshotName,
      payload as Record<string, any>,
      InspectSnapshotInRegistryPayloadSchema.typeName,
      SnapshotDigestResponseSchema.typeName,
    )

    this.logger.debug(`Created INSPECT_SNAPSHOT_IN_REGISTRY job for ${snapshotName} on runner ${this.runner.id}`)

    const waitTimeout = 30 * 1000
    const completedJob = await this.jobService.waitJobCompletion(job.id, waitTimeout)

    if (!completedJob) {
      throw new Error(`Snapshot ${snapshotName} not found in registry on runner ${this.runner.id}`)
    }

    if (completedJob.status !== JobStatus.COMPLETED) {
      throw new Error(
        `Snapshot ${snapshotName} failed to inspect in registry on runner ${this.runner.id}. Error: ${completedJob.errorMessage}`,
      )
    }

    const resultMetadata = completedJob.getResultMetadata()

    if (typeof resultMetadata?.hash !== 'string' || typeof resultMetadata?.sizeGB !== 'number') {
      throw new Error(`Snapshot ${snapshotName} inspection on runner ${this.runner.id} returned invalid metadata`)
    }

    return {
      hash: resultMetadata.hash,
      sizeGB: resultMetadata.sizeGB,
    }
  }

  async updateNetworkSettings(
    sandboxId: string,
    networkBlockAll?: boolean,
    networkAllowList?: string,
    networkLimitEgress?: boolean,
  ): Promise<void> {
    const payload = toJson(
      UpdateNetworkSettingsPayloadSchema,
      create(UpdateNetworkSettingsPayloadSchema, {
        networkBlockAll,
        networkAllowList,
        networkLimitEgress,
      }),
    )

    await this.jobService.createJob(
      null,
      JobType.UPDATE_SANDBOX_NETWORK_SETTINGS,
      this.runner.id,
      ResourceType.SANDBOX,
      sandboxId,
      payload as Record<string, any>,
      UpdateNetworkSettingsPayloadSchema.typeName,
    )

    this.logger.debug(
      `Created UPDATE_SANDBOX_NETWORK_SETTINGS job for sandbox ${sandboxId} on runner ${this.runner.id}`,
    )
  }

  async forkSandbox(sourceSandboxId: string, newSandboxId: string): Promise<void> {
    const payload = toJson(
      ForkSandboxPayloadSchema,
      create(ForkSandboxPayloadSchema, {
        sourceSandboxId,
        newSandboxId,
      }),
    )

    await this.jobService.createJob(
      null,
      JobType.FORK_SANDBOX,
      this.runner.id,
      ResourceType.SANDBOX,
      newSandboxId,
      payload as Record<string, any>,
      ForkSandboxPayloadSchema.typeName,
      ForkSandboxResultSchema.typeName,
    )

    this.logger.debug(
      `Created FORK_SANDBOX job for sandbox ${sourceSandboxId} -> ${newSandboxId} on runner ${this.runner.id}`,
    )
  }

  async createSnapshotFromSandbox(
    sandboxId: string,
    snapshotName: string,
    organizationId: string,
    registry?: DockerRegistry,
  ): Promise<void> {
    const payload = toJson(
      SnapshotSandboxPayloadSchema,
      create(SnapshotSandboxPayloadSchema, {
        sandboxId,
        name: snapshotName,
        organizationId,
        registry: registry ? this.toRegistryProto(registry) : undefined,
      }),
    )

    await this.jobService.createJob(
      null,
      JobType.SNAPSHOT_SANDBOX,
      this.runner.id,
      ResourceType.SANDBOX,
      sandboxId,
      payload as Record<string, any>,
      SnapshotSandboxPayloadSchema.typeName,
      SnapshotSandboxResultSchema.typeName,
    )

    this.logger.debug(`Created SNAPSHOT_SANDBOX job for sandbox ${sandboxId} on runner ${this.runner.id}`)
  }

  async resizeSandbox(sandboxId: string, cpu?: number, memory?: number, disk?: number): Promise<void> {
    const payload = toJson(
      ResizeSandboxPayloadSchema,
      create(ResizeSandboxPayloadSchema, {
        cpu: BigInt(cpu ?? 0),
        gpu: BigInt(0),
        memory: BigInt(memory ?? 0),
        disk: BigInt(disk ?? 0),
      }),
    )

    await this.jobService.createJob(
      null,
      JobType.RESIZE_SANDBOX,
      this.runner.id,
      ResourceType.SANDBOX,
      sandboxId,
      payload as Record<string, any>,
      ResizeSandboxPayloadSchema.typeName,
    )

    this.logger.debug(`Created RESIZE_SANDBOX job for sandbox ${sandboxId} on runner ${this.runner.id}`)
  }
}
