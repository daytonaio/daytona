/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Daytona } from '@daytonaio/sdk'
import * as dotenv from 'dotenv'
import * as readline from 'readline'
import Anthropic from '@anthropic-ai/sdk'

// Load environment variables from .env file
dotenv.config()
import { renderMarkdown } from './utils'

// ANSI color codes
const colors = {
  green: '\x1b[32m',
  reset: '\x1b[0m',
  bold: '\x1b[1m',
}

// Helper to print Project Manager messages in green
function printPM(message: string) {
  console.log(`${colors.green}${message}${colors.reset}`)
}

// Project Manager Agent - runs locally and manages the developer agent
class ProjectManagerAgent {
  private anthropic: Anthropic
  private conversationHistory: any[] = []

  constructor(apiKey: string) {
    this.anthropic = new Anthropic({ apiKey })
  }

  async processUserRequest(userMessage: string, sandbox: any, ctx: any): Promise<void> {
    // Add user message to conversation history
    this.conversationHistory.push({
      role: 'user',
      content: userMessage,
    })

    printPM('\n[Project Manager] Processing your request...\n')

    // Project Manager decides what to do and communicates with developer agent
    let continueLoop = true
    while (continueLoop) {
      const response = await this.anthropic.messages.create({
        model: 'claude-sonnet-4-20250514',
        max_tokens: 4096,
        system: `You are a Project Manager Agent. Your role is to:
1. Understand user requirements and break them down into clear tasks
2. Delegate coding tasks to a Developer Agent that works in a Daytona sandbox
3. Review the Developer Agent's responses and outputs
4. Communicate results back to the user

When you need the Developer Agent to do something:
- Use the <developer_task> tag to specify what you want the developer to do
- Wait for the developer's response which will include their output
- Analyze the results and decide if more work is needed

The Developer Agent has access to file operations, code execution, and can start services.
They have a preview URL available for port 80 and can start services on other ports.

When you're done with all tasks, say "TASK_COMPLETE" to finish.`,
        messages: this.conversationHistory,
      })

      // Extract assistant response
      const assistantMessage = response.content
        .filter((block: any) => block.type === 'text')
        .map((block: any) => block.text)
        .join('\n')

      printPM(`[Project Manager]: ${renderMarkdown(assistantMessage)}`)

      // Add assistant response to history
      this.conversationHistory.push({
        role: 'assistant',
        content: response.content,
      })

      // Check if Project Manager wants to delegate to Developer Agent
      const developerTaskMatch = assistantMessage.match(/<developer_task>([\s\S]*?)<\/developer_task>/)

      if (developerTaskMatch) {
        const developerTask = developerTaskMatch[1].trim()
        printPM('\n[Delegating to Developer Agent]...\n')

        // Get developer agent's response
        const developerOutput = await this.runDeveloperAgent(developerTask, sandbox, ctx)

        // Feed developer's response back to Project Manager
        this.conversationHistory.push({
          role: 'user',
          content: `Developer Agent completed the task. Here's their output:\n\n${developerOutput}`,
        })
      } else if (assistantMessage.includes('TASK_COMPLETE')) {
        continueLoop = false
        printPM('\n[Project Manager] All tasks completed!\n')
      } else {
        // Project Manager is done processing
        continueLoop = false
      }
    }
  }

  private async runDeveloperAgent(task: string, sandbox: any, ctx: any): Promise<string> {
    let output = ''

    console.log('[Developer Agent] Starting task...\n')

    const result = await sandbox.codeInterpreter.runCode(`coding_agent.run_query_sync(os.environ.get('PROMPT', ''))`, {
      context: ctx,
      envs: { PROMPT: task },
      onStdout: (msg: any) => {
        const rendered = renderMarkdown(msg.output)
        process.stdout.write(rendered) // White (default) text for developer
        output += msg.output
      },
      onStderr: (msg: any) => {
        const rendered = renderMarkdown(msg.output)
        process.stdout.write(rendered) // White (default) text for developer
        output += msg.output
      },
    })

    if (result.error) {
      const errorMsg = `Error: ${result.error.value}`
      console.error(errorMsg)
      output += `\n${errorMsg}`
    }

    console.log('\n[Developer Agent] Task completed.\n')

    return output || 'Developer Agent completed the task with no output.'
  }
}

async function main() {
  // Get the Daytona API key from environment variables
  const daytonaApiKey = process.env.DAYTONA_API_KEY

  if (!daytonaApiKey) {
    console.error('Error: DAYTONA_API_KEY environment variable is not set')
    console.error('Please create a .env file with your Daytona API key')
    process.exit(1)
  }

  // Check for Anthropic API keys - we need TWO now
  if (!process.env.ANTHROPIC_API_KEY) {
    console.error('Error: ANTHROPIC_API_KEY environment variable is not set')
    console.error('This is for the Project Manager Agent (local)')
    process.exit(1)
  }

  if (!process.env.SANDBOX_ANTHROPIC_API_KEY) {
    console.error('Error: SANDBOX_ANTHROPIC_API_KEY environment variable is not set')
    console.error('This is for the Developer Agent (sandbox)')
    process.exit(1)
  }

  // Initialize the Daytona client
  const daytona = new Daytona({ apiKey: daytonaApiKey })

  try {
    // Create a new Daytona sandbox for the Developer Agent
    console.log('Creating Developer Agent sandbox...')
    const sandbox = await daytona.create({
      envVars: {
        ANTHROPIC_API_KEY: process.env.SANDBOX_ANTHROPIC_API_KEY,
      },
    })

    // Install the Claude Agent SDK for Developer Agent
    console.log('Installing Developer Agent SDK...')
    await sandbox.process.executeCommand('python3 -m pip install claude-agent-sdk==0.1.16')

    // Initialize the code interpreter and upload the coding agent script
    console.log('Initializing Developer Agent...')
    const ctx = await sandbox.codeInterpreter.createContext()
    await sandbox.fs.uploadFile('src/coding_agent.py', '/tmp/coding_agent.py')
    const previewLink = await sandbox.getPreviewLink(80)
    await sandbox.codeInterpreter.runCode(`import os, coding_agent;`, {
      context: ctx,
      envs: { PREVIEW_URL: previewLink.url },
    })

    // Initialize the Project Manager Agent
    console.log('Initializing Project Manager Agent...')
    const projectManager = new ProjectManagerAgent(process.env.ANTHROPIC_API_KEY as string)

    // Set up readline interface for user input
    const rl = readline.createInterface({ input: process.stdin, output: process.stdout })
    rl.on('SIGINT', async () => {
      try {
        console.log('\nCleaning up...')
        await sandbox.delete()
      } catch (e) {
        console.error('Error deleting sandbox:', e)
      } finally {
        process.exit(0)
      }
    })

    // Start the interactive prompt loop
    console.log('\n=== Two-Agent System Ready ===')
    printPM('Project Manager Agent: Manages tasks and delegates to Developer (Green text)')
    console.log('Developer Agent: Executes code in Daytona sandbox (White text)')
    console.log('Press Ctrl+C at any time to exit.\n')

    while (true) {
      const prompt = await new Promise<string>((resolve) => rl.question('User: ', resolve))
      if (!prompt.trim()) continue
      await projectManager.processUserRequest(prompt, sandbox, ctx)
    }
  } catch (error) {
    console.error('An error occurred:', error)
    process.exit(1)
  }
}

main().catch(console.error)
