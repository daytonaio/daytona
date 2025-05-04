/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Injectable, Logger } from '@nestjs/common'
import { createProxyMiddleware, fixRequestBody, Options } from 'http-proxy-middleware'
import { IncomingMessage, ServerResponse } from 'http'
import { NextFunction } from 'express'

@Injectable()
export class ProxyService {
  private readonly logger = new Logger(ProxyService.name)

  createLogProxy(options: {
    targetUrl: string
    imageRef: string
    authToken: string
    req: IncomingMessage
    res: ServerResponse<IncomingMessage>
    next: NextFunction
  }) {
    const { targetUrl, imageRef, authToken, req, res, next } = options

    const proxyOptions: Options = {
      target: targetUrl,
      ws: true,
      secure: false,
      changeOrigin: true,
      autoRewrite: true,
      pathRewrite: () => `/images/logs?imageRef=${imageRef}`,
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
      on: {
        proxyReq: (proxyReq: any, req: any) => {
          proxyReq.setHeader('Authorization', `Bearer ${authToken}`)
          fixRequestBody(proxyReq, req)
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
          const pingInterval = setInterval(() => {
            if (proxySocket.readyState === proxySocket.OPEN) {
              proxySocket.ping()
            } else {
              clearInterval(pingInterval)
            }
          }, 5000)

          proxySocket.on('error', (err: Error) => {
            this.logger.error(`Proxy WebSocket error: ${err.message}`, err.stack)
            clearInterval(pingInterval)
          })

          proxySocket.on('close', () => {
            this.logger.debug('Proxy WebSocket connection closed')
            clearInterval(pingInterval)
          })
        },
        error: (err: Error, req: any, res: any) => {
          this.logger.error(`Proxy error: ${err.message}`, err.stack)

          if (!res.headersSent) {
            try {
              res.writeHead(500, { 'Content-Type': 'text/plain' })
              res.end(`Proxy error: ${err.message}`)
            } catch (writeErr) {
              this.logger.error(`Failed to write error response: ${writeErr.message}`)
            }
          }
        },
      },
      proxyTimeout: 60 * 1000,
      timeout: 60 * 1000,
    }

    try {
      const proxy = createProxyMiddleware(proxyOptions)
      return proxy(req, res, next)
    } catch (error) {
      this.logger.error(`Failed to create proxy: ${error.message}`, error.stack)
      if (!res.headersSent) {
        try {
          res.writeHead(500, { 'Content-Type': 'text/plain' })
          res.end(`Failed to create proxy: ${error.message}`)
        } catch (writeErr) {
          this.logger.error(`Failed to write error response: ${writeErr.message}`)
        }
      }
    }
  }
}
