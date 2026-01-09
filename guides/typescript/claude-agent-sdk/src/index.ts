/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Daytona, Sandbox, OutputMessage, ExecutionResult } from '@daytonaio/sdk'
import { InterpreterContext, ExecuteResponse } from '@daytonaio/toolbox-api-client'
import * as dotenv from 'dotenv'
import * as readline from 'readline'

// Load environment variables from .env file
dotenv.config()
import { renderMarkdown } from './utils'

async function processPrompt(prompt: string, sandbox: Sandbox, ctx: InterpreterContext): Promise<void> {
  console.log('Thinking...')

  const result = await sandbox.codeInterpreter.runCode(`coding_agent.run_query_sync(os.environ.get('PROMPT', ''))`, {
    context: ctx,
    envs: { PROMPT: prompt },
    onStdout: (msg: OutputMessage) => process.stdout.write(renderMarkdown(msg.output)),
    onStderr: (msg: OutputMessage) => process.stdout.write(renderMarkdown(msg.output)),
  })

  if (result.error) console.error('Execution error:', result.error.value)
}

async function main() {
  // Get the Daytona API key from environment variables
  const apiKey = process.env.DAYTONA_API_KEY

  if (!apiKey) {
    console.error('Error: DAYTONA_API_KEY environment variable is not set')
    console.error('Please create a .env file with your Daytona API key')
    process.exit(1)
  }

  // Check for Anthropic API key
  if (!process.env.SANDBOX_ANTHROPIC_API_KEY) {
    console.error('Error: SANDBOX_ANTHROPIC_API_KEY environment variable is not set')
    console.error('Please create a .env file with your Anthropic API key')
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
    // The sandbox language is irrelevant since we will use the code interpreter SDK
    console.log('Creating sandbox...')
    sandbox = await daytona.create({
      // Claude Code is memory intensive, so we use a medium snapshot
      snapshot: "daytona-medium", // This snapshot has 4GiB RAM and 2 vCPUs
      envVars: {
        ANTHROPIC_API_KEY: process.env.SANDBOX_ANTHROPIC_API_KEY,
      },
    })

    // Register cleanup handler on process exit
    process.once('SIGINT', cleanup)

    // Install the Claude Agent SDK
    console.log('Installing Agent SDK...')
    await sandbox.process.executeCommand('python3 -m pip install claude-agent-sdk').then((r: ExecuteResponse) => {
      if (r.exitCode) throw new Error('Error installing Agent SDK: ' + r.result)
    })

    // Initialize the code interpreter and upload the coding agent script
    console.log('Initializing Agent SDK...')
    const ctx = await sandbox.codeInterpreter.createContext()
    await sandbox.fs.uploadFile('src/coding_agent.py', '/tmp/coding_agent.py')
    const previewLink = await sandbox.getPreviewLink(80)
    await sandbox.codeInterpreter.runCode(`import os, coding_agent;`, {
      context: ctx,
      envs: { PREVIEW_URL: previewLink.url },
    }).then((r: ExecutionResult) => {
      if (r.error) throw new Error('Error initializing Agent SDK: ' + r.error.value)
    })

    // Set up readline interface for user input
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout })
    
    // Register cleanup handler on readline SIGINT
    rl.once('SIGINT', cleanup)

    // Start the interactive prompt loop
    console.log('Press Ctrl+C at any time to exit.')
    while (true) {
      const prompt = await new Promise<string>((resolve) => rl.question('User: ', resolve))
      if (!prompt.trim()) continue
      await processPrompt(prompt, sandbox, ctx)
    }
  } catch (error) {
    console.error(error)
    if (sandbox) await sandbox.delete()
    process.exit(1)
  }
}

main().catch(console.error)
