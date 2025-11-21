/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import WebSocket from 'isomorphic-ws'
import { RUNTIME, Runtime } from './Runtime'

/**
 * Creates an authenticated WebSocket connection to the sandbox toolbox.
 *
 * @param url - The websocket URL (ws[s]://...)
 * @param headers - Headers to forward when running in Node environments
 * @param getPreviewToken - Lazy getter for preview tokens (required for browser/serverless runtimes)
 */
export async function createSandboxWebSocket(
  url: string,
  headers: Record<string, any>,
  getPreviewToken: () => Promise<string>,
): Promise<WebSocket> {
  if (RUNTIME === Runtime.BROWSER || RUNTIME === Runtime.DENO || RUNTIME === Runtime.SERVERLESS) {
    const previewToken = await getPreviewToken()
    const separator = url.includes('?') ? '&' : '?'
    return new WebSocket(
      `${url}${separator}DAYTONA_SANDBOX_AUTH_KEY=${previewToken}`,
      `X-Daytona-SDK-Version~${String(headers['X-Daytona-SDK-Version'] ?? '')}`,
    )
  }

  return new WebSocket(url, { headers })
}
