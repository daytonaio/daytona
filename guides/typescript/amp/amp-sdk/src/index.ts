/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Daytona, Sandbox } from '@daytonaio/sdk'
import * as dotenv from 'dotenv'
import * as readline from 'readline'
import { AmpSession } from './session.js'

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

    // Register cleanup handler on process exit
    process.once('SIGINT', cleanup)

    // Install Amp CLI in the sandbox
    console.log('Installing Amp CLI...')
    await sandbox.process.executeCommand('npm install -g @sourcegraph/amp').then((r: any) => {
      if (r.exitCode) throw new Error('Error installing Amp CLI: ' + r.result)
    })

    // Daytona-aware system prompt (same pattern as letta-code agent)
    const previewLink = await sandbox.getPreviewLink(1234)
    const previewUrlPattern = previewLink.url.replace(/1234/, '{PORT}')
    const defaultSystemPrompt = [
      'You are running in a Daytona sandbox.',
      `When running services on localhost, they will be accessible as: ${previewUrlPattern}`,
      'ALWAYS end server commands with & so it runs in the background. For example, "(npm start) &" or "(python3 -m http.server 8000) &".',
      'ALWAYS run server commands as the last action in your turn.',
      'If you start a server, FIRST give the user the preview URL and THEN run the server command.',
    ].join(' ')

    const ampSession = new AmpSession(sandbox)
    await ampSession.initialize({
      systemPrompt: defaultSystemPrompt,
    })

    // Set up readline interface for user input
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout })

    // Register cleanup handler on readline SIGINT
    rl.once('SIGINT', cleanup)

    // Start the interactive prompt loop
    while (true) {
      const prompt = await new Promise<string>((resolve) => rl.question('User: ', resolve))
      if (prompt.trim()) await ampSession.processPrompt(prompt)
    }
  } catch (error) {
    console.error(error)
    if (sandbox) await sandbox.delete()
    process.exit(1)
  }
}

main().catch(console.error)
