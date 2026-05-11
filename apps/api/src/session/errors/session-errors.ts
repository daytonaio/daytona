/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { HttpException, HttpStatus } from '@nestjs/common'

/**
 * SessionInvalidatedError is thrown when a caller passes a sessionId whose underlying
 * SessionInstance was rolled (sandbox died, snapshot drift, autostop). Surfaces as HTTP 410
 * with `error.name = 'SessionInvalidated'`. The signal lets SDK auto-retry / dashboards
 * distinguish "your sandbox died" from "you idled out".
 */
export class SessionInvalidatedError extends HttpException {
  constructor(sessionId: string, invalidatedAt: Date | string) {
    super(
      {
        statusCode: HttpStatus.GONE,
        error: {
          name: 'SessionInvalidated',
          sessionId,
          invalidatedAt: typeof invalidatedAt === 'string' ? invalidatedAt : invalidatedAt.toISOString(),
        },
        message: `Session ${sessionId} has been invalidated.`,
      },
      HttpStatus.GONE,
    )
  }
}

/**
 * SessionExpiredError is thrown when a caller passes a sessionId whose state is EXPIRED
 * (idle TTL or absolute TTL hit). Body distinguishes idle vs absolute so dashboards can
 * surface a meaningful reason. Same HTTP 410 status as SessionInvalidatedError so a generic
 * `if (err.status === 410) reset` handler covers both.
 */
export class SessionExpiredError extends HttpException {
  constructor(sessionId: string, expiredAt: Date | string, reason: 'idle' | 'absolute') {
    super(
      {
        statusCode: HttpStatus.GONE,
        error: {
          name: 'SessionExpired',
          sessionId,
          expiredAt: typeof expiredAt === 'string' ? expiredAt : expiredAt.toISOString(),
          reason,
        },
        message: `Session ${sessionId} has expired (${reason}).`,
      },
      HttpStatus.GONE,
    )
  }
}
