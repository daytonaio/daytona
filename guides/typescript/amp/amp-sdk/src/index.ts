/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Daytona, Sandbox } from '@daytona/sdk'
import * as dotenv from 'dotenv'
import * as readline from 'readline'
import { AmpSession } from './session.js'

// Load environment variables from .env file
dotenv.config()

function formatCommandPreview(input: string, maxLength = 80): string {
  const normalized = input.replace(/\s+/g, ' ').trim()
  return normalized.length > maxLength ? `${normalized.slice(0, maxLength)}...` : normalized
}

async function main() {
  // Get the Daytona API key from environment variables
  const apiKey = process.env.DAYTONA_API_KEY

  if (!apiKey) {
    console.error('Error: DAYTONA_API_KEY environment variable is not set')
    console.error('Please create a .env file with your Daytona API key')
    process.exit(1)
  }

  // Check for Amp API key
  if (!process.env.SANDBOX_AMP_API_KEY) {
    console.error('Error: SANDBOX_AMP_API_KEY environment variable is not set')
    console.error('Please create a .env file with your Amp API key')
    console.error('')
    console.error('Note: Amp execute mode requires paid credits.')
    console.error('Add credits at https://ampcode.com/pay before running this example.')
    process.exit(1)
  }

  // Initialize the Daytona client
  const daytona = new Daytona({ apiKey })

  let sandbox: Sandbox | undefined

  // Reusable cleanup handler to delete the sandbox on exit
  const cleanup = async () => {
    try {
      console.log('\nCleaning up...')
      if (sandbox) await sandbox.delete()
    } catch (e) {
      console.error('Error deleting sandbox:', e)
    } finally {
      process.exit(0)
    }
  }

  try {
    // Create a new Daytona sandbox with Amp API key
    console.log('Creating sandbox...')
    sandbox = await daytona.create({
      envVars: { AMP_API_KEY: process.env.SANDBOX_AMP_API_KEY },
    })
    const activeSandbox = sandbox

    // Register cleanup handler on process exit
    process.once('SIGINT', cleanup)

    // Install Amp CLI in the sandbox
    console.log('Installing Amp CLI...')
    const installResult = await sandbox.process.executeCommand('npm install -g @sourcegraph/amp')
    if (installResult.exitCode !== 0) {
      throw new Error('Error installing Amp CLI: ' + installResult.result)
    }

    // Daytona-aware system prompt.
    // We ask Amp to write server commands to start.sh instead of running them directly,
    // because background server commands in Amp execute mode can block/hang the turn.
    const previewLink = await sandbox.getPreviewLink(1234)
    const previewUrlPattern = previewLink.url.replace(/1234/, '{PORT}')
    const defaultSystemPrompt = [
      'You are running in a Daytona sandbox.',
      `When running services on localhost, they will be accessible as: ${previewUrlPattern}`,
      'When you need to start a server, DO NOT run it directly.',
      'Instead, write only the server start command to /home/daytona/start.sh (one command, no markdown).',
      'After writing the start command, provide the preview URL to the user.',
      'Start the conversation with a greeting and ask the user what they would like to do.',
    ].join(' ')

    const ampSession = new AmpSession(activeSandbox)
    await ampSession.initialize({
      systemPrompt: defaultSystemPrompt,
    })
    const serverSessions: string[] = []

    const startServerFromScript = async () => {
      // Only run when Amp has produced a start script for this turn.
      const startScriptCheck = await activeSandbox.process.executeCommand('test -f /home/daytona/start.sh')
      if (startScriptCheck.exitCode !== 0) {
        return
      }

      const startScriptContents = (await activeSandbox.fs.downloadFile('/home/daytona/start.sh')).toString('utf-8')
      const clippedStartScript = formatCommandPreview(startScriptContents)
      console.log(`Running \`${clippedStartScript}\` via session command...`)
      // Execute server startup outside Amp so long-running/background commands
      // do not keep the Amp response from completing.
      const sessionId = `amp-server-session-${Date.now()}`
      await activeSandbox.process.createSession(sessionId)
      serverSessions.push(sessionId)

      await activeSandbox.process.executeSessionCommand(sessionId, {
        command: 'cd /home/daytona && chmod +x start.sh && ./start.sh',
        runAsync: true,
      })
    }

    // Set up readline interface for user input
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout })

    // Enhanced cleanup handler to also cleanup the PTY session
    const cleanupWithSession = async () => {
      try {
        console.log('\nCleaning up...')
        await ampSession.cleanup()
        await Promise.allSettled(serverSessions.map((id) => activeSandbox.process.deleteSession(id)))
        if (sandbox) await sandbox.delete()
      } catch (e) {
        console.error('Error during cleanup:', e)
      } finally {
        process.exit(0)
      }
    }

    // Re-register cleanup handler with PTY cleanup
    process.removeAllListeners('SIGINT')
    process.once('SIGINT', cleanupWithSession)
    rl.once('SIGINT', cleanupWithSession)

    // Start the interactive prompt loop
    while (true) {
      const prompt = await new Promise<string>((resolve) => rl.question('User: ', resolve))
      if (prompt.trim()) {
        await ampSession.processPrompt(prompt)
        await startServerFromScript()
      }
    }
  } catch (error) {
    console.error(error)
    if (sandbox) await sandbox.delete()
    process.exit(1)
  }
}

main().catch(console.error)
