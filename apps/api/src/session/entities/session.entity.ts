/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SessionState } from '../enums/session-state.enum'

/**
 * Session owns context identity for the API. Its `id` is the user-facing identifier AND
 * the id passed verbatim to the in-sandbox session-daemon on `POST /sessions`.
 *
 * Storage: Redis is the source of truth (see SessionRepository). This class is a plain shape
 * — NOT a TypeORM entity — so the same type flows through the service layer as it did when the
 * data lived in Postgres, without any table being created or migrated.
 *
 * `lastUsedAt` is bumped on every successful resolve and feeds the idle-TTL GC sweep. The pool
 * reconciler bulk-marks contexts INVALID when an instance is rolled; the GC marks them EXPIRED on
 * idle/absolute TTL.
 */
export class Session {
  id: string

  organizationId: string

  instanceId: string

  language: string

  cwd?: string

  state: SessionState = SessionState.ACTIVE

  invalidatedAt?: Date

  expiredAt?: Date

  createdAt: Date

  lastUsedAt: Date
}
