/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logger } from '@nestjs/common'
import { IncomingMessage, ServerResponse } from 'http'
import { NextFunction } from 'express'
import { RunnerClient } from '@daytonaio/runner-grpc-client'

export class LogProxy {
  private readonly logger = new Logger(LogProxy.name)

  constructor(
    private readonly runnerClient: RunnerClient, // Changed from targetUrl to runnerClient
    private readonly imageRef: string,
    private readonly authToken: string, // Keep for compatibility, but not used in gRPC (handled by client factory)
    private readonly follow: boolean,
    private readonly req: IncomingMessage,
    private readonly res: ServerResponse<IncomingMessage>,
    private readonly next: NextFunction,
  ) {}

  async create() {
    try {
      // Set response headers
      this.res.setHeader('Content-Type', 'application/octet-stream')

      // Create the gRPC request
      const request = {
        image_ref: this.imageRef,
        follow: this.follow,
      }

      // Get the stream from gRPC client
      const stream = await this.runnerClient.buildLogs(request)

      // Handle AsyncIterable stream
      if (stream && typeof stream[Symbol.asyncIterator] === 'function') {
        try {
          for await (const response of stream) {
            if (response && response.data) {
              this.res.write(response.data)
            }
          }
          this.res.end()
        } catch (streamError) {
          this.logger.error(`Error streaming logs: ${streamError.message}`)
          if (!this.res.headersSent) {
            this.res.statusCode = 500
          }
          this.res.end()
        }
      } else {
        // Fallback for single response (non-streaming case)
        if (stream && (stream as any).data) {
          this.res.write((stream as any).data)
        }
        this.res.end()
      }
    } catch (error) {
      this.logger.error(`Error in LogProxy: ${error.message}`)
      if (!this.res.headersSent) {
        this.res.statusCode = 500
      }
      this.res.end()
    }
  }
}
