/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { HttpException, HttpStatus, Injectable, Logger, NotFoundException } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { FindOptionsWhere, In, MoreThanOrEqual, Not, Repository } from 'typeorm'
import { Runner } from '../entities/runner.entity'
import { CreateRunnerInternalDto } from '../dto/create-runner-internal.dto'
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
import { Organization } from '../../organization/entities/organization.entity'
import { RegionService } from '../../region/services/region.service'
import * as crypto from 'crypto'

@Injectable()
export class RunnerService {
  private readonly logger = new Logger(RunnerService.name)

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
    private readonly regionService: RegionService,
  ) {}

  private generateRunnerToken(): string {
    return `dtn_${crypto.randomBytes(32).toString('hex')}`
  }

  public generateRunnerTokenHash(value: string): string {
    return crypto.createHash('sha256').update(value).digest('hex')
  }

  private getTokenPrefix(token: string): string {
    return token.substring(0, 3)
  }

  private getTokenSuffix(token: string): string {
    return token.slice(-3)
  }

  async create(
    createRunnerDto: CreateRunnerInternalDto,
    organization?: Organization,
  ): Promise<{
    runner: Runner
    token: string
  }> {
    // Validate region and class
    const region = await this.regionService.findOne(createRunnerDto.regionId, organization?.id)
    if (!region) {
      throw new NotFoundException('Region not found')
    }

    if (!this.isValidClass(createRunnerDto.class)) {
      throw new BadRequestError('Invalid class')
    }

    const token = createRunnerDto.token ?? this.generateRunnerToken()

    const runner = new Runner()
    runner.domain = createRunnerDto.domain
    runner.apiUrl = createRunnerDto.apiUrl
    runner.proxyUrl = createRunnerDto.proxyUrl
    runner.apiKey = token
    runner.tokenHash = this.generateRunnerTokenHash(token)
    runner.tokenPrefix = this.getTokenPrefix(token)
    runner.tokenSuffix = this.getTokenSuffix(token)
    runner.cpu = createRunnerDto.cpu
    runner.memoryGiB = createRunnerDto.memoryGiB
    runner.diskGiB = createRunnerDto.diskGiB
    runner.gpu = createRunnerDto.gpu
    runner.gpuType = createRunnerDto.gpuType
    runner.regionId = createRunnerDto.regionId
    runner.class = createRunnerDto.class
    runner.version = createRunnerDto.version

    const savedRunner = await this.runnerRepository.save(runner)
    return { runner: savedRunner, token }
  }

  async findAll(organizationId?: string, regionName?: string): Promise<Runner[]> {
    if (organizationId && regionName) {
      return this.findAllByRegionName(organizationId, regionName)
    } else if (organizationId) {
      return this.findAllByOrganizationId(organizationId)
    } else {
      return this.runnerRepository.find()
    }
  }

  async findAllByRegionName(organizationId: string, regionName: string): Promise<Runner[]> {
    const region = await this.regionService.findOneByName(regionName, organizationId)
    if (!region) {
      throw new NotFoundException('Region not found')
    }

    return this.runnerRepository.find({
      where: {
        regionId: region.id,
      },
    })
  }

  async findAllByOrganizationId(organizationId: string): Promise<Runner[]> {
    const regions = await this.regionService.findAll(organizationId)
    const regionIds = regions.map((region) => region.id)

    return this.runnerRepository.find({
      where: {
        regionId: In(regionIds),
      },
    })
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

  async findByToken(token: string): Promise<Runner | null> {
    return this.runnerRepository.findOneBy({ tokenHash: this.generateRunnerTokenHash(token) })
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
        : MoreThanOrEqual(this.configService.getOrThrow('runnerUsage.availabilityScoreThreshold')),
    }

    const excludedRunnerIds = params.excludedRunnerIds?.length
      ? params.excludedRunnerIds.filter((id) => !!id)
      : undefined

    if (params.snapshotRef !== undefined) {
      const snapshotRunners = await this.snapshotRunnerRepository.find({
        where: {
          state: SnapshotRunnerState.READY,
          snapshotRef: params.snapshotRef,
        },
      })

      let runnerIds = snapshotRunners.map((snapshotRunner) => snapshotRunner.runnerId)

      if (excludedRunnerIds?.length) {
        runnerIds = runnerIds.filter((id) => !excludedRunnerIds.includes(id))
      }

      if (!runnerIds.length) {
        return []
      }

      runnerFilter.id = In(runnerIds)
    } else if (excludedRunnerIds?.length) {
      runnerFilter.id = Not(In(excludedRunnerIds))
    }

    if (params.regionId !== undefined) {
      runnerFilter.regionId = params.regionId
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
    const runner = await this.findOne(id)
    if (!runner) {
      throw new NotFoundException('Runner not found')
    }

    if (!runner.unschedulable) {
      throw new HttpException(
        'Cannot delete runner which is available for scheduling sandboxes',
        HttpStatus.PRECONDITION_REQUIRED,
      )
    }

    const sandboxes = await this.sandboxRepository.find({
      where: { runnerId: id, state: Not(In([SandboxState.ARCHIVED, SandboxState.DESTROYED])) },
    })
    if (sandboxes.length > 0) {
      throw new HttpException(
        'Cannot delete runner which has sandboxes associated with it',
        HttpStatus.PRECONDITION_REQUIRED,
      )
    }

    await this.runnerRepository.remove(runner)
  }

  async getRegionId(runnerId: string): Promise<string> {
    const runner = await this.runnerRepository.findOne({
      where: {
        id: runnerId,
      },
      select: ['regionId'],
      loadEagerRelations: false,
    })

    if (!runner || !runner.regionId) {
      throw new NotFoundException('Runner not found')
    }

    return runner.regionId
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

    runner.state = newState
    runner.lastChecked = new Date()
    await this.runnerRepository.save(runner)
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

            const runnerHealthTimeoutSeconds = this.configService.get('runnerHealthTimeout')

            try {
              await Promise.race([
                (async () => {
                  this.logger.debug(`Checking runner ${runner.id}`)
                  const runnerAdapter = await this.runnerAdapterFactory.create(runner)

                  await runnerAdapter.healthCheck(abortController.signal)

                  let runnerInfo: RunnerInfo | undefined
                  try {
                    runnerInfo = await runnerAdapter.runnerInfo(abortController.signal)
                  } catch (e) {
                    this.logger.warn(`Failed to get runner info for runner ${runner.id}: ${e.message}`)
                  }

                  await this.updateRunnerStatus(runner.id, runnerInfo)
                })(),
                new Promise((_, reject) => {
                  timeoutId = setTimeout(() => {
                    abortController.abort()
                    reject(new Error('Health check timeout'))
                  }, runnerHealthTimeoutSeconds * 1000)
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
                this.logger.error(
                  `Runner ${runner.id} health check timed out after ${runnerHealthTimeoutSeconds} seconds`,
                )
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

    const metrics = runnerInfo?.metrics

    if (metrics && typeof metrics.currentCpuUsagePercentage !== 'undefined') {
      runner.currentCpuUsagePercentage = metrics.currentCpuUsagePercentage || 0
      runner.currentMemoryUsagePercentage = metrics.currentMemoryUsagePercentage || 0
      runner.currentDiskUsagePercentage = metrics.currentDiskUsagePercentage || 0
      runner.currentAllocatedCpu = metrics.currentAllocatedCpu || 0
      runner.currentAllocatedMemoryGiB = metrics.currentAllocatedMemoryGiB || 0
      runner.currentAllocatedDiskGiB = metrics.currentAllocatedDiskGiB || 0
      runner.currentSnapshotCount = metrics.currentSnapshotCount || 0

      runner.availabilityScore = this.calculateAvailabilityScore(runnerId, {
        cpuUsage: runner.currentCpuUsagePercentage,
        memoryUsage: runner.currentMemoryUsagePercentage,
        diskUsage: runner.currentDiskUsagePercentage,
        allocatedCpu: runner.currentAllocatedCpu,
        allocatedMemoryGiB: runner.currentAllocatedMemoryGiB,
        allocatedDiskGiB: runner.currentAllocatedDiskGiB,
        runnerCpu: runner.cpu,
        runnerMemoryGiB: runner.memoryGiB,
        runnerDiskGiB: runner.diskGiB,
      })
    } else {
      this.logger.warn(`Runner ${runnerId} didn't send health metrics`)
    }

    runner.state = RunnerState.READY
    runner.lastChecked = new Date()

    await this.runnerRepository.save(runner)
  }

  private isValidClass(sandboxClass: SandboxClass): boolean {
    return Object.values(SandboxClass).includes(sandboxClass)
  }

  async updateSchedulingStatus(id: string, unschedulable: boolean): Promise<Runner> {
    const runner = await this.runnerRepository.findOne({ where: { id } })
    if (!runner) {
      throw new NotFoundException('Runner not found')
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

  async getSnapshotRunner(runnerId: string, snapshotRef: string): Promise<SnapshotRunner> {
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
        snapshotRef,
      },
      order: {
        state: 'ASC', // Sorts state BUILDING_SNAPSHOT before ERROR
        createdAt: 'ASC', // Sorts first runner to start building snapshot on top
      },
    })
  }

  async createSnapshotRunnerEntry(
    runnerId: string,
    snapshotRef: string,
    state?: SnapshotRunnerState,
    errorReason?: string,
  ): Promise<void> {
    try {
      const snapshotRunner = new SnapshotRunner()
      snapshotRunner.runnerId = runnerId
      snapshotRunner.snapshotRef = snapshotRef
      if (state) {
        snapshotRunner.state = state
      }
      if (errorReason) {
        snapshotRunner.errorReason = errorReason
      }
      await this.snapshotRunnerRepository.save(snapshotRunner)
    } catch (error) {
      if (error.code === '23505') {
        // PostgreSQL unique violation error code - entry already exists, allow it
        this.logger.debug(
          `SnapshotRunner entry already exists for runnerId: ${runnerId}, snapshotRef: ${snapshotRef}. Continuing...`,
        )
        return
      }
      throw error // Re-throw any other errors
    }
  }

  // TODO: combine getRunnersWithMultipleSnapshotsBuilding and getRunnersWithMultipleSnapshotsPulling?

  async getRunnersWithMultipleSnapshotsBuilding(maxSnapshotCount = 6): Promise<string[]> {
    const runners = await this.sandboxRepository
      .createQueryBuilder('sandbox')
      .select('sandbox.runnerId', 'runnerId')
      .where('sandbox.state = :state', { state: SandboxState.BUILDING_SNAPSHOT })
      .andWhere('sandbox.buildInfoSnapshotRef IS NOT NULL')
      .groupBy('sandbox.runnerId')
      .having('COUNT(DISTINCT sandbox.buildInfoSnapshotRef) > :maxSnapshotCount', { maxSnapshotCount })
      .getRawMany()

    return runners.map((item) => item.runnerId)
  }

  async getRunnersWithMultipleSnapshotsPulling(maxSnapshotCount = 6): Promise<string[]> {
    const runners = await this.snapshotRunnerRepository
      .createQueryBuilder('snapshot_runner')
      .select('snapshot_runner.runnerId')
      .where('snapshot_runner.state = :state', { state: SnapshotRunnerState.PULLING_SNAPSHOT })
      .groupBy('snapshot_runner.runnerId')
      .having('COUNT(*) > :maxSnapshotCount', { maxSnapshotCount })
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
      params.cpuUsage < 0 ||
      params.memoryUsage < 0 ||
      params.diskUsage < 0 ||
      params.allocatedCpu < 0 ||
      params.allocatedMemoryGiB < 0 ||
      params.allocatedDiskGiB < 0
    ) {
      this.logger.warn(
        `Runner ${runnerId} has negative values for CPU, memory, disk, allocated CPU, allocated memory, or allocated disk`,
      )
      return 0
    }

    return this.calculateTOPSISScore(params)
  }

  private calculateTOPSISScore(params: AvailabilityScoreParams): number {
    // Define ideal (best) and anti-ideal (worst) values
    const ideal = {
      cpu: 0,
      memory: 0,
      disk: 0,
      allocCpu: 100, // 100% means no overallocation
      allocMem: 100,
      allocDisk: 100,
    }

    const antiIdeal = {
      cpu: 100,
      memory: 100,
      disk: 100,
      allocCpu: 500, // 500% means severe overallocation
      allocMem: 500,
      allocDisk: 500,
    }

    // Weights based on your requirements
    const weights = [
      this.configService.getOrThrow('runnerUsage.cpuUsageWeight'),
      this.configService.getOrThrow('runnerUsage.memoryUsageWeight'),
      this.configService.getOrThrow('runnerUsage.diskUsageWeight'),
      this.configService.getOrThrow('runnerUsage.allocatedCpuWeight'),
      this.configService.getOrThrow('runnerUsage.allocatedMemoryWeight'),
      this.configService.getOrThrow('runnerUsage.allocatedDiskWeight'),
    ]

    const cpuPenaltyExponent = this.configService.getOrThrow('runnerUsage.cpuPenaltyExponent')
    const memoryPenaltyExponent = this.configService.getOrThrow('runnerUsage.memoryPenaltyExponent')
    const diskPenaltyExponent = this.configService.getOrThrow('runnerUsage.diskPenaltyExponent')

    const cpuPenaltyThreshold = this.configService.getOrThrow('runnerUsage.cpuPenaltyThreshold')
    const memoryPenaltyThreshold = this.configService.getOrThrow('runnerUsage.memoryPenaltyThreshold')
    const diskPenaltyThreshold = this.configService.getOrThrow('runnerUsage.diskPenaltyThreshold')

    // Calculate allocation ratios
    const allocatedCpuRatio = (params.allocatedCpu / params.runnerCpu) * 100
    const allocatedMemoryRatio = (params.allocatedMemoryGiB / params.runnerMemoryGiB) * 100
    const allocatedDiskRatio = (params.allocatedDiskGiB / params.runnerDiskGiB) * 100

    // Current values array
    const current = [
      params.cpuUsage,
      params.memoryUsage,
      params.diskUsage,
      allocatedCpuRatio,
      allocatedMemoryRatio,
      allocatedDiskRatio,
    ]

    // Ideal and anti-ideal arrays
    const idealValues = [ideal.cpu, ideal.memory, ideal.disk, ideal.allocCpu, ideal.allocMem, ideal.allocDisk]

    const antiIdealValues = [
      antiIdeal.cpu,
      antiIdeal.memory,
      antiIdeal.disk,
      antiIdeal.allocCpu,
      antiIdeal.allocMem,
      antiIdeal.allocDisk,
    ]

    // Calculate weighted Euclidean distances
    let distanceToIdeal = 0
    let distanceToAntiIdeal = 0

    for (let i = 0; i < current.length; i++) {
      const normalizedCurrent = current[i] / 100 // Normalize to 0-1 scale for allocation ratios >100%
      const normalizedIdeal = idealValues[i] / 100
      const normalizedAntiIdeal = antiIdealValues[i] / 100

      distanceToIdeal += weights[i] * Math.pow(normalizedCurrent - normalizedIdeal, 2)
      distanceToAntiIdeal += weights[i] * Math.pow(normalizedCurrent - normalizedAntiIdeal, 2)
    }

    distanceToIdeal = Math.sqrt(distanceToIdeal)
    distanceToAntiIdeal = Math.sqrt(distanceToAntiIdeal)

    // TOPSIS relative closeness score (0 to 1)
    let topsisScore = distanceToAntiIdeal / (distanceToIdeal + distanceToAntiIdeal)

    // Apply exponential penalties for critical thresholds
    let penaltyMultiplier = 1

    if (params.cpuUsage >= cpuPenaltyThreshold) {
      penaltyMultiplier *= Math.exp(-cpuPenaltyExponent * (params.cpuUsage - cpuPenaltyThreshold))
    }

    if (params.memoryUsage >= memoryPenaltyThreshold) {
      penaltyMultiplier *= Math.exp(-memoryPenaltyExponent * (params.memoryUsage - memoryPenaltyThreshold))
    }

    if (params.diskUsage >= diskPenaltyThreshold) {
      penaltyMultiplier *= Math.exp(-diskPenaltyExponent * (params.diskUsage - diskPenaltyThreshold))
    }

    // Apply penalty
    topsisScore *= penaltyMultiplier

    return Math.round(topsisScore * 100)
  }
}

export class GetRunnerParams {
  regionId?: string
  sandboxClass?: SandboxClass
  snapshotRef?: string
  excludedRunnerIds?: string[]
  availabilityScoreThreshold?: number
}

interface AvailabilityScoreParams {
  cpuUsage: number
  memoryUsage: number
  diskUsage: number
  allocatedCpu: number
  allocatedMemoryGiB: number
  allocatedDiskGiB: number
  runnerCpu: number
  runnerMemoryGiB: number
  runnerDiskGiB: number
}
