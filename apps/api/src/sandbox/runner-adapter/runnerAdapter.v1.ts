/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import axios from 'axios'
import axiosDebug from 'axios-debug-log'

import { Injectable, Logger } from '@nestjs/common'
import { RunnerAdapter, RunnerSandboxInfo, RunnerSandboxState } from './runnerAdapter'
import { Runner } from '../entities/runner.entity'
import { Configuration, SandboxApi, EnumsSandboxState, SnapshotsApi } from '@daytonaio/runner-api-client'
import { DockerRegistryService } from '../../docker-registry/services/docker-registry.service'
import { InjectRepository } from '@nestjs/typeorm'
import { Sandbox } from '../entities/sandbox.entity'
import { Repository } from 'typeorm'
import { Snapshot } from '../entities/snapshot.entity'
import { BuildInfo } from '../entities/build-info.entity'

const isDebugEnabled = process.env.DEBUG === 'true'

@Injectable()
export class RunnerAdapterV1 implements RunnerAdapter {
  constructor(
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    private readonly dockerRegistryService: DockerRegistryService,
  ) {}

  private readonly logger = new Logger(RunnerAdapterV1.name)
  private apiClientSandbox: SandboxApi
  private apiClientSnapshot: SnapshotsApi

  private convertSandboxState(state: EnumsSandboxState): RunnerSandboxState {
    switch (state) {
      case EnumsSandboxState.SandboxStateCreating:
        return RunnerSandboxState.CREATING
      case EnumsSandboxState.SandboxStateRestoring:
        return RunnerSandboxState.RESTORING
      case EnumsSandboxState.SandboxStateDestroyed:
        return RunnerSandboxState.DESTROYED
      case EnumsSandboxState.SandboxStateDestroying:
        return RunnerSandboxState.DESTROYING
      case EnumsSandboxState.SandboxStateStarted:
        return RunnerSandboxState.STARTED
      case EnumsSandboxState.SandboxStateStopped:
        return RunnerSandboxState.STOPPED
      case EnumsSandboxState.SandboxStateStarting:
        return RunnerSandboxState.STARTING
      case EnumsSandboxState.SandboxStateStopping:
        return RunnerSandboxState.STOPPING
      case EnumsSandboxState.SandboxStateResizing:
        return RunnerSandboxState.UNKNOWN // Resizing is not in RunnerSandboxState
      case EnumsSandboxState.SandboxStateError:
        return RunnerSandboxState.ERROR
      case EnumsSandboxState.SandboxStateUnknown:
        return RunnerSandboxState.UNKNOWN
      case EnumsSandboxState.SandboxStatePullingSnapshot:
        return RunnerSandboxState.PULLING_SNAPSHOT
      default:
        return RunnerSandboxState.UNKNOWN
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

  async buildSnapshot(buildInfo: BuildInfo, organizationId?: string): Promise<void> {
    await this.apiClientSnapshot.buildSnapshot({
      snapshot: buildInfo.snapshotRef,
      dockerfile: buildInfo.dockerfileContent,
      organizationId: organizationId,
      context: buildInfo.contextHashes,
    })
  }

  async create(sandboxId: string): Promise<void> {
    const sandbox = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })

    const registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)

    await this.apiClientSandbox.create({
      id: sandboxId,
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

  async createBackup(sandboxId: string, backupSnapshotName: string): Promise<void> {
    const sandbox = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })

    const registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)

    await this.apiClientSandbox.createBackup(sandboxId, {
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

  async snapshot(sandboxId: string, snapshotName: string): Promise<void> {
    const sandbox = await this.sandboxRepository.findOneByOrFail({
      id: sandboxId,
    })

    const registry = await this.dockerRegistryService.findOne(sandbox.backupRegistryId)

    await this.apiClientSandbox.createBackup(sandboxId, {
      registry: {
        url: registry.url,
        username: registry.username,
        password: registry.password,
      },
      snapshot: snapshotName,
    })
  }

  async snapshotExists(snapshotName: string): Promise<boolean> {
    const response = await this.apiClientSnapshot.snapshotExists(snapshotName)
    return response.data.exists
  }

  async pullSnapshot(snapshotName: string): Promise<void> {
    const snapshot = await this.snapshotRepository.findOneByOrFail({
      name: snapshotName,
    })

    //  TODO: get registry from snapshot

    const registry = await this.dockerRegistryService.getDefaultInternalRegistry()

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
