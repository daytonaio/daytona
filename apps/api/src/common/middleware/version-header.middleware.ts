/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NestMiddleware } from '@nestjs/common'
import { Request, Response, NextFunction } from 'express'
import { TypedConfigService } from '../../config/typed-config.service'

@Injectable()
export class VersionHeaderMiddleware implements NestMiddleware {
  private readonly version: string | undefined

  constructor(private readonly configService: TypedConfigService) {
    this.version = this.configService.get('version')
  }

  use(req: Request, res: Response, next: NextFunction) {
    if (this.version) {
      res.setHeader('X-Daytona-Api-Version', `${this.version}`)
    }
    next()
  }
}
