/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ExceptionFilter, Catch, ArgumentsHost, HttpException, HttpStatus } from '@nestjs/common'
import { HttpAdapterHost } from '@nestjs/core'

@Catch()
export class AllExceptionsFilter implements ExceptionFilter {
  constructor(private readonly httpAdapterHost: HttpAdapterHost) {}

  private handleCustomError(errorMessage: string): {
    statusCode: number
    errorType: string
  } {
    switch (errorMessage) {
      case 'Sandbox not found':
        return {
          statusCode: HttpStatus.NOT_FOUND,
          errorType: 'Not Found',
        }
      case 'Unauthorized access':
        return {
          statusCode: HttpStatus.UNAUTHORIZED,
          errorType: 'Unauthorized',
        }
      case 'Forbidden operation':
        return {
          statusCode: HttpStatus.FORBIDDEN,
          errorType: 'Forbidden',
        }
      default:
        return {
          statusCode: HttpStatus.INTERNAL_SERVER_ERROR,
          errorType: 'Internal Server Error',
        }
    }
  }

  catch(exception: unknown, host: ArgumentsHost): void {
    const { httpAdapter } = this.httpAdapterHost
    const ctx = host.switchToHttp()

    let statusCode: number
    let message: string
    let error: string

    if (exception instanceof HttpException) {
      statusCode = exception.getStatus()
      const response = exception.getResponse()
      if (typeof response === 'object' && response !== null) {
        message = (response as any).message || exception.message
        error = (response as any).error || 'Http Exception'
      } else {
        message = exception.message
        error = 'Http Exception'
      }
    } else if (exception instanceof Error) {
      const customError = this.handleCustomError(exception.message)
      statusCode = customError.statusCode
      error = customError.errorType
      message = exception.message
    } else {
      statusCode = HttpStatus.INTERNAL_SERVER_ERROR
      message = 'Internal server error'
      error = 'Unknown Error'
    }

    const responseBody = {
      statusCode,
      timestamp: new Date().toISOString(),
      path: httpAdapter.getRequestUrl(ctx.getRequest()),
      error,
      message,
    }

    httpAdapter.reply(ctx.getResponse(), responseBody, statusCode)
  }
}
