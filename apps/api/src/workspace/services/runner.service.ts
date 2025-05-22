/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRepository } from '@nestjs/typeorm'
import { Cron } from '@nestjs/schedule'
import { In, Not, Repository } from 'typeorm'
import { Runner } from '../entities/runner.entity'
import { CreateRunnerDto } from '../dto/create-runner.dto'
import { WorkspaceClass } from '../enums/workspace-class.enum'
import { RunnerRegion } from '../enums/runner-region.enum'
import { RunnerApiFactory } from '../runner-api/runnerApi'
import { RunnerState } from '../enums/runner-state.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'
import { WorkspaceEvents } from '../constants/workspace-events.constants'
import { OnEvent } from '@nestjs/event-emitter'
import { WorkspaceStateUpdatedEvent } from '../events/workspace-state-updated.event'
import { WorkspaceState } from '../enums/workspace-state.enum'
import { Workspace } from '../entities/workspace.entity'
import { ImageRunner } from '../entities/image-runner.entity'
import { ImageRunnerState } from '../enums/image-runner-state.enum'
import { ImageManager } from '../managers/image.manager'

@Injectable()
export class RunnerService {
  private readonly logger = new Logger(RunnerService.name)
  private checkingRunners = false

  constructor(
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    private readonly runnerApiFactory: RunnerApiFactory,
    @InjectRepository(Workspace)
    private readonly workspaceRepository: Repository<Workspace>,
    @InjectRepository(ImageRunner)
    private readonly imageRunnerRepository: Repository<ImageRunner>,
    private readonly imageStateManager: ImageManager,
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

  async findAvailableRunners(
    region: RunnerRegion,
    workspaceClass: WorkspaceClass,
    imageRef?: string,
  ): Promise<Runner[]> {
    const whereCondition: any = {
      state: ImageRunnerState.READY,
    }

    if (imageRef !== undefined) {
      whereCondition.imageRef = imageRef
    }

    const imageRunners = await this.imageRunnerRepository.find({
      where: whereCondition,
    })

    const runners = await this.runnerRepository.find({
      where: {
        id: In(imageRunners.map((imageRunner) => imageRunner.runnerId)),
        state: RunnerState.READY,
        region,
        class: workspaceClass,
        unschedulable: Not(true),
      },
    })
    return runners
      .filter((runner) => runner.used < runner.capacity)
      .sort((a, b) => a.used / a.capacity - b.used / b.capacity)
      .slice(0, 10)
  }

  async remove(id: string): Promise<void> {
    await this.runnerRepository.delete(id)
  }

  @OnEvent(WorkspaceEvents.STATE_UPDATED)
  async handleWorkspaceStateUpdate(event: WorkspaceStateUpdatedEvent) {
    if (![WorkspaceState.DESTROYED, WorkspaceState.CREATING, WorkspaceState.ARCHIVED].includes(event.newState)) {
      return
    }

    await this.recalculateRunnerUsage(event.workspace.runnerId)
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
    const workspaces = await this.workspaceRepository.find({
      where: {
        runnerId: runner.id,
        state: Not(WorkspaceState.DESTROYED),
      },
    })
    runner.used = workspaces.length

    await this.runnerRepository.save(runner)
  }

  private isValidRegion(region: RunnerRegion): boolean {
    return Object.values(RunnerRegion).includes(region)
  }

  private isValidClass(workspaceClass: WorkspaceClass): boolean {
    return Object.values(WorkspaceClass).includes(workspaceClass)
  }

  async updateSchedulingStatus(id: string, unschedulable: boolean): Promise<Runner> {
    const runner = await this.runnerRepository.findOne({ where: { id } })
    if (!runner) {
      throw new Error('Runner not found')
    }

    runner.unschedulable = unschedulable
    return this.runnerRepository.save(runner)
  }

  async getRandomAvailableRunner(
    region: RunnerRegion,
    workspaceClass: WorkspaceClass,
    imageRef?: string,
  ): Promise<string> {
    const availableRunners = await this.findAvailableRunners(region, workspaceClass, imageRef)

    //  TODO: implement a better algorithm to get a random available runner based on the runner's usage

    if (availableRunners.length === 0) {
      throw new BadRequestError('No available runners')
    }

    availableRunners.sort((a, b) => a.used / a.capacity - b.used / b.capacity)
    //  use the first 10 runners
    const optimalRunners = availableRunners.slice(0, 10)

    // Get random runner from available runners using inclusive bounds
    const randomIntFromInterval = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1) + min)

    return optimalRunners[randomIntFromInterval(0, optimalRunners.length - 1)].id
  }

  async getImageRunner(runnerId, imageRef: string): Promise<ImageRunner> {
    return this.imageRunnerRepository.findOne({
      where: {
        runnerId: runnerId,
        imageRef,
      },
    })
  }

  async getImageRunners(imageRef: string): Promise<ImageRunner[]> {
    return this.imageRunnerRepository.find({
      where: {
        imageRef,
      },
      order: {
        state: 'ASC', // Sorts state BUILDING_IMAGE before ERROR
        createdAt: 'ASC', // Sorts first runner to start building image on top
      },
    })
  }

  async createImageRunner(
    runnerId: string,
    imageRef: string,
    state: ImageRunnerState,
    errorReason?: string,
  ): Promise<void> {
    const imageRunner = new ImageRunner()
    imageRunner.runnerId = runnerId
    imageRunner.imageRef = imageRef
    imageRunner.state = state
    if (errorReason) {
      imageRunner.errorReason = errorReason
    }
    await this.imageRunnerRepository.save(imageRunner)
    if (state != ImageRunnerState.ERROR) {
      this.imageStateManager.syncRunnerImageState(imageRunner)
    }
  }
}
