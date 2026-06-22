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
 * @param subprotocols - Additional WebSocket subprotocol tokens to negotiate (forwarded uniformly across runtimes)
 */
export async function createSandboxWebSocket(
  url: string,
  headers: Record<string, any>,
  getPreviewToken: () => Promise<string>,
  subprotocols?: string[],
): Promise<WebSocket> {
  if (RUNTIME === Runtime.BROWSER || RUNTIME === Runtime.DENO || RUNTIME === Runtime.SERVERLESS) {
    const previewToken = await getPreviewToken()
    const separator = url.includes('?') ? '&' : '?'
    return new WebSocket(`${url}${separator}DAYTONA_SANDBOX_AUTH_KEY=${previewToken}`, [
      `X-Daytona-SDK-Version~${String(headers['X-Daytona-SDK-Version'] ?? '')}`,
      ...(subprotocols ?? []),
    ])
  }

  // `ws` rejects the handshake if it offered subprotocols but the server selected none, and the
  // daemon only echoes the SDK-version token — so include it whenever we add others.
  const nodeSubprotocols =
    subprotocols && subprotocols.length
      ? [`X-Daytona-SDK-Version~${String(headers['X-Daytona-SDK-Version'] ?? '')}`, ...subprotocols]
      : []

  return new WebSocket(url, nodeSubprotocols, { headers })
}
