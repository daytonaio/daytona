/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron } from '@nestjs/schedule'
import { FindOptionsWhere, In, Not, Raw, Repository } from 'typeorm'
import { Runner } from '../entities/runner.entity'
import { CreateRunnerDto } from '../dto/create-runner.dto'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { RunnerState } from '../enums/runner-state.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { OnEvent } from '@nestjs/event-emitter'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'
import { SandboxState } from '../enums/sandbox-state.enum'
import { Sandbox } from '../entities/sandbox.entity'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { Snapshot } from '../entities/snapshot.entity'
import { RunnerSnapshotDto } from '../dto/runner-snapshot.dto'
import { RunnerAdapterFactory, RunnerInfo } from '../runner-adapter/runnerAdapter'

@Injectable()
export class RunnerService {
  private readonly logger = new Logger(RunnerService.name)
  private checkingRunners = false

  constructor(
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    private readonly runnerAdapterFactory: RunnerAdapterFactory,
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    @InjectRepository(SnapshotRunner)
    private readonly snapshotRunnerRepository: Repository<SnapshotRunner>,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
  ) {}

  async create(createRunnerDto: CreateRunnerDto): Promise<Runner> {
    // Validate region and class
    if (createRunnerDto.region.trim().length === 0) {
      throw new Error('Invalid region')
    }
    if (!this.isValidClass(createRunnerDto.class)) {
      throw new Error('Invalid class')
    }

    const runner = new Runner()
    runner.domain = createRunnerDto.domain
    runner.apiUrl = createRunnerDto.apiUrl
    runner.proxyUrl = createRunnerDto.proxyUrl
    runner.apiKey = createRunnerDto.apiKey
    runner.cpu = createRunnerDto.cpu
    runner.memoryGiB = createRunnerDto.memoryGiB
    runner.diskGiB = createRunnerDto.diskGiB
    runner.gpu = createRunnerDto.gpu
    runner.gpuType = createRunnerDto.gpuType
    runner.used = 0
    runner.capacity = createRunnerDto.capacity
    runner.region = createRunnerDto.region
    runner.class = createRunnerDto.class
    runner.version = createRunnerDto.version

    return this.runnerRepository.save(runner)
  }

  async findAll(): Promise<Runner[]> {
    return this.runnerRepository.find()
  }

  async findOne(id: string): Promise<Runner | null> {
    return this.runnerRepository.findOneBy({ id })
  }

  async findByIds(runnerIds: string[]): Promise<Runner[]> {
    if (runnerIds.length === 0) {
      return []
    }

    return this.runnerRepository.find({
      where: { id: In(runnerIds) },
    })
  }

