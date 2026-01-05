/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Daytona } from '@daytonaio/sdk'
import * as dotenv from 'dotenv'

// Load environment variables from .env file
dotenv.config()

// Port for the OpenCode web UI
const OPENCODE_PORT = 3000

// Generate a string to inject an environment variable with base64 decoding
function injectEnvVar(name: string, content: string): string {
  const base64 = Buffer.from(content).toString('base64')
  return `${name}=$(echo '${base64}' | base64 -d)`
}

async function main() {
  // Get the Daytona API key from environment variables
  if (!process.env.DAYTONA_API_KEY) {
    console.error('Error: DAYTONA_API_KEY environment variable is not set')
    console.error('Please create a .env file with your Daytona API key')
    process.exit(1)
  }

  // Initialize the Daytona client
  const daytona = new Daytona({ apiKey: process.env.DAYTONA_API_KEY })

  console.log('Creating sandbox...')
  const sandbox = await daytona.create()

  try {
    // Install OpenCode in the sandbox
    console.log('Installing OpenCode...')
    await sandbox.process.executeCommand('npm i -g opencode-ai@1.1.1')

    // Get the preview URL for the sandbox
    const previewLink = await sandbox.getPreviewLink(1234)
    const previewUrlPattern = previewLink.url.replace(/1234/, '{PORT}')

    // Configure the system prompt
    const systemPrompt = [
      'You are running in a Daytona sandbox.',
      'Use the /home/daytona directory instead of /workspace for file operations.',
      `When running services on localhost, they will be accessible as: ${previewUrlPattern}`,
      'When starting a server, always give the user the preview URL to access it.',
      'When starting a server, start it in the background with & so the command does not block further instructions.',
    ].join(' ')

    // OpenCode config with Daytona-aware agent
    const opencodeConfig = {
      $schema: 'https://opencode.ai/config.json',
      default_agent: 'daytona',
      agent: {
        daytona: {
          description: 'Daytona sandbox-aware coding agent',
          mode: 'primary',
          prompt: systemPrompt,
        },
      },
    }

    // Start OpenCode web server with config
    console.log('Starting OpenCode web server...')
    const configJson = JSON.stringify(opencodeConfig)

    // Create a session for running OpenCode
    const sessionId = `opencode-session-${Date.now()}`
    await sandbox.process.createSession(sessionId)

    // Run OpenCode web server asynchronously with config injected via environment variable
    const envVar = injectEnvVar('OPENCODE_CONFIG_CONTENT', configJson)
    const command = await sandbox.process.executeSessionCommand(sessionId, {
      command: `${envVar} opencode web --port ${OPENCODE_PORT}`,
      runAsync: true,
    })

    // Function to replace localhost URLs with the actual preview link
    const opencodePreviewLink = await sandbox.getPreviewLink(OPENCODE_PORT)
    const replaceUrl = (text: string) =>
      text.replace(new RegExp(`http:\\/\\/127\\.0\\.0\\.1:${OPENCODE_PORT}`, 'g'), opencodePreviewLink.url)

    // Stream output from the OpenCode server
    if (!command.cmdId) throw new Error('Failed to start OpenCode command in sandbox')
    sandbox.process.getSessionCommandLogs(
      sessionId,
      command.cmdId,
      (stdout: string) => console.log(replaceUrl(stdout).trim()),
      (stderr: string) => console.error(replaceUrl(stderr).trim()),
    )

    // Keep the process running until Ctrl+C is pressed
    console.log('Press Ctrl+C to stop.\n')
    process.stdin.resume()
    
    await new Promise<void>((resolve) => {
      process.once('SIGINT', async () => {
        console.log('\nCleaning up...')
        await sandbox.delete()
        process.stdin.pause()
        resolve()
      })
    })

  } catch (error) {
    console.error('Error:', error)
    await sandbox.delete()
    process.exit(1)
  }
}

main().catch((err) => {
  console.error('An error occurred:', err)
  process.exit(1)
})
