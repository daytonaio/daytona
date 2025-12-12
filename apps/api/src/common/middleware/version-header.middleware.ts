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

  use(req: Request, res: Response, next: NextFunction) {
    const version = this.configService.get('version')
    if (version) {
      res.setHeader('X-Daytona-Api-Version', `${version}`)
    }
    next()
  }
}
