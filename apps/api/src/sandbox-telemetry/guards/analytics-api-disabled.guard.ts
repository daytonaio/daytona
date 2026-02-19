/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { CanActivate, Injectable, ForbiddenException } from '@nestjs/common'
import { TypedConfigService } from '../../config/typed-config.service'

@Injectable()
export class AnalyticsApiDisabledGuard implements CanActivate {
  constructor(private readonly configService: TypedConfigService) {}

  canActivate(): boolean {
    if (this.configService.get('analyticsApiUrl')) {
      throw new ForbiddenException('Telemetry endpoints are disabled when Analytics API is configured')
    }
    return true
  }
}
