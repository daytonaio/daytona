/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import type { Sandbox } from '../entities/sandbox.entity'
import type { SandboxLifecycle } from '../entities/sandbox-lifecycle.entity'

/**
 * Column membership for the sandbox table split.
 *
 * The original wide `sandbox` table is split by write cadence:
 * - `sandbox` (config, unpartitioned) — rare writes after creation. Bulk of
 *   the JSONB lives here and gets TOAST'd.
 * - `sandbox_lifecycle` (hot, LIST-partitioned by `lifecyclePhase`) — every
 *   state-machine UPDATE and every backup-poll UPDATE hits this table.
 *
 * These arrays are the single source of truth for routing writes when
 * `SANDBOX_LIFECYCLE_WRITE_TO_NEW_TABLE` is enabled. The repository's
 * `update()` / `updateWhere()` / `insert()` methods split `updateData` by
 * column membership and route each part to the correct table.
 *
 * IMPORTANT: `organizationId` appears in both arrays because it is
 * denormalized to `sandbox_lifecycle` so the optimistic-concurrency
 * predicate (which checks `organizationId`) can stay single-table. Writes
 * that change `organizationId` (warm-pool assignment) must update both
 * tables atomically — the repository handles this via the
 * `SANDBOX_DUAL_WRITE_COLUMNS` list below.
 */

/**
 * Columns that live on the `sandbox` (config) table. Written rarely after
 * creation. Includes `organizationId` which is also denormalized to
 * `sandbox_lifecycle` for optimistic-concurrency predicates.
 */
export const SANDBOX_CONFIG_COLUMNS = [
  'id',
  'organizationId',
  'name',
  'region',
  'sandboxClass',
  'snapshot',
  'osUser',
  'env',
  'labels',
  'volumes',
  'public',
  'networkBlockAll',
  'networkAllowList',
  'cpu',
  'gpu',
  'mem',
  'disk',
  'authToken',
  'autoStopInterval',
  'autoArchiveInterval',
  'autoDeleteInterval',
  'buildInfo',
  'createdAt',
] as const satisfies ReadonlyArray<keyof Sandbox>

/**
 * Columns that live on the `sandbox_lifecycle` (hot) table. Every
 * state-machine transition and every backup poll writes one or more of
 * these. `organizationId` is denormalized from `sandbox` so the
 * optimistic-concurrency predicate stays single-table.
 *
 * `lifecyclePhase` is the partition key, derived from `state` by
 * `Sandbox.enforceInvariants()` — callers never write it directly.
 */
export const SANDBOX_LIFECYCLE_COLUMNS = [
  'sandboxId',
  'lifecyclePhase',
  'organizationId',
  'state',
  'desiredState',
  'pending',
  'errorReason',
  'recoverable',
  'daemonVersion',
  'runnerId',
  'prevRunnerId',
  'backupState',
  'lastBackupAt',
  'backupSnapshot',
  'backupRegistryId',
  'backupErrorReason',
  'existingBackupSnapshots',
  'updatedAt',
] as const satisfies ReadonlyArray<keyof SandboxLifecycle>

/**
 * Columns whose updates must propagate to BOTH tables when
 * `SANDBOX_LIFECYCLE_WRITE_TO_NEW_TABLE` is enabled.
 *
 * Currently only `organizationId` — denormalized to support multi-tenant
 * race detection in warm-pool assignment without crossing tables in the
 * optimistic predicate.
 */
export const SANDBOX_DUAL_WRITE_COLUMNS: readonly (keyof Sandbox & keyof SandboxLifecycle)[] = ['organizationId']

/**
 * Columns that exist on the `Sandbox` entity today but logically belong to
 * `sandbox_lifecycle`. Once the app has switched to writing
 * `sandbox_lifecycle` directly, these columns on `sandbox` stop receiving
 * fresh writes and become stale. The repository's read shim uses this list
 * to overlay fresh `sandbox_lifecycle` values onto the in-memory entity so
 * caller code keeps reading `sandbox.state` etc. transparently.
 *
 * These columns will be dropped from `sandbox` entirely in a follow-up
 * cleanup PR (deferred from this PR for rollback safety — keeping the
 * columns around lets us un-flip the WRITE flag and resync to `sandbox`
 * if anything goes wrong).
 */
export const SANDBOX_INERT_AFTER_CUTOVER = [
  'state',
  'desiredState',
  'pending',
  'errorReason',
  'recoverable',
  'daemonVersion',
  'runnerId',
  'prevRunnerId',
  'backupState',
  'lastBackupAt',
  'backupSnapshot',
  'backupRegistryId',
  'backupErrorReason',
  'existingBackupSnapshots',
  'updatedAt',
] as const satisfies ReadonlyArray<keyof Sandbox>
