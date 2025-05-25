/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NestMiddleware, Logger } from '@nestjs/common'
import { Request, Response, NextFunction } from 'express'
import getRawBody from 'raw-body'

@Injectable()
export class RawRequestMiddleware implements NestMiddleware {
  private logger = new Logger('RawRequest')

  async use(req: Request, res: Response, next: NextFunction) {
    if (req.method === 'POST' && req.headers['content-type'] !== 'application/json') {
      const rawBody = await getRawBody(req)
      const bodyStr = rawBody.toString()

      try {
        // Parse JSON and set it directly on req.body
        req.body = JSON.parse(bodyStr)

        // Add content-type header
        req.headers['content-type'] = 'application/json'
      } catch (e) {
        this.logger.error('Failed to parse body:', e)
      }
    }
    next()
  }
}
