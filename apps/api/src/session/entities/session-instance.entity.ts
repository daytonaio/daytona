/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SessionInstanceState } from '../enums/session-instance-state.enum'
import { SessionInstanceRole } from '../enums/session-instance-role.enum'

/**
 * SessionInstance represents a single warm sandbox backing one (organization, template) pair.
 * The pool service owns lifecycle: PROVISIONING → READY → (rolled on snapshot drift, sandbox
 * death, or autostop).
 *
 * Storage: Redis is the source of truth (see SessionInstanceStore). This class is a plain shape
 * — NOT a TypeORM entity. Because Redis may be wiped, the pool runs an orphan-sandbox reconciler
 * that destroys any session sandbox whose instance no longer exists in Redis.
 *
 * Scale-out: there can be MANY instances per (organizationId, templateId): a `warm` floor plus
 * `overflow` instances the autoscaler adds under load and reaps first when idle. `lastActiveAt`
 * tracks when the instance last served a request, driving scale-in.
 *
 * `snapshotId` is denormalized from the template at instance-create time so the pool reconciler
 * can detect drift (instance.snapshotId != template.snapshotId) without an extra join.
 */
export class SessionInstance {
  id: string

  organizationId: string

  templateId: string

  snapshotId: string

  sandboxId?: string

  state: SessionInstanceState = SessionInstanceState.PROVISIONING

  errorReason?: string

  role: SessionInstanceRole = SessionInstanceRole.WARM

  lastUsedAt?: Date

  /**
   * When this instance last served (or was selected to serve) a request. Distinct from
   * `lastUsedAt`: the scheduler stamps this on every pick so scale-in only reaps `overflow`
   * instances that have been idle long enough.
   */
  lastActiveAt?: Date

  createdAt: Date

  updatedAt: Date
}
