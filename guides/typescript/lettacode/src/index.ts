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

interface AccumulatedContent {
  content: string
}

function handleParsedMessage(parsed: any, accumulated: AccumulatedContent, state: any): string | null {
  if (parsed.type === 'system' && parsed.subtype === 'init') {
    state.isInitialized = true
    return null
  }

  if (parsed.type === 'message') {
    const msgType = parsed.message_type

    if (msgType === 'approval_request_message') {
      const toolCall = parsed.tool_call
      if (!toolCall) return null

      const currentToolId = toolCall.tool_call_id

      // New tool call - flush the previous one
      if (currentToolId && state.lastToolId && currentToolId !== state.lastToolId) {
        const output = flushToolCall(accumulated, state)
        accumulated.content = toolCall.name || ''
        state.toolArgs = toolCall.arguments || ''
        state.lastToolId = currentToolId
        return output
      }

      // First tool call or same tool
      if (toolCall.name && !state.lastToolId) {
        accumulated.content = toolCall.name
        state.toolArgs = ''
        state.lastToolId = currentToolId
      }
      if (toolCall.arguments) {
        state.toolArgs = (state.toolArgs || '') + toolCall.arguments
      }
      return null
    }

    if (msgType === 'stop_reason') return flushToolCall(accumulated, state)
  }

  if (parsed.type === 'result') {
    accumulated.content = ''
    state.isComplete = true

    return `\n${renderMarkdown(parsed.result)}`
  }

  return null
}

function flushToolCall(accumulated: any, state: any): string | null {
  if (!accumulated.content) return null

  const toolName = accumulated.content
  let description = toolName

  try {
    const args = JSON.parse(state.toolArgs || '{}')
    description =
      args.description ||
      args.command ||
      (args.file_path && `${toolName} ${args.file_path}`) ||
      (args.query && `${toolName}: ${args.query}`) ||
      (args.url && `${toolName} ${args.url}`) ||
      toolName
  } catch {}

  accumulated.content = ''
  state.toolArgs = ''
  state.lastToolId = null
  return `\n🔧 ${description}`
}

async function processPrompt(prompt: string, ptyHandle: any, state: any): Promise<void> {
  console.log('Thinking...')

  state.isComplete = false
  state.lastActivityTime = Date.now()

  await ptyHandle.sendInput(
    JSON.stringify({
      type: 'user',
      message: { role: 'user', content: prompt },
    }) + '\n',
  )

  // Wait for the response to complete or timeout after 30 seconds of inactivity
  while (!state.isComplete) {
    await new Promise((resolve) => setTimeout(resolve, 100))

    // Check for timeout (30 seconds of no activity)
    if (Date.now() - state.lastActivityTime > 30000) {
      console.log('\n\n⏱️  Response timeout - no activity for 30 seconds')
      break
    }
  }

  console.log('\n')
}

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
    console.log('Starting Letta Code...')

    // Shared state for tracking completion and activity
    const state = { isComplete: false, isInitialized: false, lastActivityTime: Date.now() }
    let buffer = ''
    const accumulated: AccumulatedContent = { content: '' }

    const ptyHandle = await sandbox.process.createPty({
      id: `letta-pty-${Date.now()}`,
      cols: 120,
      rows: 30,
      onData: (data: Uint8Array) => {
        buffer += new TextDecoder().decode(data)
        state.lastActivityTime = Date.now()

        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          const trimmed = line.trim()
          try {
            const output = handleParsedMessage(JSON.parse(trimmed), accumulated, state)
            if (output) process.stdout.write(output)
          } catch {}
        }
      },
    })

    // Wait for PTY connection
    await ptyHandle.waitForConnection()

    // Start Letta Code command in the PTY with custom system prompt
    await ptyHandle.sendInput(
      `letta --new --system-custom "${systemPrompt.replace(/"/g, '\\"')}" --input-format stream-json --output-format stream-json --yolo -p\n`,
    )

    // Wait for agent to initialize
    console.log('Initializing agent...')
    while (!state.isInitialized) {
      await new Promise((resolve) => setTimeout(resolve, 100))
      if (Date.now() - state.lastActivityTime > 30000) throw new Error('Agent initialization timeout')
    }
    console.log('Agent initialized. Press Ctrl+C at any time to exit.\n')

    // Set up readline interface for user input
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout })
    
    // Register cleanup handler on readline SIGINT
    rl.once('SIGINT', cleanup)

    // Start the interactive prompt loop
    while (true) {
      const prompt = await new Promise<string>((resolve) => rl.question('User: ', resolve))
      if (prompt.trim()) await processPrompt(prompt, ptyHandle, state)
    }
  } catch (error) {
    console.error(error)
    if (sandbox) await sandbox.delete()
    process.exit(1)
  }
}

main().catch(console.error)
