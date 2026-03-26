/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Daytona, Sandbox } from '@daytonaio/sdk'
import * as dotenv from 'dotenv'
import * as readline from 'readline'
import { LettaSession } from './letta-session'

// Load environment variables from .env file
dotenv.config()

async function main() {
  // Get the Daytona API key from environment variables
  const apiKey = process.env.DAYTONA_API_KEY

  if (!apiKey) {
    console.error('Error: DAYTONA_API_KEY environment variable is not set')
    console.error('Please create a .env file with your Daytona API key')
    process.exit(1)
  }

  // Check for Letta API key
  if (!process.env.SANDBOX_LETTA_API_KEY) {
    console.error('Error: SANDBOX_LETTA_API_KEY environment variable is not set')
    console.error('Please create a .env file with your Letta API key')
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
    // Create a new Daytona sandbox
    console.log('Creating sandbox...')
    sandbox = await daytona.create({
      envVars: { LETTA_API_KEY: process.env.SANDBOX_LETTA_API_KEY },
    })

    // Register cleanup handler on process exit
    process.once('SIGINT', cleanup)

    // Install Letta Code in the sandbox
    console.log('Installing Letta Code...')
    await sandbox.process.executeCommand('npm install -g @letta-ai/letta-code@0.12.5').then((r: any) => {
      if (r.exitCode) throw new Error('Error installing Letta Code: ' + r.result)
    })

    // Create the URL pattern for Daytona preview links
    // This is a URL where {PORT} is a placeholder for the port number
    // We first generate a preview link with the dummy port 1234, then replace it with {PORT}
    const previewLink = await sandbox.getPreviewLink(1234)
    const previewUrlPattern = previewLink.url.replace(/1234/, '{PORT}')

    // Configure the system prompt
    const systemPrompt = [
      'You are running in a Daytona sandbox.',
      `When running services on localhost, they will be accessible as: ${previewUrlPattern}`,
      'When starting a server, always give the user the preview URL to access it.',
    ].join(' ')

    // Start Letta Code using PTY for bidirectional communication
    const lettaSession = new LettaSession(sandbox)
    await lettaSession.initialize(systemPrompt)

    // Set up readline interface for user input
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout })

    // Register cleanup handler on readline SIGINT
    rl.once('SIGINT', cleanup)

    // Start the interactive prompt loop
    while (true) {
      const prompt = await new Promise<string>((resolve) => rl.question('User: ', resolve))
      if (prompt.trim()) await lettaSession.processPrompt(prompt)
    }
  } catch (error) {
    console.error(error)
    if (sandbox) await sandbox.delete()
    process.exit(1)
  }
}

main().catch(console.error)
