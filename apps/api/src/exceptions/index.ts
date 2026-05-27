/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

// Coded exceptions: each one carries an `ApiErrorCode` so callers can
// distinguish the condition beyond its HTTP status. For plain status-only
// responses (400/404/etc.), throw NestJS's built-in `BadRequestException`,
// `NotFoundException`, etc. directly.
export { SandboxStateError } from './sandbox-state-error.exception'
export { SandboxBackupStateError } from './sandbox-backup-state-error.exception'
export { SandboxOperationNotSupportedError } from './sandbox-operation-not-supported.exception'
export { SandboxDiskExpansionLimitError } from './sandbox-disk-expansion-limit.exception'
export { StateChangeInProgressError } from './state-change-in-progress.exception'
export { SnapshotStateChangeInProgressError } from './snapshot-state-change-in-progress.exception'
export { OrganizationQuotaExceededError } from './organization-quota-exceeded.exception'
export { OrganizationSuspendedError } from './organization-suspended.exception'
export { ApiKeyExpiredError } from './api-key-expired.exception'
export { VolumeInUseError } from './volume-in-use.exception'
export { NoAvailableRunnersError } from './no-available-runners.exception'
export { DefaultRegionRequiredError } from './default-region-required.exception'
export { SandboxRunnerNotFoundError } from './sandbox-runner-not-found.exception'
