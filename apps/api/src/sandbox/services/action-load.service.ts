/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { OnEvent } from '@nestjs/event-emitter'
import { InjectRepository } from '@nestjs/typeorm'
import { Repository } from 'typeorm'
import { InjectRedis } from '@nestjs-modules/ioredis'
import { Redis } from 'ioredis'
import { Runner } from '../entities/runner.entity'
import { SandboxState } from '../enums/sandbox-state.enum'
import { SandboxDesiredState } from '../enums/sandbox-desired-state.enum'
import { SandboxEvents } from '../constants/sandbox-events.constants'
import { SandboxStateUpdatedEvent } from '../events/sandbox-state-updated.event'
import { TypedConfigService } from '../../config/typed-config.service'

/**
 * Service responsible for managing runner action load tracking.
 *
 * Action load represents the current workload on a runner from pending sandbox operations.
 * The action load penalty is derived from points and used to adjust runner availability scores
 * during sandbox assignment, helping distribute load evenly across runners.
 */
@Injectable()
export class ActionLoadService {
  private readonly logger = new Logger(ActionLoadService.name)

  constructor(
    @InjectRedis() private readonly redis: Redis,
    @InjectRepository(Runner)
    private readonly runnerRepository: Repository<Runner>,
    private readonly configService: TypedConfigService,
  ) {}

  private getActionLoadPointsRedisKey(runnerId: string): string {
    return `runner:${runnerId}:actionLoadPoints`
  }

  private getActionLoadPenaltyRedisKey(runnerId: string): string {
    return `runner:${runnerId}:actionLoadPenalty`
  }

  /**
   * Handle sandbox state updates for action load tracking.
   *
   * Action load model:
   * - ONE increment when runner starts working on a sandbox
   * - ONE decrement when job completes (sandbox reaches desired state or error)
   *
   * Increment triggers:
   * 1. Runner just assigned (oldRunnerId=null, runnerId!=null) - for build/pull/restore flows
   * 2. Entering CREATING state without buildInfo - for direct starts (runnerId set during insert)
   *
   * State transitions within a job (e.g., BUILDING_SNAPSHOT → UNKNOWN) do NOT
   * trigger increment or decrement - they're just progress within the same job.
   */
  @OnEvent(SandboxEvents.STATE_UPDATED)
  async handleSandboxStateUpdatedForActionLoad(event: SandboxStateUpdatedEvent): Promise<void> {
    const { sandbox, oldState, newState, oldRunnerId } = event

    // Skip if no runner assigned
    if (!sandbox.runnerId) {
      return
    }

    // INCREMENT: When runner starts working on a sandbox
    // Case 1: Runner just assigned (detected via oldRunnerId change)
    // Case 2: Entering CREATING state for direct starts (runnerId was set during insert, not update)
    const runnerJustAssigned = oldRunnerId === null && sandbox.runnerId !== null
    const directStartBeginning = newState === SandboxState.CREATING && !sandbox.buildInfo && oldRunnerId !== null

    if (runnerJustAssigned || directStartBeginning) {
      // Determine points based on the state the sandbox is entering
      let pointsState: SandboxState

      switch (newState) {
        case SandboxState.BUILDING_SNAPSHOT:
          pointsState = SandboxState.BUILDING_SNAPSHOT
          break
        case SandboxState.PULLING_SNAPSHOT:
          pointsState = SandboxState.PULLING_SNAPSHOT
          break
        case SandboxState.RESTORING:
          pointsState = SandboxState.RESTORING
          break
        default:
          // UNKNOWN (line 166 case), CREATING (direct start), or other
          pointsState = SandboxState.STARTING
          break
      }

      const points = this.calculateActionLoadPoints(pointsState, sandbox.desiredState)
      if (points > 0) {
        await this.incrementActionLoad(sandbox.runnerId, pointsState, sandbox.desiredState)
        this.logger.debug(
          `Incremented action load for runner ${sandbox.runnerId} by ${points} points (state=${pointsState}) for sandbox ${sandbox.id}`,
        )
      }
      return // Don't also check for decrement in the same event
    }

    // DECREMENT: When job completes (reaches desired state or error)
    const wasPending = oldState !== sandbox.desiredState.toString()
    const isNowResolved =
      newState === sandbox.desiredState.toString() ||
      newState === SandboxState.ERROR ||
      newState === SandboxState.BUILD_FAILED

    if (wasPending && isNowResolved) {
      // Determine points based on sandbox type
      let pointsState = oldState as SandboxState
      if (sandbox.desiredState === SandboxDesiredState.STARTED) {
        if (sandbox.buildInfo) {
          pointsState = SandboxState.BUILDING_SNAPSHOT
        } else if (sandbox.prevRunnerId !== null) {
          pointsState = SandboxState.RESTORING
        } else {
          pointsState = SandboxState.STARTING
        }
      }

      const points = this.calculateActionLoadPoints(pointsState, sandbox.desiredState)
      if (points > 0) {
        await this.decrementActionLoad(sandbox.runnerId, points)
        this.logger.debug(
          `Decremented action load for runner ${sandbox.runnerId} by ${points} points (state=${pointsState}) after sandbox ${sandbox.id} reached ${newState}`,
        )
      }
    }
  }

