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
  UnauthorizedException,
} from '@nestjs/common'
import { ThrottlerException } from '@nestjs/throttler'
import { FailedAuthTrackerService } from '../auth/failed-auth-tracker.service'

const secretPatterns = [
  /(api[_-]?key|access[_-]?key|secret[_-]?key|secret|token|password|passwd|authorization)[:=]\s*([^\s,;]+)/gi,
  /(bearer\s+)[A-Za-z0-9._\-~+/=]+/gi,
  /(AKIA[0-9A-Z]{16})/g,
  /(gh[pousr]_[A-Za-z0-9]{20,})/g,
  /(sk_live_[A-Za-z0-9]{16,})/g,
]

@Catch()
export class AllExceptionsFilter implements ExceptionFilter {
  private readonly logger = new Logger(AllExceptionsFilter.name)

  constructor(private readonly failedAuthTracker: FailedAuthTrackerService) {}

  async catch(exception: unknown, host: ArgumentsHost): Promise<void> {
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

    // Track failed authentication attempts
    if (exception instanceof UnauthorizedException) {
      try {
        await this.failedAuthTracker.incrementFailedAuth(request, response)
      } catch (trackingError) {
        if (trackingError instanceof ThrottlerException) {
          exception = trackingError
        } else {
          this.logger.error('Failed to track authentication failure:', trackingError)
        }
      }
    }

    if (exception instanceof HttpException) {
      statusCode = exception.getStatus()
      error = STATUS_CODES[statusCode]
      const exceptionResponse = exception.getResponse()
      if (typeof exceptionResponse === 'string') {
        message = exceptionResponse
      } else {
        const responseMessage = (exceptionResponse as Record<string, unknown>).message
        message = Array.isArray(responseMessage)
          ? responseMessage.join(', ')
          : (responseMessage as string) || exception.message
      }
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
      message: redactString(message),
    })
  }
}

function redactString(input: string): string {
  let redacted = input
  for (const pattern of secretPatterns) {
    redacted = redacted.replace(pattern, '$1[REDACTED]')
  }
  return redacted
}
