/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron } from '@nestjs/schedule'
import { FindOptionsWhere, In, Not, Raw, Repository } from 'typeorm'
import { Runner } from '../entities/runner.entity'
import { CreateRunnerDto } from '../dto/create-runner.dto'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { RunnerRegion } from '../enums/runner-region.enum'
import { RunnerApiFactory } from '../runner-api/runnerApi'
import { RunnerState } from '../enums/runner-state.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { OnEvent } from '@nestjs/event-emitter'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'
import { SandboxState } from '../enums/sandbox-state.enum'
import { Sandbox } from '../entities/sandbox.entity'
import { SnapshotRunner } from '../entities/snapshot-runner.entity'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'

@Injectable()
export class RunnerService {
  private readonly logger = new Logger(RunnerService.name)
  private checkingRunners = false

  constructor(
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    private readonly runnerApiFactory: RunnerApiFactory,
    @InjectRepository(Sandbox)
    private readonly sandboxRepository: Repository<Sandbox>,
    @InjectRepository(SnapshotRunner)
    private readonly snapshotRunnerRepository: Repository<SnapshotRunner>,
  ) {}

  async create(createRunnerDto: CreateRunnerDto): Promise<Runner> {
    // Validate region and class
    if (!this.isValidRegion(createRunnerDto.region)) {
      throw new Error('Invalid region')
    }
    if (!this.isValidClass(createRunnerDto.class)) {
      throw new Error('Invalid class')
    }

    const runner = new Runner()
    runner.domain = createRunnerDto.domain
    runner.apiUrl = createRunnerDto.apiUrl
    runner.apiKey = createRunnerDto.apiKey
    runner.cpu = createRunnerDto.cpu
    runner.memory = createRunnerDto.memory
    runner.disk = createRunnerDto.disk
    runner.gpu = createRunnerDto.gpu
    runner.gpuType = createRunnerDto.gpuType
    runner.used = 0
    runner.capacity = createRunnerDto.capacity
    runner.region = createRunnerDto.region
    runner.class = createRunnerDto.class

    return this.runnerRepository.save(runner)
  }

  async findAll(): Promise<Runner[]> {
    return this.runnerRepository.find()
  }

  findOne(id: string): Promise<Runner | null> {
    return this.runnerRepository.findOneBy({ id })
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

  @Cron('45 * * * * *')
  private async handleCheckRunners() {
    if (this.checkingRunners) {
      return
    }
    this.checkingRunners = true
    const runners = await this.runnerRepository.find({
      where: {
        unschedulable: Not(true),
      },
    })
    for (const runner of runners) {
      this.logger.debug(`Checking runner ${runner.id}`)
      try {
        // Do something with the runner
        const runnerApi = this.runnerApiFactory.createRunnerApi(runner)
        await runnerApi.healthCheck()
        await this.runnerRepository.update(runner.id, {
          state: RunnerState.READY,
          lastChecked: new Date(),
        })

        await this.recalculateRunnerUsage(runner.id)
      } catch (e) {
        if (e.code === 'ECONNREFUSED') {
          this.logger.error('Runner not reachable')
        } else {
          this.logger.error(`Error checking runner ${runner.id}: ${e.message}`)
          this.logger.error(e)
        }

        await this.runnerRepository.update(runner.id, {
          state: RunnerState.UNRESPONSIVE,
          lastChecked: new Date(),
        })
      }
    }
    this.checkingRunners = false
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

  private isValidRegion(region: RunnerRegion): boolean {
    return Object.values(RunnerRegion).includes(region)
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
}

export class GetRunnerParams {
  region?: RunnerRegion
  sandboxClass?: SandboxClass
  snapshotRef?: string
  excludedRunnerIds?: string[]
}
