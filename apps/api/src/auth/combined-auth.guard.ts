/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, HttpException, Logger, UnauthorizedException } from '@nestjs/common'
import { AuthGuard } from '@nestjs/passport'

/**
 * Main authentication guard for the application.
 *
 * Strategies are tried in array order.
 * On first success, the rest are skipped.
 *
 * `handleRequest` is invoked once — either when a strategy succeeds or when all strategies fail.
 * It returns the authenticated user object or throws a generic `UnauthorizedException`.
 */
@Injectable()
export class CombinedAuthGuard extends AuthGuard(['api-key', 'jwt']) {
  private readonly logger = new Logger(CombinedAuthGuard.name)

  handleRequest(err: any, user: any) {
    if (err instanceof HttpException && err.getStatus() >= 500) {
      throw err
    }

    if (err || !user) {
      this.logger.debug('Authentication failed', { err, user })
      throw new UnauthorizedException('Invalid credentials')
    }

    return user
  }
}
