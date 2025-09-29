/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger, UnauthorizedException } from '@nestjs/common'
import { PassportStrategy } from '@nestjs/passport'
import { Strategy } from 'passport-http-bearer'
import { SandboxService } from '../sandbox/services/sandbox.service'
import { DaemonContext } from '../common/interfaces/daemon-context.interface'

@Injectable()
export class DaemonStrategy extends PassportStrategy(Strategy, 'daemon') {
  private readonly logger = new Logger(DaemonStrategy.name)

  constructor(private readonly sandboxService: SandboxService) {
    super()
  }

  async validate(token: string): Promise<DaemonContext> {
    this.logger.debug('Validating daemon token')

    try {
      // Find sandbox by auth token
      const sandbox = await this.sandboxService.findByAuthToken(token)
      if (!sandbox) {
        throw new UnauthorizedException('Invalid daemon token')
      }

      return {
        role: 'daemon',
        sandboxId: sandbox.id,
      }
    } catch (error) {
      this.logger.debug('Error validating daemon token:', error)
      throw new UnauthorizedException('Invalid daemon token')
    }
  }
}
