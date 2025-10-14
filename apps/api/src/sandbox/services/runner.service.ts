/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { FindOptionsWhere, In, MoreThanOrEqual, Not, Repository } from 'typeorm'
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
import { RedisLockProvider } from '../common/redis-lock.provider'
import { TypedConfigService } from '../../config/typed-config.service'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'

@Injectable()
export class RunnerService {
  private readonly logger = new Logger(RunnerService.name)
  private readonly scoreConfig: AvailabilityScoreConfig

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
    private readonly redisLockProvider: RedisLockProvider,
    private readonly configService: TypedConfigService,
  ) {
    this.scoreConfig = this.getAvailabilityScoreConfig()
  }

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
    runner.region = createRunnerDto.region
    runner.class = createRunnerDto.class
    runner.version = createRunnerDto.version

    return this.runnerRepository.save(runner)
  }

  async findAll(): Promise<Runner[]> {
    return this.runnerRepository.find()
  }

  async findAllReady(): Promise<Runner[]> {
    return this.runnerRepository.find({
      where: {
        state: RunnerState.READY,
      },
    })
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

  async findByApiKey(apiKey: string): Promise<Runner | null> {
    return this.runnerRepository.findOneBy({ apiKey })
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
      availabilityScore: params.availabilityScoreThreshold
        ? MoreThanOrEqual(params.availabilityScoreThreshold)
        : MoreThanOrEqual(this.scoreConfig.availabilityThreshold),
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

    return runners.sort((a, b) => b.availabilityScore - a.availabilityScore).slice(0, 10)
  }

  async remove(id: string): Promise<void> {
    await this.runnerRepository.delete(id)
  }

  @OnEvent(SandboxEvents.STATE_UPDATED)
  async handleSandboxStateUpdate(event: SandboxStateUpdatedEvent) {
    if (![SandboxState.DESTROYED, SandboxState.CREATING, SandboxState.ARCHIVED].includes(event.newState)) {
      return
    }
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

  @Cron(CronExpression.EVERY_10_SECONDS, { name: 'check-runners', waitForCompletion: true })
  @LogExecution('check-runners')
  @WithInstrumentation()
  private async handleCheckRunners() {
    const lockKey = 'check-runners'
    const hasLock = await this.redisLockProvider.lock(lockKey, 60)
    if (!hasLock) {
      return
    }

    try {
      const runners = await this.runnerRepository.find({
        where: {
          state: Not(RunnerState.DECOMMISSIONED),
        },
        order: {
          lastChecked: {
            direction: 'ASC',
            nulls: 'FIRST',
          },
        },
        take: 100,
      })

      await Promise.allSettled(
        runners.map(async (runner) => {
          const shouldRetry = runner.state === RunnerState.READY
          const retryDelays = shouldRetry ? [500, 1000] : []

          for (let attempt = 0; attempt <= retryDelays.length; attempt++) {
            if (attempt > 0) {
              await new Promise((resolve) => setTimeout(resolve, retryDelays[attempt - 1]))
            }

            const abortController = new AbortController()
            let timeoutId: NodeJS.Timeout | null = null

            try {
              await Promise.race([
                (async () => {
                  this.logger.debug(`Checking runner ${runner.id}`)
                  const runnerAdapter = await this.runnerAdapterFactory.create(runner)

                  await runnerAdapter.healthCheck(abortController.signal)

                  let runnerInfo: RunnerInfo | undefined
                  let runnerInfoError: Error | undefined
                  try {
                    runnerInfo = await runnerAdapter.runnerInfo(abortController.signal)
                  } catch (e) {
                    this.logger.warn(
                      `Failed to get runner info for runner ${runner.id}: ${e.message}. Setting runner score to -1`,
                    )
                    runnerInfoError = e
                  }

                  await this.updateRunnerStatus(runner.id, runnerInfo, runnerInfoError)
                })(),
                new Promise((_, reject) => {
                  timeoutId = setTimeout(() => {
                    abortController.abort()
                    reject(new Error('Health check timeout'))
                  }, 5000)
                }),
              ])

              if (timeoutId) {
                clearTimeout(timeoutId)
              }
              return // Success, exit retry loop
            } catch (e) {
              if (timeoutId) {
                clearTimeout(timeoutId)
              }

              if (e.message === 'Health check timeout') {
                this.logger.error(`Runner ${runner.id} health check timed out after 3 seconds`)
              } else if (e.code === 'ECONNREFUSED') {
                this.logger.error(`Runner ${runner.id} not reachable`)
              } else if (e.name === 'AbortError') {
                this.logger.error(`Runner ${runner.id} health check was aborted due to timeout`)
              } else {
                this.logger.error(`Error checking runner ${runner.id}`, e)
              }

              // If last attempt, mark as unresponsive
              if (attempt === retryDelays.length) {
                await this.updateRunnerState(runner.id, RunnerState.UNRESPONSIVE)
              }
            }
          }
        }),
      )
    } finally {
      await this.redisLockProvider.unlock(lockKey)
    }
  }

  private async updateRunnerStatus(runnerId: string, runnerInfo?: RunnerInfo, runnerInfoError?: Error) {
    const runner = await this.runnerRepository.findOne({ where: { id: runnerId } })
    if (!runner) {
      this.logger.error(`Runner ${runnerId} not found when trying to update status`)
      return
    }

    if (runner.state === RunnerState.DECOMMISSIONED) {
      this.logger.debug(`Runner ${runnerId} is decommissioned, not updating status`)
      return
    }

    const updateData: Partial<Runner> = {
      runnerInfoError: runnerInfoError ? runnerInfoError.message : null,
      lastChecked: new Date(),
    }

    if (runnerInfoError) {
      updateData.availabilityScore = -1
      await this.runnerRepository.update(runnerId, updateData)
      return
    }

    updateData.state = RunnerState.READY

    if (runnerInfo && runnerInfo.metrics) {
      updateData.currentCpuUsagePercentage = runnerInfo.metrics.currentCpuUsagePercentage
      updateData.currentCpuLoadAverage = runnerInfo.metrics.currentCpuLoadAverage
      updateData.currentMemoryUsagePercentage = runnerInfo.metrics.currentMemoryUsagePercentage
      updateData.currentDiskUsagePercentage = runnerInfo.metrics.currentDiskUsagePercentage
      updateData.currentAllocatedCpu = runnerInfo.metrics.currentAllocatedCpu
      updateData.currentAllocatedMemoryGiB = runnerInfo.metrics.currentAllocatedMemoryGiB
      updateData.currentAllocatedDiskGiB = runnerInfo.metrics.currentAllocatedDiskGiB
      updateData.currentSnapshotCount = runnerInfo.metrics.currentSnapshotCount
    } else {
      this.logger.warn(`Runner ${runnerId} didn't send health metrics`)
    }

    updateData.availabilityScore = this.calculateAvailabilityScore(runnerId, {
      cpuUsagePercentage: updateData.currentCpuUsagePercentage,
      cpuLoadAverage: updateData.currentCpuLoadAverage,
      memoryUsagePercentage: updateData.currentMemoryUsagePercentage,
      diskUsagePercentage: updateData.currentDiskUsagePercentage,
      allocatedCpu: updateData.currentAllocatedCpu,
      allocatedMemoryGiB: updateData.currentAllocatedMemoryGiB,
      allocatedDiskGiB: updateData.currentAllocatedDiskGiB,
      runnerCpu: runner.cpu,
      runnerMemoryGiB: runner.memoryGiB,
      runnerDiskGiB: runner.diskGiB,
    })

    await this.runnerRepository.update(runnerId, updateData)
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

    if (availableRunners.length === 0) {
      throw new BadRequestError('No available runners')
    }

    // Get random runner from the best available runners
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

  private calculateAvailabilityScore(runnerId: string, params: AvailabilityScoreParams): number {
    if (
      params.cpuUsagePercentage < 0 ||
      params.cpuLoadAverage < 0 ||
      params.memoryUsagePercentage < 0 ||
      params.diskUsagePercentage < 0 ||
      params.allocatedCpu < 0 ||
      params.allocatedMemoryGiB < 0 ||
      params.allocatedDiskGiB < 0
    ) {
      this.logger.warn(
        `Runner ${runnerId} has negative values for CPU, load, memory, disk, allocated CPU, allocated memory, or allocated disk`,
      )
      return 0
    }

    return this.calculateTOPSISScore(params)
  }

  private calculateTOPSISScore(params: AvailabilityScoreParams): number {
    const current = [
      params.cpuUsagePercentage,
      params.memoryUsagePercentage,
      params.diskUsagePercentage,
      // Allocation ratios percentage
      (params.allocatedCpu / params.runnerCpu) * 100,
      (params.allocatedMemoryGiB / params.runnerMemoryGiB) * 100,
      (params.allocatedDiskGiB / params.runnerDiskGiB) * 100,
    ]

    // Calculate weighted Euclidean distances
    let distanceToIdeal = 0
    let distanceToAntiIdeal = 0

    for (let i = 0; i < current.length; i++) {
      // Normalize to 0-1 scale
      const normalizedCurrent = current[i] / 100
      const normalizedIdeal = this.scoreConfig.targetValues.ideal[i] / 100
      const normalizedAntiIdeal = this.scoreConfig.targetValues.antiIdeal[i] / 100

      distanceToIdeal += this.scoreConfig.weights[i] * Math.pow(normalizedCurrent - normalizedIdeal, 2)
      distanceToAntiIdeal += this.scoreConfig.weights[i] * Math.pow(normalizedCurrent - normalizedAntiIdeal, 2)
    }

    distanceToIdeal = Math.sqrt(distanceToIdeal)
    distanceToAntiIdeal = Math.sqrt(distanceToAntiIdeal)

    // TOPSIS relative closeness score (0 to 1)
    let topsisScore = distanceToAntiIdeal / (distanceToIdeal + distanceToAntiIdeal)

    // Apply exponential penalties for critical thresholds
    let penaltyMultiplier = 1

    if (params.cpuUsagePercentage >= this.scoreConfig.penalty.thresholds.cpu) {
      penaltyMultiplier *= Math.exp(
        -this.scoreConfig.penalty.exponents.cpu * (params.cpuUsagePercentage - this.scoreConfig.penalty.thresholds.cpu),
      )
    }

    if (params.cpuLoadAverage >= this.scoreConfig.penalty.thresholds.cpuLoadAvg) {
      penaltyMultiplier *= Math.exp(
        -this.scoreConfig.penalty.exponents.cpuLoadAvg *
          (params.cpuLoadAverage - this.scoreConfig.penalty.thresholds.cpuLoadAvg),
      )
    }

    if (params.memoryUsagePercentage >= this.scoreConfig.penalty.thresholds.memory) {
      penaltyMultiplier *= Math.exp(
        -this.scoreConfig.penalty.exponents.memory *
          (params.memoryUsagePercentage - this.scoreConfig.penalty.thresholds.memory),
      )
    }

    if (params.diskUsagePercentage >= this.scoreConfig.penalty.thresholds.disk) {
      penaltyMultiplier *= Math.exp(
        -this.scoreConfig.penalty.exponents.disk *
          (params.diskUsagePercentage - this.scoreConfig.penalty.thresholds.disk),
      )
    }

    // Apply penalty
    topsisScore *= penaltyMultiplier

    return Math.round(topsisScore * 100)
  }

  private getAvailabilityScoreConfig(): AvailabilityScoreConfig {
    return {
      availabilityThreshold: this.configService.getOrThrow('runnerScore.thresholds.availability'),
      weights: [
        this.configService.getOrThrow('runnerScore.weights.cpuUsage'),
        this.configService.getOrThrow('runnerScore.weights.memoryUsage'),
        this.configService.getOrThrow('runnerScore.weights.diskUsage'),
        this.configService.getOrThrow('runnerScore.weights.allocatedCpu'),
        this.configService.getOrThrow('runnerScore.weights.allocatedMemory'),
        this.configService.getOrThrow('runnerScore.weights.allocatedDisk'),
      ],
      penalty: {
        exponents: {
          cpu: this.configService.getOrThrow('runnerScore.penalty.exponents.cpu'),
          cpuLoadAvg: this.configService.getOrThrow('runnerScore.penalty.exponents.cpuLoadAvg'),
          memory: this.configService.getOrThrow('runnerScore.penalty.exponents.memory'),
          disk: this.configService.getOrThrow('runnerScore.penalty.exponents.disk'),
        },
        thresholds: {
          cpu: this.configService.getOrThrow('runnerScore.penalty.thresholds.cpu'),
          cpuLoadAvg: this.configService.getOrThrow('runnerScore.penalty.thresholds.cpuLoadAvg'),
          memory: this.configService.getOrThrow('runnerScore.penalty.thresholds.memory'),
          disk: this.configService.getOrThrow('runnerScore.penalty.thresholds.disk'),
        },
      },
      targetValues: {
        ideal: [
          this.configService.getOrThrow('runnerScore.targetValues.ideal.cpu'),
          this.configService.getOrThrow('runnerScore.targetValues.ideal.memory'),
          this.configService.getOrThrow('runnerScore.targetValues.ideal.disk'),
          this.configService.getOrThrow('runnerScore.targetValues.ideal.allocCpu'),
          this.configService.getOrThrow('runnerScore.targetValues.ideal.allocMem'),
          this.configService.getOrThrow('runnerScore.targetValues.ideal.allocDisk'),
        ],
        antiIdeal: [
          this.configService.getOrThrow('runnerScore.targetValues.antiIdeal.cpu'),
          this.configService.getOrThrow('runnerScore.targetValues.antiIdeal.memory'),
          this.configService.getOrThrow('runnerScore.targetValues.antiIdeal.disk'),
          this.configService.getOrThrow('runnerScore.targetValues.antiIdeal.allocCpu'),
          this.configService.getOrThrow('runnerScore.targetValues.antiIdeal.allocMem'),
          this.configService.getOrThrow('runnerScore.targetValues.antiIdeal.allocDisk'),
        ],
      },
    }
  }
}

export class GetRunnerParams {
  region?: string
  sandboxClass?: SandboxClass
  snapshotRef?: string
  excludedRunnerIds?: string[]
  availabilityScoreThreshold?: number
}

interface AvailabilityScoreParams {
  cpuUsagePercentage: number
  cpuLoadAverage: number
  memoryUsagePercentage: number
  diskUsagePercentage: number
  allocatedCpu: number
  allocatedMemoryGiB: number
  allocatedDiskGiB: number
  runnerCpu: number
  runnerMemoryGiB: number
  runnerDiskGiB: number
}

interface AvailabilityScoreConfig {
  availabilityThreshold: number
  weights: number[]
  penalty: {
    exponents: {
      cpu: number
      cpuLoadAvg: number
      memory: number
      disk: number
    }
    thresholds: {
      cpu: number
      cpuLoadAvg: number
      memory: number
      disk: number
    }
  }
  targetValues: {
    ideal: number[]
    antiIdeal: number[]
  }
}
