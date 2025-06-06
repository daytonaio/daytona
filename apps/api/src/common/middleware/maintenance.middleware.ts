/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NestMiddleware, HttpException, HttpStatus } from '@nestjs/common'
import { Request, Response, NextFunction } from 'express'
import { TypedConfigService } from '../../config/typed-config.service'

@Injectable()
export class MaintenanceMiddleware implements NestMiddleware {
  constructor(private readonly configService: TypedConfigService) {}

  use(req: Request, res: Response, next: NextFunction) {
    const isMaintenanceMode = this.configService.get('maintananceMode')

    if (isMaintenanceMode) {
      throw new HttpException(
        {
          statusCode: HttpStatus.SERVICE_UNAVAILABLE,
          message: 'Service is currently under maintenance. Please try again later.',
          error: 'Service Unavailable',
        },
        HttpStatus.SERVICE_UNAVAILABLE,
      )
    }

    next()
  }
}