  /**
   * Calculate the action load points for a given state and desired state combination.
   */
  calculateActionLoadPoints(state: SandboxState, desiredState: SandboxDesiredState): number {
    const points = this.configService.getOrThrow('actionLoad.points')

    // <any> | destroyed = anyDestroyed points
    if (desiredState === SandboxDesiredState.DESTROYED) {
      return points.anyDestroyed
    }

    // started | stopped = startedStopped points
    if (state === SandboxState.STARTED && desiredState === SandboxDesiredState.STOPPED) {
      return points.startedStopped
    }

    // For desiredState = started
    if (desiredState === SandboxDesiredState.STARTED) {
      // building_snapshot | started = buildingSnapshotStarted points (declarative build)
      if (state === SandboxState.BUILDING_SNAPSHOT) {
        return points.buildingSnapshotStarted
      }

      // restoring | started = restoringStarted points (backup restoration)
      if (state === SandboxState.RESTORING) {
        return points.restoringStarted
      }

      // unknown | started = unknownStarted points
      if (state === SandboxState.UNKNOWN) {
        return points.unknownStarted
      }

      // stopped | started = stoppedStarted points
      if (state === SandboxState.STOPPED) {
        return points.stoppedStarted
      }

      // any other state | started = anyStarted points
      return points.anyStarted
    }

    return 0
  }

  /**
   * Increment the action load for a runner in Redis when a sandbox is assigned.
   */
  async incrementActionLoad(runnerId: string, state: SandboxState, desiredState: SandboxDesiredState): Promise<void> {
    const points = this.calculateActionLoadPoints(state, desiredState)
    if (points === 0) {
      return
    }

    const pointsKey = this.getActionLoadPointsRedisKey(runnerId)
    const penaltyKey = this.getActionLoadPenaltyRedisKey(runnerId)
    const ttlSeconds = 300

    const script = `
      local pointsKey = KEYS[1]
      local penaltyKey = KEYS[2]
      local increment = tonumber(ARGV[1])
      local ttl = tonumber(ARGV[2])
      local divisor = tonumber(ARGV[3])
      local maximum = tonumber(ARGV[4])

      local currentLoad = tonumber(redis.call("GET", pointsKey)) or 0
      local newLoad = currentLoad + increment
      redis.call("SET", pointsKey, newLoad, "EX", ttl)

      local penalty = math.floor(newLoad / divisor)
      if penalty > maximum then
        penalty = maximum
      end
      redis.call("SET", penaltyKey, penalty, "EX", ttl)

      return {newLoad, penalty}
    `

    const divisor = this.configService.getOrThrow('actionLoad.penalty.divisor')
    const maximum = this.configService.getOrThrow('actionLoad.penalty.maximum')

    await this.redis.eval(
      script,
      2,
      pointsKey,
      penaltyKey,
      points.toString(),
      ttlSeconds.toString(),
      divisor.toString(),
      maximum.toString(),
    )
  }

