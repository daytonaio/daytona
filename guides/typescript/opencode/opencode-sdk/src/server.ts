/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Sandbox } from '@daytonaio/sdk'

const PORT = 4096
const HOSTNAME = '0.0.0.0'
const SERVER_READY_LINE = 'opencode server listening'

// Inject an environment variable into a command string.
function injectEnvVar(name: string, content: string): string {
  const base64 = Buffer.from(content).toString('base64')
  return `${name}=$(echo '${base64}' | base64 -d)`
}

export class Server {

  // Start an OpenCode server in the sandbox with Daytona-aware agent config
  static async start(sandbox: Sandbox): Promise<{ baseUrl: string; ready: Promise<void> }> {
    const previewLink = await sandbox.getPreviewLink(PORT)
    const baseUrl = previewLink.url.replace(/\/$/, '')
    const previewUrlPattern = (await sandbox.getPreviewLink(1234)).url.replace(/1234/, '{PORT}')
    const systemPrompt = [
      'You are running in a Daytona sandbox.',
      'Use the /home/daytona directory instead of /workspace for file operations.',
      `When running services on localhost, they will be accessible as: ${previewUrlPattern}`,
      'When starting a server, always give the user the preview URL to access it.',
      'When starting a server, start it in the background with & so the command does not block further instructions.',
    ].join(' ')
    const opencodeConfig = JSON.stringify({
      $schema: 'https://opencode.ai/config.json',
      default_agent: 'daytona',
      agent: {
        daytona: {
          description: 'Daytona sandbox-aware coding agent',
          mode: 'primary',
          prompt: systemPrompt,
        },
      },
    })
    let resolveReady: () => void
    const ready = new Promise<void>((r) => { resolveReady = r })
    const sessionId = `opencode-serve-${Date.now()}`
    await sandbox.process.createSession(sessionId)
    const envVar = injectEnvVar('OPENCODE_CONFIG_CONTENT', opencodeConfig)
    const command = await sandbox.process.executeSessionCommand(sessionId, {
      command: `${envVar} opencode serve --port ${PORT} --hostname ${HOSTNAME}`,
      runAsync: true,
    })
    if (!command.cmdId) throw new Error('Failed to start OpenCode server in sandbox')

    // Resolve ready when stdout contains the server listening line.
    sandbox.process.getSessionCommandLogs(
      sessionId,
      command.cmdId,
      (stdout: string) => { if (stdout.includes(SERVER_READY_LINE)) resolveReady() },
      () => {},
    )
    return { baseUrl, ready }
  }
}
