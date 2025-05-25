/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ArgumentsHost, Catch, ExceptionFilter, HttpException, HttpStatus, NotFoundException } from '@nestjs/common'
import { join } from 'path'

@Catch(NotFoundException)
export class NotFoundExceptionFilter implements ExceptionFilter {
  catch(exception: unknown, host: ArgumentsHost) {
    const ctx = host.switchToHttp()
    const request = ctx.getRequest()

    if (!request.path.startsWith('/api/')) {
      const response = ctx.getResponse()
      response.sendFile(join(__dirname, '..', 'dashboard', 'index.html'))
      return
    }

    // TODO: refactor (this is duplicate code from /src/filters/all-exceptions.filter.ts)
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
      path: request.path,
      error,
      message,
    }
    ctx.getResponse().status(statusCode).json(responseBody)
    return
  }

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
}