  /**
   * Decrement the action load for a runner in Redis when a sandbox transitions out of pending.
   */
  async decrementActionLoad(runnerId: string, points: number): Promise<void> {
    if (points === 0) {
      return
    }

    const pointsKey = this.getActionLoadPointsRedisKey(runnerId)
    const penaltyKey = this.getActionLoadPenaltyRedisKey(runnerId)
    const ttlSeconds = 300 // 5 minutes TTL as safeguard

    const script = `
      local pointsKey = KEYS[1]
      local penaltyKey = KEYS[2]
      local decrement = tonumber(ARGV[1])
      local ttl = tonumber(ARGV[2])
      local divisor = tonumber(ARGV[3])
      local maximum = tonumber(ARGV[4])

      local currentLoad = tonumber(redis.call("GET", pointsKey)) or 0
      local newLoad = currentLoad - decrement
      if newLoad < 0 then
        newLoad = 0
      end

      if newLoad > 0 then
        redis.call("SET", pointsKey, newLoad, "EX", ttl)
      else
        redis.call("DEL", pointsKey)
      end

      local penalty = math.floor(newLoad / divisor)
      if penalty > maximum then
        penalty = maximum
      end
      if penalty > 0 then
        redis.call("SET", penaltyKey, penalty, "EX", ttl)
      else
        redis.call("DEL", penaltyKey)
      end

      return {newLoad, penalty}
    `

    const divisor = this.configService.getOrThrow('actionLoad.penalty.divisor')
    const maximum = this.configService.getOrThrow('actionLoad.penalty.maximum')

    await this.redis.eval(
      script,
      2,
      pointsKey,
      penaltyKey,
      points.toString(),
      ttlSeconds.toString(),
      divisor.toString(),
      maximum.toString(),
    )
  }

  /**
   * Recalculate action load for a runner based on pending sandboxes.
   * First persists current Redis values to DB, then calculates new values from sandbox states.
   */
  async recalculateActionLoad(
    runnerId: string,
    pendingSandboxes: { state: SandboxState; desiredState: SandboxDesiredState }[],
    maxSandboxesPerRunner: number,
  ): Promise<void> {
    const pointsKey = this.getActionLoadPointsRedisKey(runnerId)
    const penaltyKey = this.getActionLoadPenaltyRedisKey(runnerId)
    const divisor = this.configService.getOrThrow('actionLoad.penalty.divisor')
    const maximum = this.configService.getOrThrow('actionLoad.penalty.maximum')
    const ttlSeconds = 300 // 5 minutes TTL as safeguard

    // First, get current values from Redis and persist to DB
    const [currentPointsStr, currentPenaltyStr] = await this.redis.mget(pointsKey, penaltyKey)
    const currentPoints = parseInt(currentPointsStr || '0', 10)
    const currentPenalty = parseInt(currentPenaltyStr || '0', 10)

    // Persist current Redis values to DB
    await this.runnerRepository.update(runnerId, {
      actionLoadPoints: currentPoints,
      actionLoadPenalty: currentPenalty,
    })

    // Calculate new points based on pending sandboxes
    let newPoints: number
    let newPenalty: number

    if (pendingSandboxes.length >= maxSandboxesPerRunner) {
      // Runner is at capacity - set maximum values
      newPoints = 9999
      newPenalty = maximum
    } else {
      // Sum up points from all pending sandbox state+desiredState combinations
      newPoints = pendingSandboxes.reduce((total, sandbox) => {
        return total + this.calculateActionLoadPoints(sandbox.state, sandbox.desiredState)
      }, 0)
      newPenalty = Math.min(Math.floor(newPoints / divisor), maximum)
    }

    // Store new values in Redis
    if (newPoints > 0) {
      await this.redis.set(pointsKey, newPoints.toString(), 'EX', ttlSeconds)
      await this.redis.set(penaltyKey, newPenalty.toString(), 'EX', ttlSeconds)
    } else {
      await this.redis.del(pointsKey)
      await this.redis.del(penaltyKey)
    }
  }

  async getActionLoadPenalties(runnerIds: string[]): Promise<Map<string, number>> {
    if (runnerIds.length === 0) {
      return new Map()
    }

    const keys = runnerIds.map((id) => this.getActionLoadPenaltyRedisKey(id))
    const values = await this.redis.mget(keys)

    const penalties = new Map<string, number>()

    runnerIds.forEach((id, index) => {
      const penalty = parseInt(values[index] || '0', 10)
      penalties.set(id, penalty)
    })

    return penalties
  }
}
