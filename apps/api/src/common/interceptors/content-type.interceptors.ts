/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, NestInterceptor, ExecutionContext, CallHandler, Logger } from '@nestjs/common'
import { Observable } from 'rxjs'

@Injectable()
export class ContentTypeInterceptor implements NestInterceptor {
  private readonly logger = new Logger(ContentTypeInterceptor.name)

  async intercept(context: ExecutionContext, next: CallHandler): Promise<Observable<any>> {
    const request = context.switchToHttp().getRequest()

    // Check if we have raw body data but no parsed body
    if (request.readable) {
      // Create a promise to handle the body parsing
      await new Promise<void>((resolve, reject) => {
        let rawBody = ''

        // Collect the raw body data
        request.on('data', (chunk: Buffer) => {
          rawBody += chunk.toString()
        })

        // Once we have all the data, try to parse it as JSON
        request.on('end', () => {
          try {
            if (rawBody) {
              request.body = JSON.parse(rawBody)
              request.headers['content-type'] = 'application/json'
            }
            resolve()
          } catch (e) {
            this.logger.error('Failed to parse JSON body:', e)
            resolve() // Still resolve even on error to prevent hanging
          }
        })

        // Handle potential errors
        request.on('error', (error) => {
          this.logger.error('Error reading request body:', error)
          reject(error)
        })
      })
    }

    // Add Content-Type header if it's missing and there's a request body
    if (request.body && Object.keys(request.body).length > 0 && !request.get('content-type')) {
      request.headers['content-type'] = 'application/json'
    }

    return next.handle()
  }
}
