/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Logger } from '@nestjs/common'
import { createProxyMiddleware, fixRequestBody, Options } from 'http-proxy-middleware'
import { IncomingMessage, ServerResponse } from 'http'
import { NextFunction } from 'express'

export class LogProxy {
  private readonly logger = new Logger(LogProxy.name)

  constructor(
    private readonly targetUrl: string,
    private readonly imageRef: string,
    private readonly authToken: string,
    private readonly req: IncomingMessage,
    private readonly res: ServerResponse<IncomingMessage>,
    private readonly next: NextFunction,
  ) {}

  create() {
    const proxyOptions: Options = {
      target: this.targetUrl,
      ws: true,
      secure: false,
      changeOrigin: true,
      autoRewrite: true,
      pathRewrite: () => `/images/logs?imageRef=${this.imageRef}`,
      on: {
        proxyReq: (proxyReq: any, req: any) => {
          proxyReq.setHeader('Authorization', `Bearer ${this.authToken}`)
          fixRequestBody(proxyReq, req)
        },
        proxyReqWs: (proxyReq: any, req: any, socket: any, options: any, head: any) => {
          this.logger.debug('WebSocket connection upgrading')
          proxyReq.setHeader('Authorization', `Bearer ${this.authToken}`)
        },
        proxyRes: (proxyRes: any, req: any, res: any) => {
          Object.keys(proxyRes.headers).forEach((key) => {
            try {
              res.setHeader(key, proxyRes.headers[key])
            } catch (err) {
              this.logger.warn(`Failed to set header ${key}: ${err.message}`)
            }
          })
        },
        open: (proxySocket: any) => {
          this.logger.debug('WebSocket connection opened')

          // Set socket timeout
          proxySocket.setTimeout(60000)

          // Track active connections if needed
          // this.activeConnections.add(proxySocket)

          // Listen for socket-specific events
          proxySocket.on('close', () => {
            this.logger.debug('WebSocket connection closed')
            // this.activeConnections.delete(proxySocket)
          })
        },
        error: (err: Error, req: any, res: any) => {
          this.logger.error(`Proxy error: ${err.message}`, err.stack)

          // Check if res is a valid HTTP response object before using writeHead
          if (!res.headersSent && typeof res.writeHead === 'function') {
            try {
              res.writeHead(500, { 'Content-Type': 'text/plain' })
              res.end(`Proxy error: ${err.message}`)
            } catch (writeErr) {
              this.logger.error(`Failed to write error response: ${writeErr.message}`)
            }
          } else {
            this.logger.error(`Proxy WebSocket error: ${err.message}`)
          }
        },
      },
      proxyTimeout: 60 * 1000,
      timeout: 60 * 1000,
    }

    try {
      const proxy = createProxyMiddleware(proxyOptions)
      return proxy(this.req, this.res, this.next)
    } catch (error) {
      this.logger.error(`Failed to create proxy: ${error.message}`, error.stack)
      if (!this.res.headersSent) {
        try {
          this.res.writeHead(500, { 'Content-Type': 'text/plain' })
          this.res.end(`Failed to create proxy: ${error.message}`)
        } catch (writeErr) {
          this.logger.error(`Failed to write error response: ${writeErr.message}`)
        }
      }
    }
  }
}
