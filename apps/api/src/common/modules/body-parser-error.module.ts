/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Module, OnModuleInit, BadRequestException } from '@nestjs/common'
import { HttpAdapterHost } from '@nestjs/core'
import { Request, Response, NextFunction } from 'express'
@Module({})
export class BodyParserErrorModule implements OnModuleInit {
  constructor(private readonly httpAdapterHost: HttpAdapterHost) {}

  onModuleInit() {
    const app = this.httpAdapterHost.httpAdapter.getInstance()

    app.use((err: Error & { body?: unknown }, req: Request, res: Response, next: NextFunction) => {
      if (err instanceof SyntaxError && 'body' in err) {
        const response = new BadRequestException('Invalid JSON in request body').getResponse()
        return res.status(400).json(response)
      }

      next(err)
    })
  }
}
