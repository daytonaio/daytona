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

  // Per-resize sidecar: stamps this resize's contribution to the runner hash so the
  // release path can look it up by sandboxId without needing the sandbox row.
  private getResizeReservationKey(runnerId: string, sandboxId: string): string {
    return `runner:${runnerId}:resize:${sandboxId}`
  }

  // Atomic: bump the per-runner hash and write the matching sidecar in one EVAL.
  async reservePendingRunnerUsageForResize(
    runnerId: string,
    sandboxId: string,
    cpu: number,
    memory: number,
    disk: number,
  ): Promise<void> {
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

      redis.call("HSET", KEYS[2], "cpu", ARGV[1], "memory", ARGV[2], "disk", ARGV[3])
      redis.call("EXPIRE", KEYS[2], ARGV[4])
    `
    await this.redis.eval(
      script,
      2,
      this.getPendingUsageKey(runnerId),
      this.getResizeReservationKey(runnerId, sandboxId),
      cpu.toString(),
      memory.toString(),
      disk.toString(),
      this.PENDING_USAGE_TTL_S.toString(),
    )
  }

  // Atomic: read sidecar → decrement runner hash by it → delete sidecar. Missing sidecar
  // is a no-op, so double-release is safe. Swallows Redis errors; TTL is the backstop.
  // HINCRBY rejects `-0` / float-stringified deltas — skip zero-deltas and format the
  // negated integer explicitly so the script stays safe across field combinations.
  async safeReleasePendingRunnerUsageForResize(runnerId: string, sandboxId: string): Promise<void> {
    const script = `
      local reserved = redis.call("HMGET", KEYS[2], "cpu", "memory", "disk")
      if not reserved[1] and not reserved[2] and not reserved[3] then
        return 0
      end
      local function adjust(field, delta_str)
        local d = tonumber(delta_str)
        if d == nil or d == 0 then return end
        local value = redis.call("HINCRBY", KEYS[1], field, string.format("%d", -d))
        if value < 0 then
          redis.call("HSET", KEYS[1], field, 0)
        end
      end
      adjust("cpu",    reserved[1])
      adjust("memory", reserved[2])
      adjust("disk",   reserved[3])
      redis.call("EXPIRE", KEYS[1], ARGV[1])
      redis.call("DEL", KEYS[2])
      return 1
    `
    try {
      await this.redis.eval(
        script,
        2,
        this.getPendingUsageKey(runnerId),
        this.getResizeReservationKey(runnerId, sandboxId),
        this.PENDING_USAGE_TTL_S.toString(),
      )
    } catch (e) {
      this.logger.warn(`Failed to release pending runner usage for ${runnerId}/${sandboxId}: ${e}`)
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
