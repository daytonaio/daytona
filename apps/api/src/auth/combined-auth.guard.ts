/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, ExecutionContext, Logger } from '@nestjs/common'
import { AuthGuard } from '@nestjs/passport'

@Injectable()
export class CombinedAuthGuard extends AuthGuard(['jwt', 'api-key']) {
  private readonly logger = new Logger(CombinedAuthGuard.name)

  constructor() {
    super()
    this.logger.debug('CombinedAuthGuard constructor called')
  }

  async canActivate(context: ExecutionContext): Promise<boolean> {
    this.logger.debug('CombinedAuthGuard.canActivate called')
    try {
      const result = await super.canActivate(context)
      this.logger.debug('Authentication result:', result)
      return result as boolean
    } catch (error) {
      this.logger.debug('Authentication error:', error)
      throw error
    }
  }

  handleRequest(err: any, user: any, info: any, context: ExecutionContext) {
    this.logger.debug('CombinedAuthGuard.handleRequest called')
    this.logger.debug('Error:', err)
    this.logger.debug('User:', user)
    this.logger.debug('Info:', info)

    if (err || !user) {
      this.logger.debug('Authentication failed')
      return super.handleRequest(err, user, info, context)
    }

    this.logger.debug('Authentication successful')
    return user
  }
}
