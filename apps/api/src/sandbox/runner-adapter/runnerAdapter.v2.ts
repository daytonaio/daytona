/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository, IsNull } from 'typeorm'
import { RunnerAdapter, RunnerInfo, RunnerSandboxInfo, RunnerSnapshotInfo, StartSandboxResponse } from './runnerAdapter'
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
import {
  CreateSandboxDTO,
  CreateBackupDTO,
  BuildSnapshotRequestDTO,
  PullSnapshotRequestDTO,
  UpdateNetworkSettingsDTO,
} from '@daytonaio/runner-api-client'

/**
 * RunnerAdapterV2 implements RunnerAdapter for v2 runners.
 * Instead of making direct API calls to the runner, it creates jobs in the database
 * that the v2 runner polls and processes asynchronously.
 */
@Injectable()
export class RunnerAdapterV2 implements RunnerAdapter {
  private readonly logger = new Logger(RunnerAdapterV2.name)
  private runner: Runner

  constructor(
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    @InjectRepository(Job)
    private readonly jobRepository: Repository<Job>,
    private readonly jobService: JobService,
  ) {}

  async init(runner: Runner): Promise<void> {
    this.runner = runner
  }

  async healthCheck(_signal?: AbortSignal): Promise<void> {
    throw new Error('healthCheck is not supported for V2 runners')
  }

  async runnerInfo(_signal?: AbortSignal): Promise<RunnerInfo> {
    throw new Error('runnerInfo is not supported for V2 runners')
  }

  async sandboxInfo(sandboxId: string): Promise<RunnerSandboxInfo> {
    // Query the sandbox entity
    const sandbox = await this.sandboxRepository.findOne({
      where: { id: sandboxId },
    })

    if (!sandbox) {
      throw new Error(`Sandbox ${sandboxId} not found`)
    }

    // Query for any incomplete jobs for this sandbox to determine transitional state
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

    // If there's an incomplete job, infer the transitional state from job type
    if (incompleteJob) {
      state = this.inferStateFromJob(incompleteJob, sandbox)
      daemonVersion = incompleteJob.getResultMetadata()?.daemonVersion
    } else {
      // Look for latest job for this sandbox
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
    // Map job types to transitional states
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
        // For other job types (backup, etc.), return current sandbox state
        return sandbox.state
    }
  }

  async createSandbox(
    sandbox: Sandbox,
    registry?: DockerRegistry,
    entrypoint?: string[],
    metadata?: { [key: string]: string },
  ): Promise<StartSandboxResponse | undefined> {
    const payload: CreateSandboxDTO = {
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
    }

    await this.jobService.createJob(
      null,
      JobType.CREATE_SANDBOX,
      this.runner.id,
      ResourceType.SANDBOX,
      sandbox.id,
      payload,
    )

    this.logger.debug(`Created CREATE_SANDBOX job for sandbox ${sandbox.id} on runner ${this.runner.id}`)

    // Daemon version will be set in the job result metadata
    return undefined
  }

  async startSandbox(
    sandboxId: string,
    metadata?: { [key: string]: string },
  ): Promise<StartSandboxResponse | undefined> {
    await this.jobService.createJob(
      null,
      JobType.START_SANDBOX,
      this.runner.id,
      ResourceType.SANDBOX,
      sandboxId,
      metadata,
    )

    this.logger.debug(`Created START_SANDBOX job for sandbox ${sandboxId} on runner ${this.runner.id}`)

    // Daemon version will be set in the job result metadata
    return undefined
  }

  async stopSandbox(sandboxId: string): Promise<void> {
    await this.jobService.createJob(null, JobType.STOP_SANDBOX, this.runner.id, ResourceType.SANDBOX, sandboxId)

    this.logger.debug(`Created STOP_SANDBOX job for sandbox ${sandboxId} on runner ${this.runner.id}`)
  }

  async destroySandbox(sandboxId: string): Promise<void> {
    await this.jobService.createJob(null, JobType.DESTROY_SANDBOX, this.runner.id, ResourceType.SANDBOX, sandboxId)

    this.logger.debug(`Created DESTROY_SANDBOX job for sandbox ${sandboxId} on runner ${this.runner.id}`)
  }

  async removeDestroyedSandbox(_sandboxId: string): Promise<void> {
    throw new Error('removeDestroyedSandbox is not supported for V2 runners')
  }

  async createBackup(sandbox: Sandbox, backupSnapshotName: string, registry?: DockerRegistry): Promise<void> {
    const payload: CreateBackupDTO = {
      snapshot: backupSnapshotName,
      registry: undefined,
    }

    if (registry) {
      payload.registry = {
        project: registry.project,
        url: registry.url.replace(/^(https?:\/\/)/, ''),
        username: registry.username,
        password: registry.password,
      }
    }

    await this.jobService.createJob(
      null,
      JobType.CREATE_BACKUP,
      this.runner.id,
      ResourceType.SANDBOX,
      sandbox.id,
      payload,
    )

    this.logger.debug(`Created CREATE_BACKUP job for sandbox ${sandbox.id} on runner ${this.runner.id}`)
  }

