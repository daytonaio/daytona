/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { join } from 'node:path'
import { STATUS_CODES } from 'node:http'
import { Request, Response } from 'express'
import {
  ExceptionFilter,
  Catch,
  ArgumentsHost,
  HttpException,
  Logger,
  HttpStatus,
  NotFoundException,
} from '@nestjs/common'

@Catch()
export class AllExceptionsFilter implements ExceptionFilter {
  private readonly logger = new Logger(AllExceptionsFilter.name)

  catch(exception: unknown, host: ArgumentsHost): void {
    const ctx = host.switchToHttp()
    const response = ctx.getResponse<Response>()
    const request = ctx.getRequest<Request>()

    let statusCode: number
    let error: string
    let message: string

    // If the exception is a NotFoundException and the request path is not an API request, serve the dashboard index.html file
    if (exception instanceof NotFoundException && !request.path.startsWith('/api/')) {
      const response = ctx.getResponse()
      response.sendFile(join(__dirname, '..', 'dashboard', 'index.html'))
      return
    }

    if (exception instanceof HttpException) {
      statusCode = exception.getStatus()
      error = STATUS_CODES[statusCode]
      message = exception.message
    } else {
      this.logger.error(exception)
      error = STATUS_CODES[HttpStatus.INTERNAL_SERVER_ERROR]
      message = 'An unexpected error occurred.'
      statusCode = HttpStatus.INTERNAL_SERVER_ERROR
    }

    response.status(statusCode).json({
      path: request.url,
      timestamp: new Date().toISOString(),
      statusCode,
      error,
      message,
    })
  }
}
