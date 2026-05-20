/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { InjectRedis } from '@nestjs-modules/ioredis'
import Redis from 'ioredis'

/**
 * Owns the per-runner "pending usage" Redis hash — the in-flight resource reservations
 * (cpu/memory/disk) used to throttle on-runner work (e.g. resizes) before the next
 * healthcheck reflects them. The runner analog of OrganizationUsageService's pending usage.
 *
 * Extracted from RunnerService so JobStateHandlerService can release a slot without
 * importing RunnerService, which would close a runner.service <-> job.service import cycle.
 * This service deliberately depends on nothing but Redis.
 */
@Injectable()
export class RunnerUsageService {
  private readonly logger = new Logger(RunnerUsageService.name)

  // Safety-net TTL only; the normal path releases each slot explicitly on completion.
  private readonly PENDING_USAGE_TTL_S = 2 * 60

  constructor(@InjectRedis() private readonly redis: Redis) {}

  private getPendingUsageKey(runnerId: string): string {
    return `runner:${runnerId}:usage:pending`
  }

  /**
   * Atomically adjusts the per-runner pending usage hash and refreshes its TTL.
   * Signed deltas; each field is clamped at 0 so a double-release can't go negative.
   */
  async incrementPendingRunnerUsage(runnerId: string, cpu: number, memory: number, disk: number): Promise<void> {
    const script = `
      local function adjust(field, delta)
        local value = redis.call("HINCRBY", KEYS[1], field, delta)
        if value < 0 then
          redis.call("HSET", KEYS[1], field, 0)
        end
      end
      adjust("cpu",    ARGV[1])
      adjust("memory", ARGV[2])
      adjust("disk",   ARGV[3])
      redis.call("EXPIRE", KEYS[1], ARGV[4])
    `
    await this.redis.eval(
      script,
      1,
      this.getPendingUsageKey(runnerId),
      cpu.toString(),
      memory.toString(),
      disk.toString(),
      this.PENDING_USAGE_TTL_S.toString(),
    )
  }

  /**
   * Releases a pending usage slot. Symmetric with incrementPendingRunnerUsage — passes
   * negatives through the same Lua so behavior (incl. TTL refresh) stays identical.
   */
  async decrementPendingRunnerUsage(runnerId: string, cpu: number, memory: number, disk: number): Promise<void> {
    return this.incrementPendingRunnerUsage(runnerId, -cpu, -memory, -disk)
  }

  /**
   * decrementPendingRunnerUsage that swallows + warns on failure, so a Redis blip in a
   * terminal hook never shadows the original error; the TTL is the safety net.
   */
  async safeDecrementPendingRunnerUsage(runnerId: string, cpu: number, memory: number, disk: number): Promise<void> {
    try {
      await this.decrementPendingRunnerUsage(runnerId, cpu, memory, disk)
    } catch (e) {
      this.logger.warn(`Failed to decrement pending runner usage for ${runnerId}: ${e}`)
    }
  }

  /**
   * Reads the current pending usage counters (cpu/memory/disk) for a runner.
   * Missing fields read as 0.
   */
  async getPendingRunnerUsage(runnerId: string): Promise<{ cpu: number; memory: number; disk: number }> {
    const [cpu, memory, disk] = await this.redis.hmget(this.getPendingUsageKey(runnerId), 'cpu', 'memory', 'disk')
    return {
      cpu: Number(cpu ?? 0),
      memory: Number(memory ?? 0),
      disk: Number(disk ?? 0),
    }
  }
}
