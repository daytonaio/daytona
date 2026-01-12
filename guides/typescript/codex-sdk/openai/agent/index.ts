/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

// This is the NodeJS Codex agent used inside the Daytona sandbox.
// This script is uploaded to the sandbox and invoked with PROMPT in the environment.

import { Codex, Thread } from '@openai/codex-sdk'
import type {
  ThreadOptions,
  ThreadItem,
  TodoListItem,
  CommandExecutionItem,
  FileChangeItem,
  McpToolCallItem,
  WebSearchItem,
  ErrorItem,
  AgentMessageItem,
  Usage,
} from '@openai/codex-sdk'
import fs from 'fs/promises'

// Convert output from the agent to a printable string
function toolCallToString(item: ThreadItem): string {
  if (item.type === 'command_execution') {
    const commandItem = item as CommandExecutionItem
    const status = commandItem.status === 'completed' && commandItem.exit_code === 0 ? '✓' : '✗'
    return `🔨 ${status} Run: \`${commandItem.command}\``
  } else if (item.type === 'file_change') {
    const fileChangeItem = item as FileChangeItem
    const capitalize = (s: string) => s.charAt(0).toUpperCase() + s.slice(1)
    return fileChangeItem.changes.map((change) => `📝 ${capitalize(change.kind)} ${change.path}`).join('\n')
  } else if (item.type === 'mcp_tool_call') {
    const mcpToolCallItem = item as McpToolCallItem
    return `🔧 MCP Tool: ${mcpToolCallItem.tool}`
  } else if (item.type === 'web_search') {
    const webSearchItem = item as WebSearchItem
    return `🌐 Web search: ${webSearchItem.query}`
  } else if (item.type === 'todo_list') {
    const todoListItem = item as TodoListItem
    const todoList = todoListItem.items.map((todo) => `  - [${todo.completed ? 'x' : ' '}] ${todo.text}`).join('\n')
    return `🗒️ To-do list:\n` + todoList
  } else if (item.type === 'error') {
    const errorItem = item as ErrorItem
    return `❌ Error: ${errorItem.message}`
  } else if (item.type === 'agent_message') {
    const agentMessageItem = item as AgentMessageItem
    return agentMessageItem.text
  }
  return ''
}

// Read a file if it exists, or return undefined
async function readFileIfExisting(path: string): Promise<string | undefined> {
  return fs.readFile(path, 'utf8').catch(() => undefined)
}

async function main(): Promise<void> {
  // Get the OpenAI API key from environment variables
  const apiKey = process.env.OPENAI_API_KEY
  if (!apiKey) {
    console.error('Error: OpenAI API key not provided in sandbox environment')
    process.exit(1)
  }

  // Initialize the Codex SDK
  // Using an empty environment object prevents environment variables from being passed to the agent
  const codex = new Codex({ apiKey, env: {} })
  const threadIdPath = '/tmp/codex-thread-id'
  const threadId = (await readFileIfExisting(threadIdPath))?.trim()

  // Configure Codex options
  const options: ThreadOptions = {
    workingDirectory: '/home/daytona',
    skipGitRepoCheck: true,
    sandboxMode: 'danger-full-access',
  }
  const thread: Thread = threadId ? codex.resumeThread(threadId, options) : codex.startThread(options)

  // Run the Codex agent and stream the output
  const { events } = await thread.runStreamed(process.env.PROMPT ?? '')
  for await (const event of events) {
    if (event.type === 'item.completed') {
      // Print each completed item
      const output = toolCallToString(event.item).trim()
      if (output) console.log(output)
    } else if (event.type === 'turn.completed') {
      // Print usage summary when the agent is finished
      const { cached_input_tokens, input_tokens, output_tokens } = event.usage as Usage
      console.log(`*Usage Summary: Cached: ${cached_input_tokens}, Input: ${input_tokens}, Output: ${output_tokens}*`)
    }
  }
  
  // Save the thread ID for future runs
  if (!threadId && thread.id) {
    await fs.writeFile(threadIdPath, thread.id, 'utf8')
  }
}

main().catch((err) => {
  console.error('Agent error:', err)
  process.exit(1)
})
