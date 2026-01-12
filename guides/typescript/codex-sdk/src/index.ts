/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Daytona, Sandbox } from '@daytonaio/sdk'
import * as dotenv from 'dotenv'
import * as readline from 'readline'

// Load environment variables from .env file
dotenv.config()
import { renderMarkdown } from './utils'

// Generate a string of environment variables to prefix a shell command
function environmentPrefix(variables: Record<string, string>): string {
  const b64 = (v: string) => Buffer.from(v, 'utf8').toString('base64')
  return Object.entries(variables)
    .map(([name, value]) => `${name}="$(printf '%s' '${b64(value)}' | base64 --decode)"`)
    .join(' ')
}

async function processPrompt(prompt: string, sandbox: Sandbox): Promise<void> {
  console.log('Thinking...')

  // Create a session to stream the agent output
  const sessionId = `codex-session-${Date.now()}`
  await sandbox.process.createSession(sessionId)

  // Run the agent asynchronously, passing the prompt and OpenAI API key
  const command = await sandbox.process.executeSessionCommand(sessionId, {
    command: `${environmentPrefix({ PROMPT: prompt })} node /tmp/agent/index.ts`,
    runAsync: true,
  })

  // Stream agent output as it arrives
  if (!command.cmdId) throw new Error('Failed to start agent command in sandbox')
  await sandbox.process.getSessionCommandLogs(
    sessionId,
    command.cmdId,
    (stdout: string) => console.log(renderMarkdown(stdout.trim())),
    (stderr: string) => console.error(stderr.trim()),
  )

  // Delete the session
  await sandbox.process.deleteSession(sessionId)
}

async function main() {
  // Get the Daytona API key from environment variables
  const apiKey = process.env.DAYTONA_API_KEY

  if (!apiKey) {
    console.error('Error: DAYTONA_API_KEY environment variable is not set')
    console.error('Please create a .env file with your Daytona API key')
    process.exit(1)
  }

  // Check for OpenAI API key for the sandbox
  if (!process.env.SANDBOX_OPENAI_API_KEY) {
    console.error('Error: SANDBOX_OPENAI_API_KEY environment variable is not set')
    console.error('Please create a .env file with your OpenAI API key')
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
    console.log('Creating sandbox...')
    sandbox = await daytona.create({
      envVars: {
        OPENAI_API_KEY: process.env.SANDBOX_OPENAI_API_KEY || '',
      },
    })

    // Register cleanup handler on process exit
    process.once('SIGINT', cleanup)

    // Create the URL pattern for Daytona preview links
    // This is a URL where PORTNUMBER is a placeholder for the port number
    // We first generate a preview link with the dummy port 1234, then replace it with PORTNUMBER
    const previewLink = await sandbox.getPreviewLink(1234)
    const previewUrlPattern = previewLink.url.replace(/1234/, 'PORTNUMBER')

    // Configure the Codex system prompt
    const systemPrompt = [
      'You are running in a Daytona sandbox.',
      'Use the /home/daytona directory instead of /workspace for file operations.',
      `When running services on localhost, they will be accessible as: ${previewUrlPattern}`,
    ].join(' ')
    const config = `developer_instructions = "${systemPrompt}"`
    await sandbox.fs.createFolder('.codex', '755')
    await sandbox.fs.uploadFile(Buffer.from(config, 'utf8'), '.codex/config.toml')

    // Upload the NodeJS agent package into a temporary directory in the sandbox
    console.log('Installing Codex agent in sandbox...')
    await sandbox.fs.createFolder('/tmp/agent', '755')
    await sandbox.fs.uploadFile('./agent/index.ts', '/tmp/agent/index.ts')
    await sandbox.fs.uploadFile('./agent/package.json', '/tmp/agent/package.json')
    await sandbox.process.executeCommand('npm install --prefix /tmp/agent')

    // Set up readline interface for user input
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout })

    // Register cleanup handler on readline SIGINT
    rl.once('SIGINT', cleanup)

    // Start the interactive prompt loop
    console.log('Press Ctrl+C at any time to exit.')
    while (true) {
      const prompt = await new Promise<string>((resolve) => rl.question('User: ', resolve))
      if (!prompt.trim()) continue
      await processPrompt(prompt, sandbox)
    }
  } catch (error) {
    console.error('An error occurred:', error)
    if (sandbox) await sandbox.delete()
    process.exit(1)
  }
}

main().catch(console.error)
