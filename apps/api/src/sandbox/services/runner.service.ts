/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  BadRequestException,
  ConflictException,
  HttpException,
  HttpStatus,
  Inject,
  Injectable,
  Logger,
  NotFoundException,
} from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron, CronExpression } from '@nestjs/schedule'
import { DataSource, FindOptionsWhere, In, MoreThanOrEqual, Not, Repository } from 'typeorm'
import { Runner } from '../entities/runner.entity'
import { CreateRunnerInternalDto } from '../dto/create-runner-internal.dto'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { RunnerState } from '../enums/runner-state.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { EventEmitter2 } from '@nestjs/event-emitter'
import { SandboxState } from '../enums/sandbox-state.enum'
import { Sandbox } from '../entities/sandbox.entity'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { RunnerSnapshotDto } from '../dto/runner-snapshot.dto'
import { RunnerAdapterFactory, RunnerInfo } from '../runner-adapter/runnerAdapter'
import { RedisLockProvider } from '../common/redis-lock.provider'
import { TypedConfigService } from '../../config/typed-config.service'
import { LogExecution } from '../../common/decorators/log-execution.decorator'
import { WithInstrumentation } from '../../common/decorators/otel.decorator'
import { RegionService } from '../../region/services/region.service'
import { RUNNER_NAME_REGEX } from '../constants/runner-name-regex.constant'
import { RegionType } from '../../region/enums/region-type.enum'
import { RunnerDto } from '../dto/runner.dto'
import { RunnerEvents } from '../constants/runner-events'
import { RunnerStateUpdatedEvent } from '../events/runner-state-updated.event'
import { RunnerDeletedEvent } from '../events/runner-deleted.event'
import { generateApiKeyValue } from '../../common/utils/api-key'
import { RunnerFullDto } from '../dto/runner-full.dto'
import { Snapshot } from '../entities/snapshot.entity'

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
    private readonly redisLockProvider: RedisLockProvider,
    private readonly configService: TypedConfigService,
    private readonly regionService: RegionService,
    @InjectRepository(Snapshot)
    private readonly snapshotRepository: Repository<Snapshot>,
    @Inject(EventEmitter2)
    private eventEmitter: EventEmitter2,
    private readonly dataSource: DataSource,
  ) {}

  /**
   * @throws {BadRequestException} If the runner name or class is invalid.
   * @throws {NotFoundException} If the region is not found.
   * @throws {ConflictException} If a runner with the same values already exists.
   */
  async create(createRunnerDto: CreateRunnerInternalDto): Promise<{
    runner: Runner
    apiKey: string
  }> {
    if (!RUNNER_NAME_REGEX.test(createRunnerDto.name)) {
      throw new BadRequestException('Runner name must contain only letters, numbers, underscores, periods, and hyphens')
    }
    if (createRunnerDto.name.length < 2 || createRunnerDto.name.length > 255) {
      throw new BadRequestException('Runner name must be between 3 and 255 characters')
    }

    const apiKey = createRunnerDto.apiKey ?? generateApiKeyValue()

    let runner: Runner

    switch (createRunnerDto.apiVersion) {
      case '0':
        runner = new Runner({
          region: createRunnerDto.regionId,
          name: createRunnerDto.name,
          apiVersion: createRunnerDto.apiVersion,
          apiKey: apiKey,
          cpu: createRunnerDto.cpu,
          memoryGiB: createRunnerDto.memoryGiB,
          diskGiB: createRunnerDto.diskGiB,
          domain: createRunnerDto.domain,
          apiUrl: createRunnerDto.apiUrl,
          proxyUrl: createRunnerDto.proxyUrl,
          appVersion: createRunnerDto.appVersion,
        })
        break
      case '2':
        runner = new Runner({
          region: createRunnerDto.regionId,
          name: createRunnerDto.name,
          apiVersion: createRunnerDto.apiVersion,
          apiKey: apiKey,
          appVersion: createRunnerDto.appVersion,
        })
        break
      default:
        throw new BadRequestException('Invalid runner version')
    }

    try {
      const savedRunner = await this.runnerRepository.save(runner)
      return { runner: savedRunner, apiKey }
    } catch (error) {
      if (error.code === '23505') {
        if (error.detail.includes('domain')) {
          throw new ConflictException('This domain is already in use')
        }
        if (error.detail.includes('name')) {
          throw new ConflictException(`Runner with name ${createRunnerDto.name} already exists in this region`)
        }
        throw new ConflictException('A runner with these values already exists')
      }
      throw error
    }
  }

  async findAllFull(): Promise<RunnerFullDto[]> {
    const runners = await this.runnerRepository.find()

    const regionIds = new Set(runners.map((runner) => runner.region))
    const regions = await this.regionService.findByIds(Array.from(regionIds))

    const regionTypeMap = new Map<string, RegionType>()
    regions.forEach((region) => {
      regionTypeMap.set(region.id, region.regionType)
    })

    return runners.map((runner) => RunnerFullDto.fromRunner(runner, regionTypeMap.get(runner.region)))
  }

  async findAllByRegion(regionId: string): Promise<RunnerDto[]> {
    const runners = await this.runnerRepository.find({
      where: {
        region: regionId,
      },
    })

    return runners.map(RunnerDto.fromRunner)
  }

  async findAllByRegionFull(regionId: string): Promise<RunnerFullDto[]> {
    const runners = await this.runnerRepository.find({
      where: {
        region: regionId,
      },
    })

    const region = await this.regionService.findOne(regionId)

    return runners.map((runner) => RunnerFullDto.fromRunner(runner, region?.regionType))
  }

  async findAllByOrganization(organizationId: string, regionType?: RegionType): Promise<RunnerDto[]> {
    const regions = await this.regionService.findAllByOrganization(organizationId, regionType)
    const regionIds = regions.map((region) => region.id)

    const runners = await this.runnerRepository.find({
      where: {
        region: In(regionIds),
      },
    })

    return runners.map(RunnerDto.fromRunner)
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

  async findOneFullOrFail(id: string): Promise<RunnerFullDto> {
    const runner = await this.findOne(id)
    if (!runner) {
      throw new NotFoundException('Runner not found')
    }

    const region = await this.regionService.findOne(runner.region)

    return RunnerFullDto.fromRunner(runner, region?.regionType)
  }

  async findOneByDomain(domain: string): Promise<Runner | null> {
    return this.runnerRepository.findOneBy({ domain })
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

  async getRegionId(runnerId: string): Promise<string> {
    const runner = await this.runnerRepository.findOne({
      where: {
        id: runnerId,
      },
      select: ['region'],
      loadEagerRelations: false,
    })

    if (!runner || !runner.region) {
      throw new NotFoundException('Runner not found')
    }

    return runner.region
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

    if (params.regions?.length) {
      runnerFilter.region = In(params.regions)
    }

    if (params.sandboxClass !== undefined) {
      runnerFilter.class = params.sandboxClass
    }

    const runners = await this.runnerRepository.find({
      where: runnerFilter,
    })

    return runners.sort((a, b) => b.availabilityScore - a.availabilityScore).slice(0, 10)
  }

  /**
   * @throws {NotFoundException} If the runner is not found.
   * @throws {HttpException} If the runner is not unschedulable.
   * @throws {HttpException} If the runner has sandboxes associated with it.
   */
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

    const sandboxCount = await this.sandboxRepository.count({
      where: { runnerId: id, state: Not(In([SandboxState.ARCHIVED, SandboxState.DESTROYED])) },
    })
    if (sandboxCount > 0) {
      throw new HttpException(
        'Cannot delete runner which has sandboxes associated with it',
        HttpStatus.PRECONDITION_REQUIRED,
      )
    }

    await this.dataSource.transaction(async (em) => {
      await em.delete(Runner, id)
      await this.eventEmitter.emitAsync(RunnerEvents.DELETED, new RunnerDeletedEvent(em, id))
    })
  }

  async updateRunnerHealth(
    runnerId: string,
    domain?: string,
    apiUrl?: string,
    proxyUrl?: string,
    metrics?: {
      currentCpuUsagePercentage?: number
      currentMemoryUsagePercentage?: number
      currentDiskUsagePercentage?: number
      currentAllocatedCpu?: number
      currentAllocatedMemoryGiB?: number
      currentAllocatedDiskGiB?: number
      currentSnapshotCount?: number
      cpu?: number
      memoryGiB?: number
      diskGiB?: number
    },
    appVersion?: string,
  ): Promise<void> {
    const runner = await this.runnerRepository.findOne({ where: { id: runnerId } })
    if (!runner) {
      this.logger.error(`Runner ${runnerId} not found when trying to update health`)
      return
    }

    if (runner.state === RunnerState.DECOMMISSIONED) {
      this.logger.debug(`Runner ${runnerId} is decommissioned, not updating health`)
      return
    }

    const updateData: Partial<Runner> = {
      state: RunnerState.READY,
      lastChecked: new Date(),
    }

    if (domain) {
      updateData.domain = domain
    }

    if (apiUrl) {
      updateData.apiUrl = apiUrl
    }

    if (proxyUrl) {
      updateData.proxyUrl = proxyUrl
    }

    if (appVersion) {
      updateData.appVersion = appVersion
    }

    if (metrics) {
      updateData.currentCpuUsagePercentage = metrics.currentCpuUsagePercentage || 0
      updateData.currentMemoryUsagePercentage = metrics.currentMemoryUsagePercentage || 0
      updateData.currentDiskUsagePercentage = metrics.currentDiskUsagePercentage || 0
      updateData.currentAllocatedCpu = metrics.currentAllocatedCpu || 0
      updateData.currentAllocatedMemoryGiB = metrics.currentAllocatedMemoryGiB || 0
      updateData.currentAllocatedDiskGiB = metrics.currentAllocatedDiskGiB || 0
      updateData.currentSnapshotCount = metrics.currentSnapshotCount || 0
      updateData.cpu = metrics.cpu
      updateData.memoryGiB = metrics.memoryGiB
      updateData.diskGiB = metrics.diskGiB

      updateData.availabilityScore = this.calculateAvailabilityScore(runnerId, {
        cpuUsage: updateData.currentCpuUsagePercentage,
        memoryUsage: updateData.currentMemoryUsagePercentage,
        diskUsage: updateData.currentDiskUsagePercentage,
        allocatedCpu: updateData.currentAllocatedCpu,
        allocatedMemoryGiB: updateData.currentAllocatedMemoryGiB,
        allocatedDiskGiB: updateData.currentAllocatedDiskGiB,
        runnerCpu: updateData.cpu || runner.cpu,
        runnerMemoryGiB: updateData.memoryGiB || runner.memoryGiB,
        runnerDiskGiB: updateData.diskGiB || runner.diskGiB,
      })
    }

    await this.runnerRepository.update(runnerId, updateData)
    this.logger.debug(`Updated health for runner ${runnerId}`)

    this.eventEmitter.emit(
      RunnerEvents.STATE_UPDATED,
      new RunnerStateUpdatedEvent(runner, runner.state, updateData.state),
    )
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

    this.eventEmitter.emit(RunnerEvents.STATE_UPDATED, new RunnerStateUpdatedEvent(runner, runner.state, newState))
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
        where: [
          {
            apiVersion: '0',
            state: Not(RunnerState.DECOMMISSIONED),
          },
          {
            // v2 runners report health via healthcheck endpoint, so we only check if the health is stale (lastChecked timestamp)
            apiVersion: '2',
            state: RunnerState.READY,
          },
        ],
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
          // v2 runners report health via healthcheck endpoint, check based on lastChecked timestamp
          if (runner.apiVersion === '2') {
            await this.checkRunnerV2Health(runner)
            return
          }

          // v0 runners: imperative health check via adapter
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

                  await this.updateRunnerHealth(
                    runner.id,
                    undefined,
                    undefined,
                    undefined,
                    runnerInfo?.metrics,
                    runnerInfo?.appVersion,
                  )
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

  /**
   * Check v2 runner health based on lastChecked timestamp.
   * v2 runners report health via the healthcheck endpoint, so we check if lastChecked is within threshold.
   */
  private async checkRunnerV2Health(runner: Runner): Promise<void> {
    if (!runner.lastChecked) {
      return
    }

    // v2 runners report health every ~10 seconds via the healthcheck endpoint
    // Allow 60 seconds (6 missed healthchecks) before marking as UNRESPONSIVE
    const healthCheckThresholdMs = 60 * 1000

    const timeSinceLastCheck = Date.now() - runner.lastChecked.getTime()

    if (timeSinceLastCheck > healthCheckThresholdMs) {
      this.logger.warn(
        `v2 Runner ${runner.id} health check stale (last: ${Math.round(timeSinceLastCheck / 1000)}s ago), marking as UNRESPONSIVE`,
      )
      // TODO: if api is restarted, all runners will go unresponsive
      await this.updateRunnerState(runner.id, RunnerState.UNRESPONSIVE)
    }
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

  async getInitialRunnerBySnapshotId(snapshotId: string): Promise<Runner> {
    const snapshot = await this.snapshotRepository.findOne({ where: { id: snapshotId } })
    if (!snapshot) {
      throw new NotFoundException('Snapshot runner not found')
    }
    if (!snapshot.initialRunnerId) {
      throw new BadRequestException('Initial runner not found')
    }

    const runner = await this.runnerRepository.findOne({ where: { id: snapshot.initialRunnerId } })

    if (!runner) {
      throw new NotFoundException('Runner not found')
    }

    return runner
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
  regions?: string[]
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
