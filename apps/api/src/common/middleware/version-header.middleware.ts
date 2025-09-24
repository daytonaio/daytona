/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NestMiddleware } from '@nestjs/common'
import { Request, Response, NextFunction } from 'express'
import { TypedConfigService } from '../../config/typed-config.service'

@Injectable()
export class VersionHeaderMiddleware implements NestMiddleware {
  constructor(private readonly configService: TypedConfigService) {}

  use(_: Request, res: Response, next: NextFunction) {
    res.setHeader('X-Daytona-Api-Version', this.configService.get('version') || 'unknown')
    next()
  }
}
