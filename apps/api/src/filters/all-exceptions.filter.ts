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
import { ApiErrorCode, HTTP_STATUS_TO_API_CODE } from '../common/errors/api-error-code.enum'

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
    // Approach A: code resolved from status; Approach B: code taken from exception body
    let code: string

    // Serve dashboard SPA for non-API 404s (preserve existing behaviour)
    if (exception instanceof NotFoundException && !request.path.startsWith('/api/')) {
      const res = ctx.getResponse()
      res.sendFile(join(__dirname, '..', 'dashboard', 'index.html'))
      return
    }

    // Track failed authentication attempts (preserve existing behaviour)
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
        // Approach A: derive code from HTTP status
        code = HTTP_STATUS_TO_API_CODE[statusCode] ?? ApiErrorCode.INTERNAL_SERVER_ERROR
      } else {
        const body = exceptionResponse as Record<string, unknown>
        const responseMessage = body.message
        message = Array.isArray(responseMessage)
          ? responseMessage.join(', ')
          : (responseMessage as string) || exception.message

        // Approach B detection: controller explicitly set a typed code → honour it.
        // Approach A fallback: derive from HTTP status.
        code = typeof body.code === 'string' ? body.code : (HTTP_STATUS_TO_API_CODE[statusCode] ?? ApiErrorCode.INTERNAL_SERVER_ERROR)
      }
    } else {
      this.logger.error(exception)
      error = STATUS_CODES[HttpStatus.INTERNAL_SERVER_ERROR]
      message = 'An unexpected error occurred.'
      statusCode = HttpStatus.INTERNAL_SERVER_ERROR
      code = ApiErrorCode.INTERNAL_SERVER_ERROR
    }

    response.status(statusCode).json({
      statusCode,
      source: 'DAYTONA_API',
      code,
      message,
      // Legacy fields kept for backward compat
      error,
      path: request.url,
      timestamp: new Date().toISOString(),
      method: request.method,
    })
  }
}