  async buildSnapshot(
    buildInfo: BuildInfo,
    organizationId?: string,
    sourceRegistries?: DockerRegistry[],
    registry?: DockerRegistry,
    pushToInternalRegistry?: boolean,
  ): Promise<void> {
    const payload: BuildSnapshotRequestDTO = {
      snapshot: buildInfo.snapshotRef,
      dockerfile: buildInfo.dockerfileContent,
      organizationId: organizationId,
      context: buildInfo.contextHashes,
      pushToInternalRegistry: pushToInternalRegistry,
    }

    if (sourceRegistries) {
      payload.sourceRegistries = sourceRegistries.map((sourceRegistry) => ({
        project: sourceRegistry.project,
        url: sourceRegistry.url.replace(/^(https?:\/\/)/, ''),
        username: sourceRegistry.username,
        password: sourceRegistry.password,
      }))
    }

    if (registry) {
      payload.registry = {
        project: registry.project,
        url: registry.url.replace(/^(https?:\/\/)/, ''),
        username: registry.username,
        password: registry.password,
      }
    }

    await this.jobService.createJob(
      null,
      JobType.BUILD_SNAPSHOT,
      this.runner.id,
      ResourceType.SNAPSHOT,
      buildInfo.snapshotRef,
      payload,
    )

    this.logger.debug(`Created BUILD_SNAPSHOT job for ${buildInfo.snapshotRef} on runner ${this.runner.id}`)
  }

  async pullSnapshot(
    snapshotName: string,
    registry?: DockerRegistry,
    destinationRegistry?: DockerRegistry,
    destinationRef?: string,
  ): Promise<void> {
    const payload: PullSnapshotRequestDTO = {
      snapshot: snapshotName,
    }

    if (registry) {
      payload.registry = {
        project: registry.project,
        url: registry.url.replace(/^(https?:\/\/)/, ''),
        username: registry.username,
        password: registry.password,
      }
    }

    if (destinationRegistry) {
      payload.destinationRegistry = {
        project: destinationRegistry.project,
        url: destinationRegistry.url.replace(/^(https?:\/\/)/, ''),
        username: destinationRegistry.username,
        password: destinationRegistry.password,
      }
    }

    if (destinationRef) {
      payload.destinationRef = destinationRef
    }

    await this.jobService.createJob(
      null,
      JobType.PULL_SNAPSHOT,
      this.runner.id,
      ResourceType.SNAPSHOT,
      destinationRef || snapshotName,
      payload,
    )

    this.logger.debug(`Created PULL_SNAPSHOT job for ${snapshotName} on runner ${this.runner.id}`)
  }

  async removeSnapshot(snapshotName: string): Promise<void> {
    await this.jobService.createJob(null, JobType.REMOVE_SNAPSHOT, this.runner.id, ResourceType.SNAPSHOT, snapshotName)

    this.logger.debug(`Created REMOVE_SNAPSHOT job for ${snapshotName} on runner ${this.runner.id}`)
  }

  async tagImage(_sourceImage: string, _targetImage: string): Promise<void> {
    throw new Error('tagImage is not supported for V2 runners')
  }

  async snapshotExists(snapshotRef: string): Promise<boolean> {
    // Find the latest job for this snapshot on this runner
    // We need to check both ResourceType.SNAPSHOT (for PULL/REMOVE) and ResourceType.SANDBOX (for BUILD)
    const latestJob = await this.jobRepository.findOne({
      where: [
        {
          runnerId: this.runner.id,
          resourceType: ResourceType.SNAPSHOT,
          resourceId: snapshotRef,
        },
      ],
      order: { createdAt: 'DESC' },
    })

    // If no job exists, snapshot doesn't exist
    if (!latestJob) {
      return false
    }

    // If the latest job is a REMOVE_SNAPSHOT, the snapshot no longer exists
    if (latestJob.type === JobType.REMOVE_SNAPSHOT) {
      return false
    }

    // If the latest job is PULL_SNAPSHOT or BUILD_SNAPSHOT, check if it completed successfully
    if (latestJob.type === JobType.PULL_SNAPSHOT || latestJob.type === JobType.BUILD_SNAPSHOT) {
      return latestJob.status === JobStatus.COMPLETED
    }

    // For any other job type, snapshot doesn't exist
    return false
  }

  async getSnapshotInfo(snapshotRef: string): Promise<RunnerSnapshotInfo> {
    const latestJob = await this.jobRepository.findOne({
      where: [
        {
          runnerId: this.runner.id,
          resourceType: ResourceType.SNAPSHOT,
          resourceId: snapshotRef,
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
          return {
            name: latestJob.resourceId,
            sizeGB: metadata?.sizeGB,
            entrypoint: metadata?.entrypoint,
            cmd: metadata?.cmd,
            hash: metadata?.hash,
          }
        }
        throw new Error(
          `Snapshot ${snapshotRef} is in an unknown state (${latestJob.status}) on runner ${this.runner.id}`,
        )
      case JobStatus.FAILED:
        throw new Error(`Snapshot ${snapshotRef} failed to build on runner ${this.runner.id}`)
      default:
        throw new Error(
          `Snapshot ${snapshotRef} is in an unknown state (${latestJob.status}) on runner ${this.runner.id}`,
        )
    }
  }

  async getSandboxDaemonVersion(_sandboxId: string): Promise<string> {
    throw new Error('getSandboxDaemonVersion is not supported for V2 runners')
  }

  async updateNetworkSettings(
    sandboxId: string,
    networkBlockAll?: boolean,
    networkAllowList?: string,
    networkLimitEgress?: boolean,
  ): Promise<void> {
    const payload: UpdateNetworkSettingsDTO = {
      networkBlockAll: networkBlockAll,
      networkAllowList: networkAllowList,
      networkLimitEgress: networkLimitEgress,
    }

    await this.jobService.createJob(
      null,
      JobType.UPDATE_SANDBOX_NETWORK_SETTINGS,
      this.runner.id,
      ResourceType.SANDBOX,
      sandboxId,
      payload,
    )

    this.logger.debug(
      `Created UPDATE_SANDBOX_NETWORK_SETTINGS job for sandbox ${sandboxId} on runner ${this.runner.id}`,
    )
  }
}
