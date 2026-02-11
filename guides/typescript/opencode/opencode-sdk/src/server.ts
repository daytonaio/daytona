/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Sandbox } from '@daytonaio/sdk'

const PORT = 4096
const HOSTNAME = '0.0.0.0'
const SERVER_READY_LINE = 'opencode server listening'

export class Server {
  static async start(sandbox: Sandbox): Promise<{ baseUrl: string; ready: Promise<void> }> {
    const previewLink = await sandbox.getPreviewLink(PORT)
    const baseUrl = previewLink.url.replace(/\/$/, '')
    let resolveReady: () => void
    const ready = new Promise<void>((r) => { resolveReady = r })
    const sessionId = `opencode-serve-${Date.now()}`
    await sandbox.process.createSession(sessionId)
    const command = await sandbox.process.executeSessionCommand(sessionId, {
      command: `opencode serve --port ${PORT} --hostname ${HOSTNAME}`,
      runAsync: true,
    })
    if (!command.cmdId) throw new Error('Failed to start OpenCode server in sandbox')
    sandbox.process.getSessionCommandLogs(
      sessionId,
      command.cmdId,
      (stdout: string) => { if (stdout.includes(SERVER_READY_LINE)) resolveReady() },
      () => {},
    )
    return { baseUrl, ready }
  }
}
