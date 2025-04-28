/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NestMiddleware, Logger } from '@nestjs/common'
import { Request, Response, NextFunction } from 'express'

@Injectable()
export class RequestLoggerMiddleware implements NestMiddleware {
  private readonly logger = new Logger('HTTP')

  use(req: Request, res: Response, next: NextFunction) {
    const { method, originalUrl, headers, body } = req

    this.logger.debug(`${method} ${originalUrl}`, {
      headers,
      body,
      contentType: headers['content-type'],
      contentLength: headers['content-length'],
    })

    next()
  }
}