  async findBySandboxId(sandboxId: string): Promise<Runner | null> {
    const sandbox = await this.sandboxRepository.findOneBy({ id: sandboxId, state: Not(SandboxState.DESTROYED) })
    if (!sandbox) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} not found`)
    }
    if (!sandbox.runnerId) {
      throw new NotFoundException(`Sandbox with ID ${sandboxId} does not have a runner`)
    }

    return this.runnerRepository.findOneBy({ id: sandbox.runnerId })
  }

  async findAvailableRunners(params: GetRunnerParams): Promise<Runner[]> {
    const runnerFilter: FindOptionsWhere<Runner> = {
      state: RunnerState.READY,
      unschedulable: Not(true),
      used: Raw((alias) => `${alias} < capacity`),
    }

    if (params.snapshotRef !== undefined) {
      const snapshotRunners = await this.snapshotRunnerRepository.find({
        where: {
          state: SnapshotRunnerState.READY,
          snapshotRef: params.snapshotRef,
        },
      })

      let runnerIds = snapshotRunners.map((snapshotRunner) => snapshotRunner.runnerId)

      if (params.excludedRunnerIds?.length) {
        runnerIds = runnerIds.filter((id) => !params.excludedRunnerIds.includes(id))
      }

      if (!runnerIds.length) {
        return []
      }

      runnerFilter.id = In(runnerIds)
    } else if (params.excludedRunnerIds?.length) {
      runnerFilter.id = Not(In(params.excludedRunnerIds))
    }

    if (params.region !== undefined) {
      runnerFilter.region = params.region
    }

    if (params.sandboxClass !== undefined) {
      runnerFilter.class = params.sandboxClass
    }

    const runners = await this.runnerRepository.find({
      where: runnerFilter,
    })

    return runners.sort((a, b) => a.used / a.capacity - b.used / b.capacity).slice(0, 10)
  }

  async remove(id: string): Promise<void> {
    await this.runnerRepository.delete(id)
  }

  @OnEvent(SandboxEvents.STATE_UPDATED)
  async handleSandboxStateUpdate(event: SandboxStateUpdatedEvent) {
    if (![SandboxState.DESTROYED, SandboxState.CREATING, SandboxState.ARCHIVED].includes(event.newState)) {
      return
    }

    await this.recalculateRunnerUsage(event.sandbox.runnerId)
  }

  private async updateRunnerState(runnerId: string, newState: RunnerState): Promise<void> {
    const runner = await this.runnerRepository.findOne({ where: { id: runnerId } })
    if (!runner) {
      this.logger.error(`Runner ${runnerId} not found when trying to update state`)
      return
    }

    // Don't change state if runner is decommissioned
    if (runner.state === RunnerState.DECOMMISSIONED) {
      this.logger.debug(`Runner ${runnerId} is decommissioned, not updating state`)
      return
    }

    await this.runnerRepository.update(runnerId, {
      state: newState,
      lastChecked: new Date(),
    })
  }

  @Cron('45 * * * * *')
  private async handleCheckRunners() {
    if (this.checkingRunners) {
      return
    }
    this.checkingRunners = true
    const runners = await this.runnerRepository.find({
      where: {
        state: Not(RunnerState.DECOMMISSIONED),
      },
    })
    for (const runner of runners) {
      this.logger.debug(`Checking runner ${runner.id}`)
      try {
        // Get health check with status metrics
        const runnerAdapter = await this.runnerAdapterFactory.create(runner)
        await runnerAdapter.healthCheck()

        let runnerInfo: RunnerInfo | undefined
        try {
          runnerInfo = await runnerAdapter.runnerInfo()
        } catch (e) {
          this.logger.warn(`Failed to get runner info for runner ${runner.id}: ${e.message}`)
        }

        await this.updateRunnerStatus(runner.id, runnerInfo)

        await this.recalculateRunnerUsage(runner.id)
      } catch (e) {
        if (e.code === 'ECONNREFUSED') {
          this.logger.error('Runner not reachable')
        } else {
          this.logger.error(`Error checking runner ${runner.id}: ${e.message}`)
          this.logger.error(e)
        }

        await this.updateRunnerState(runner.id, RunnerState.UNRESPONSIVE)
      }
    }
    this.checkingRunners = false
  }

  private async updateRunnerStatus(runnerId: string, runnerInfo?: RunnerInfo) {
    const runner = await this.runnerRepository.findOne({ where: { id: runnerId } })
    if (!runner) {
      this.logger.error(`Runner ${runnerId} not found when trying to update status`)
      return
    }

    if (runner.state === RunnerState.DECOMMISSIONED) {
      this.logger.debug(`Runner ${runnerId} is decommissioned, not updating status`)
      return
    }

    const updateData: any = {
      state: RunnerState.READY,
      lastChecked: new Date(),
    }

    const metrics = runnerInfo?.metrics

    if (metrics && typeof metrics.currentCpuUsagePercentage !== 'undefined') {
      updateData.currentCpuUsagePercentage = metrics.currentCpuUsagePercentage || 0
      updateData.currentMemoryUsagePercentage = metrics.currentMemoryUsagePercentage || 0
      updateData.currentDiskUsagePercentage = metrics.currentDiskUsagePercentage || 0
      updateData.currentAllocatedCpu = metrics.currentAllocatedCpu || 0
      updateData.currentAllocatedMemoryGiB = metrics.currentAllocatedMemoryGiB || 0
      updateData.currentAllocatedDiskGiB = metrics.currentAllocatedDiskGiB || 0
      updateData.currentSnapshotCount = metrics.currentSnapshotCount || 0

      updateData.availabilityScore = this.calculateAvailabilityScore({
        cpuUsage: updateData.currentCpuUsagePercentage,
        memoryUsage: updateData.currentMemoryUsagePercentage,
        diskUsage: updateData.currentDiskUsagePercentage,
        allocatedCpu: updateData.currentAllocatedCpu,
        allocatedMemoryGiB: updateData.currentAllocatedMemoryGiB,
        allocatedDiskGiB: updateData.currentAllocatedDiskGiB,
        capacity: runner.capacity,
        runnerCpu: runner.cpu,
        runnerMemoryGiB: runner.memoryGiB,
        runnerDiskGiB: runner.diskGiB,
      })
    } else {
      this.logger.warn(`Runner ${runnerId} didn't send health metrics`)
    }

    await this.runnerRepository.update(runnerId, updateData)
  }

  async recalculateRunnerUsage(runnerId: string) {
    const runner = await this.runnerRepository.findOne({ where: { id: runnerId } })
    if (!runner) {
      throw new Error('Runner not found')
    }
    //  recalculate runner usage
    const sandboxes = await this.sandboxRepository.find({
      where: {
        runnerId: runner.id,
        state: Not(SandboxState.DESTROYED),
      },
    })
    runner.used = sandboxes.length

    await this.runnerRepository.save(runner)
  }

  private isValidClass(sandboxClass: SandboxClass): boolean {
    return Object.values(SandboxClass).includes(sandboxClass)
  }

  async updateSchedulingStatus(id: string, unschedulable: boolean): Promise<Runner> {
    const runner = await this.runnerRepository.findOne({ where: { id } })
    if (!runner) {
      throw new Error('Runner not found')
    }

    runner.unschedulable = unschedulable
    return this.runnerRepository.save(runner)
  }

  async getRandomAvailableRunner(params: GetRunnerParams): Promise<Runner> {
    const availableRunners = await this.findAvailableRunners(params)

    //  TODO: implement a better algorithm to get a random available runner based on the runner's usage

    if (availableRunners.length === 0) {
      throw new BadRequestError('No available runners')
    }

    // Get random runner from available runners using inclusive bounds
    const randomIntFromInterval = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1) + min)

    return availableRunners[randomIntFromInterval(0, availableRunners.length - 1)]
  }

  async getSnapshotRunner(runnerId, snapshotRef: string): Promise<SnapshotRunner> {
    return this.snapshotRunnerRepository.findOne({
      where: {
        runnerId: runnerId,
        snapshotRef: snapshotRef,
      },
    })
  }

  async getSnapshotRunners(snapshotRef: string): Promise<SnapshotRunner[]> {
    return this.snapshotRunnerRepository.find({
      where: {
        snapshotRef: snapshotRef,
      },
      order: {
        state: 'ASC', // Sorts state BUILDING_SNAPSHOT before ERROR
        createdAt: 'ASC', // Sorts first runner to start building snapshot on top
      },
    })
  }

  async createSnapshotRunner(
    runnerId: string,
    snapshotRef: string,
    state: SnapshotRunnerState,
    errorReason?: string,
  ): Promise<void> {
    const snapshotRunner = new SnapshotRunner()
    snapshotRunner.runnerId = runnerId
    snapshotRunner.snapshotRef = snapshotRef
    snapshotRunner.state = state
    if (errorReason) {
      snapshotRunner.errorReason = errorReason
    }
    await this.snapshotRunnerRepository.save(snapshotRunner)
  }

  async getRunnersWithMultipleSnapshotsBuilding(maxSnapshotCount = 2): Promise<string[]> {
    const runners = await this.sandboxRepository
      .createQueryBuilder('sandbox')
      .select('sandbox.runnerId')
      .where('sandbox.state = :state', { state: SandboxState.BUILDING_SNAPSHOT })
      .andWhere('sandbox.buildInfoSnapshotRef IS NOT NULL')
      .groupBy('sandbox.runnerId')
      .having('COUNT(DISTINCT sandbox.buildInfoSnapshotRef) > :maxSnapshotCount', { maxSnapshotCount })
      .getRawMany()

    return runners.map((item) => item.runnerId)
  }

  async getRunnersBySnapshotRef(ref: string): Promise<RunnerSnapshotDto[]> {
    const snapshotRunners = await this.snapshotRunnerRepository.find({
      where: {
        snapshotRef: ref,
        state: Not(SnapshotRunnerState.ERROR),
      },
      select: ['runnerId', 'id'],
    })

    // Extract distinct runnerIds from snapshot runners
    const runnerIds = [...new Set(snapshotRunners.map((sr) => sr.runnerId))]

    // Find all runners with these IDs
    const runners = await this.runnerRepository.find({
      where: { id: In(runnerIds) },
      select: ['id', 'domain'],
    })

    this.logger.debug(`Found ${runners.length} runners with IDs: ${runners.map((r) => r.id).join(', ')}`)

    // Map to DTO format, including the snapshot runner ID
    return runners.map((runner) => {
      const snapshotRunner = snapshotRunners.find((sr) => sr.runnerId === runner.id)
      return new RunnerSnapshotDto(snapshotRunner.id, runner.id, runner.domain)
    })
  }

  private calculateAvailabilityScore(params: AvailabilityScoreParams): number {
    let penalty = 0

    // CPU Penalty (reduced impact, starts at 40%)
    if (params.cpuUsage > 0) {
      const low = (Math.min(params.cpuUsage, 40) / 40) * 1 // 0-40%: gradual 0-1 points
      const medium = params.cpuUsage > 40 ? ((Math.min(params.cpuUsage, 85) - 40) / 45) * 18 : 0 // 40-85%: 18 more points
      const high = params.cpuUsage > 85 ? ((Math.min(params.cpuUsage, 95) - 85) / 10) * 12 : 0 // 85-95%: 12 more points
      const critical = params.cpuUsage > 95 ? (params.cpuUsage - 95) * 3 : 0 // 95%+: 3 points per % over 95%
      penalty += low + medium + high + critical
    }

    // RAM Penalty (starts at 40%, high impact, critical at 85%)
    if (params.memoryUsage > 0) {
      const low = (Math.min(params.memoryUsage, 40) / 40) * 3 // 0-40%: gradual 0-3 points
      const medium = params.memoryUsage > 40 ? ((Math.min(params.memoryUsage, 85) - 40) / 45) * 35 : 0 // 40-85%: 35 more points
      const critical = params.memoryUsage > 85 ? (params.memoryUsage - 85) * 6 : 0 // 85%+: 6 points per % over 85%
      penalty += low + medium + critical
    }

    // Disk Penalty (high impact, critical at 85%)
    if (params.diskUsage > 0) {
      const low = (Math.min(params.diskUsage, 60) / 60) * 3 // 0-60%: gradual 0-3 points
      const medium = params.diskUsage > 60 ? ((Math.min(params.diskUsage, 85) - 60) / 25) * 25 : 0 // 60-85%: 25 more points
      const critical = params.diskUsage > 85 ? (params.diskUsage - 85) * 7 : 0 // 85%+: 7 points per % over 85%
      penalty += low + medium + critical
    }

    // Allocated CPU Ratio Penalty (minimal impact, starts at 250% of runner CPU capacity)
    const allocatedCpuRatio = (params.allocatedCpu / params.runnerCpu) * 100
    if (allocatedCpuRatio > 250) {
      const low = allocatedCpuRatio > 250 ? ((Math.min(allocatedCpuRatio, 300) - 250) / 50) * 0.5 : 0 // 250-300%: 0.5 points
      const medium = allocatedCpuRatio > 300 ? ((Math.min(allocatedCpuRatio, 350) - 300) / 50) * 1 : 0 // 300-350%: 1 more points
      const critical = allocatedCpuRatio > 350 ? (allocatedCpuRatio - 350) * 0.025 : 0 // 350%+: 0.025 points per % over 350%
      penalty += low + medium + critical
    }

    // Allocated Memory Ratio Penalty (minimal impact, starts at 250% of runner memory capacity)
    const allocatedMemoryRatio = (params.allocatedMemoryGiB / params.runnerMemoryGiB) * 100
    if (allocatedMemoryRatio > 250) {
      const low = allocatedMemoryRatio > 250 ? ((Math.min(allocatedMemoryRatio, 300) - 250) / 50) * 0.5 : 0 // 250-300%: 0.5 points
      const medium = allocatedMemoryRatio > 300 ? ((Math.min(allocatedMemoryRatio, 350) - 300) / 50) * 1 : 0 // 300-350%: 1 more points
      const critical = allocatedMemoryRatio > 350 ? (allocatedMemoryRatio - 350) * 0.025 : 0 // 350%+: 0.025 points per % over 350%
      penalty += low + medium + critical
    }

    // Allocated Disk Ratio Penalty (minimal impact, starts at 250% of runner disk capacity)
    const allocatedDiskRatio = (params.allocatedDiskGiB / params.runnerDiskGiB) * 100
    if (allocatedDiskRatio > 250) {
      const low = allocatedDiskRatio > 250 ? ((Math.min(allocatedDiskRatio, 300) - 250) / 50) * 0.5 : 0 // 250-300%: 0.5 points
      const medium = allocatedDiskRatio > 300 ? ((Math.min(allocatedDiskRatio, 350) - 300) / 50) * 1 : 0 // 300-350%: 1 more points
      const critical = allocatedDiskRatio > 350 ? (allocatedDiskRatio - 350) * 0.025 : 0 // 350%+: 0.025 points per % over 350%
      penalty += low + medium + critical
    }

    return Math.max(0, Math.round(100 - penalty))
  }
}

export class GetRunnerParams {
  region?: string
  sandboxClass?: SandboxClass
  snapshotRef?: string
  excludedRunnerIds?: string[]
}

interface AvailabilityScoreParams {
  cpuUsage: number
  memoryUsage: number
  diskUsage: number
  allocatedCpu: number
  allocatedMemoryGiB: number
  allocatedDiskGiB: number
  capacity: number
  runnerCpu: number
  runnerMemoryGiB: number
  runnerDiskGiB: number
}
