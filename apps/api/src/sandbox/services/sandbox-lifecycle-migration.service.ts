/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'

/**
 * Boot-time configuration for the sandbox table split.
 *
 * The wide `sandbox` table is being split into a config-only `sandbox`
 * plus a hot, LIST-partitioned `sandbox_lifecycle` that holds the
 * state-machine columns. The migration is gated by two independent env
 * flags so the rollout can proceed incrementally and roll back safely.
 *
 * - `SANDBOX_LIFECYCLE_READ_FROM_NEW_TABLE` — when `true`, read paths
 *   JOIN `sandbox_lifecycle` and overlay state-machine columns from it
 *   onto the returned entity. The trigger keeps the two tables in sync
 *   while writes still flow through `sandbox`, so this flag is safe to
 *   flip on as soon as the trigger is installed and the backfill has
 *   completed.
 *
 * - `SANDBOX_LIFECYCLE_WRITE_TO_NEW_TABLE` — when `true`, write paths
 *   route state-machine columns to `sandbox_lifecycle` directly and
 *   stop touching those columns on `sandbox`. After this flips, the
 *   legacy columns on `sandbox` freeze in place and `sandbox_lifecycle`
 *   becomes the only source of truth. Reads MUST already be served
 *   from `sandbox_lifecycle` (the constructor enforces this — otherwise
 *   the app would write to a table the readers can't see).
 *
 * Valid combinations:
 *
 * | READ_FROM_NEW_TABLE | WRITE_TO_NEW_TABLE | Source of truth                  |
 * |---------------------|--------------------|----------------------------------|
 * | false               | false              | `sandbox` (baseline)             |
 * | true                | false              | `sandbox` (trigger mirrors)       |
 * | true                | true               | `sandbox_lifecycle`               |
 *
 * The combination `READ_FROM_NEW_TABLE=false, WRITE_TO_NEW_TABLE=true` is
 * rejected at boot — it would silently corrupt state visibility (writes
 * targeting a table the readers can't see).
 *
 * This service is intentionally distinct from any product-level feature
 * flag system. It is a boot-time, env-driven, process-wide switch for a
 * one-time infrastructure migration and will be deleted once the follow-up
 * cleanup PR has dropped the legacy columns from `sandbox`.
 */
@Injectable()
export class SandboxLifecycleMigrationService {
  private readonly logger = new Logger(SandboxLifecycleMigrationService.name)

  private readonly readFromLifecycle: boolean
  private readonly writeToLifecycle: boolean

  constructor() {
    this.readFromLifecycle = SandboxLifecycleMigrationService.parseEnvBool(
      process.env.SANDBOX_LIFECYCLE_READ_FROM_NEW_TABLE,
    )
    this.writeToLifecycle = SandboxLifecycleMigrationService.parseEnvBool(
      process.env.SANDBOX_LIFECYCLE_WRITE_TO_NEW_TABLE,
    )

    if (this.writeToLifecycle && !this.readFromLifecycle) {
      throw new Error(
        'Invalid sandbox lifecycle migration configuration: ' +
          'SANDBOX_LIFECYCLE_WRITE_TO_NEW_TABLE=true requires ' +
          'SANDBOX_LIFECYCLE_READ_FROM_NEW_TABLE=true. ' +
          'Writing state-machine columns to sandbox_lifecycle while reading them ' +
          'from sandbox would mask all state transitions from the application.',
      )
    }

    this.logger.log(
      `Sandbox lifecycle migration: READ_FROM_NEW_TABLE=${this.readFromLifecycle}, ` +
        `WRITE_TO_NEW_TABLE=${this.writeToLifecycle}`,
    )
  }

  /**
   * `true` iff the new `sandbox_lifecycle` table is the source of truth
   * for reads of state-machine columns. When `true`, repository
   * find/findOne overrides auto-JOIN `sandbox_lifecycle` and hydrate
   * state-machine columns from it; QueryBuilder call sites use
   * `lifecycle.state` instead of `sandbox.state` for WHERE/SELECT/ORDER BY.
   *
   * When `false`, the legacy `sandbox` table is the read source.
   */
  useLifecycleTableForReads(): boolean {
    return this.readFromLifecycle
  }

  /**
   * `true` iff the new `sandbox_lifecycle` table is the write target
   * for state-machine columns. When `true`, the repository routes
   * state-machine writes directly to `sandbox_lifecycle` and stops
   * touching the corresponding columns on `sandbox`.
   *
   * When `false`, writes go to `sandbox` and the trigger propagates
   * them into `sandbox_lifecycle`.
   */
  useLifecycleTableForWrites(): boolean {
    return this.writeToLifecycle
  }

  private static parseEnvBool(value: string | undefined): boolean {
    if (!value) return false
    const normalized = value.trim().toLowerCase()
    return normalized === 'true' || normalized === '1' || normalized === 'yes'
  }
}
