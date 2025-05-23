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
    private readonly snapshotRef: string,
    private readonly authToken: string,
    private readonly follow: boolean,
    private readonly req: IncomingMessage,
    private readonly res: ServerResponse<IncomingMessage>,
    private readonly next: NextFunction,
  ) {}

  create() {
    const proxyOptions: Options = {
      target: this.targetUrl,
      secure: false,
      changeOrigin: true,
      autoRewrite: true,
      pathRewrite: () => `/snapshots/logs?snapshotRef=${this.snapshotRef}&follow=${this.follow}`,
      on: {
        proxyReq: (proxyReq: any, req: any) => {
          proxyReq.setHeader('Authorization', `Bearer ${this.authToken}`)
          proxyReq.setHeader('Accept', 'application/octet-stream')
          fixRequestBody(proxyReq, req)
        },
      },
      proxyTimeout: 5 * 60 * 1000,
    }

    return createProxyMiddleware(proxyOptions)(this.req, this.res, this.next)
  }
}
