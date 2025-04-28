/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NestMiddleware } from '@nestjs/common'
import { Request, Response, NextFunction } from 'express'
// import { version } from '../../../package.json'

@Injectable()
export class VersionHeaderMiddleware implements NestMiddleware {
  use(req: Request, res: Response, next: NextFunction) {
    // TODO: Fetch version from package.json
    // res.setHeader('X-Daytona-Api-Version', `v${version}`)
    next()
  }
}
