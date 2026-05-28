/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { anthropic } from '@ai-sdk/anthropic'
import { Daytona, Sandbox } from '@daytona/sdk'
import { ToolLoopAgent, stepCountIs, tool } from 'ai'
import { writeFileSync } from 'fs'
import { randomUUID } from 'node:crypto'
import * as dotenv from 'dotenv'
import { z } from 'zod'

dotenv.config()

const MODEL = anthropic('claude-sonnet-4-6')

const INSTRUCTIONS = `You are a careful, methodical performance research assistant running in a Linux sandbox.

Benchmark methodology:
- Warm up before timing (run a few untimed iterations first).
- Use >= 100,000 inner iterations per timing cell unless the workload is heavy.
- Use each language's high-resolution timer (e.g., time.perf_counter() in Python, performance.now() in Node).
- Repeat 5-7 times and report the median.
- Vary at least one axis (input size, hit rate, etc.).
- Be honest about surprising results. Acknowledge measurement noise.`

const BENCHMARK_PROMPT = `Implement the Sieve of Eratosthenes (find all primes up to N) in both Python and TypeScript. Benchmark each across N = 1_000, 10_000, 100_000, and 1_000_000.

Produce two artifacts on my local disk: ./sieve_benchmark.png with a chart comparing the languages, and ./findings.md with a markdown summary of your findings.

Keep your final answer concise; the full report is in the downloaded files.`

async function main(): Promise<void> {
  if (!process.env.DAYTONA_API_KEY) {
    console.error('Error: DAYTONA_API_KEY environment variable is not set')
    console.error('Please create a .env file with your Daytona API key')
    process.exit(1)
  }
  if (!process.env.ANTHROPIC_API_KEY) {
    console.error('Error: ANTHROPIC_API_KEY environment variable is not set')
    console.error('Please create a .env file with your Anthropic API key')
    process.exit(1)
  }

  const daytona = new Daytona()
  let sandbox: Sandbox | undefined

  try {
    console.log('Creating sandbox...')
    sandbox = await daytona.create()

    // Five tools that wrap Daytona primitives. Each has a single, non-overlapping
    // job so the agent never has to choose between two paths for the same goal:
    const runCode = tool({
      description:
        'Execute Python, JavaScript, or TypeScript source code in the sandbox. Pass standalone code; no project setup is required for any language. Use print() / console.log() to surface values. Returns the exit code and combined stdout/stderr.',
      inputSchema: z.object({
        code: z.string().describe('Source code to execute.'),
        language: z.enum(['python', 'javascript', 'typescript']).describe('Language of the code.'),
      }),
      execute: async ({ code, language }) => {
        const ext = { python: 'py', javascript: 'js', typescript: 'ts' }[language]
        const path = `/tmp/_run_${randomUUID()}.${ext}`
        await sandbox!.fs.uploadFile(Buffer.from(code, 'utf-8'), path)
        const cmd = {
          python: `python3 ${path}`,
          javascript: `node ${path}`,
          typescript:
            `ts-node --transpile-only --skipProject ` +
            `--compilerOptions '{"module":"commonjs","moduleResolution":"node"}' ` +
            path,
        }[language]
        const r = await sandbox!.process.executeCommand(cmd)
        return { exitCode: r.exitCode, output: r.result }
      },
    })

    const runCommand = tool({
      description:
        'Execute a bash shell command in the sandbox. Use for installing packages (pip, npm), running shell utilities (ls, head, wc, find), or chaining commands with pipes.',
      inputSchema: z.object({
        command: z.string().describe('Shell command to execute.'),
      }),
      execute: async ({ command }) => {
        const r = await sandbox!.process.executeCommand(command)
        return { exitCode: r.exitCode, output: r.result }
      },
    })

    const writeFile = tool({
      description:
        'Write text content to a file in the sandbox filesystem. Use for non-code artifacts: markdown reports, JSON data, config files. Overwrites any existing file at the path.',
      inputSchema: z.object({
        path: z.string().describe('Absolute path of the file to write.'),
        content: z.string().describe('UTF-8 text content to write.'),
      }),
      execute: async ({ path, content }) => {
        await sandbox!.fs.uploadFile(Buffer.from(content, 'utf-8'), path)
        return { path, bytes: Buffer.byteLength(content, 'utf-8') }
      },
    })

    const readFile = tool({
      description: 'Read a file from the sandbox filesystem and return its contents as text.',
      inputSchema: z.object({
        path: z.string().describe('Absolute path of the file to read.'),
      }),
      execute: async ({ path }) => {
        const buf = await sandbox!.fs.downloadFile(path)
        return { content: buf.toString('utf-8') }
      },
    })

    const downloadFile = tool({
      description:
        'Download a file from the sandbox to the local filesystem (where this script runs). Use this to extract generated artifacts (plots, datasets, reports) so they remain available after the sandbox is destroyed.',
      inputSchema: z.object({
        remotePath: z.string().describe('Absolute path of the file in the sandbox.'),
        localPath: z.string().describe('Path on the local filesystem to write the downloaded bytes to.'),
      }),
      execute: async ({ remotePath, localPath }) => {
        const buf = await sandbox!.fs.downloadFile(remotePath)
        writeFileSync(localPath, buf)
        return { localPath, bytes: buf.length }
      },
    })

    // Define the agent once. The same instance can be invoked from a CLI
    // (.generate), a streaming UI route (.stream), or via createAgentUIStream*
    // helpers - that is the win over passing the same config to generateText
    // at every call site.
    const agent = new ToolLoopAgent({
      model: MODEL,
      instructions: INSTRUCTIONS,
      tools: { runCode, runCommand, writeFile, readFile, downloadFile },
      stopWhen: stepCountIs(25),
    })

    console.log(`Prompt:\n${BENCHMARK_PROMPT}\n`)
    console.log('Running agent...')

    // Stream the agent's progress live: each tool call, each tool result, and
    // the final text answer all print as they happen. The alternative
    // (agent.generate(...)) waits silently for the whole run to finish and
    // dumps everything at the end - bad UX for runs that take a few minutes.
    const stream = await agent.stream({ prompt: BENCHMARK_PROMPT })

    let stepNum = 0

    for await (const part of stream.fullStream) {
      switch (part.type) {
        case 'start-step':
          stepNum++
          break
        case 'tool-call': {
          const input = JSON.stringify(part.input, null, 2)
          const preview = input.length > 600 ? input.slice(0, 600) + '\n... (truncated)' : input
          console.log(`\n--- Step ${stepNum}: ${part.toolName} ---\n${preview}`)
          break
        }
        case 'tool-result': {
          const out = JSON.stringify(part.output)
          const preview = out.length > 400 ? out.slice(0, 400) + '... (truncated)' : out
          console.log(`--- Step ${stepNum}: result ---\n${preview}`)
          break
        }
        case 'tool-error': {
          const message = part.error instanceof Error ? part.error.message : JSON.stringify(part.error)
          console.error(`\n--- Step ${stepNum}: ${part.toolName} ERROR ---\n${message}`)
          break
        }
        case 'text-start':
          console.log(`\n--- Step ${stepNum}: text ---`)
          break
        case 'text-delta':
          // Text streams in token-sized chunks. Use process.stdout.write so the
          // chunks concatenate on the same line as a continuous string, rather
          // than each chunk landing on its own line (which is what console.log
          // would do because it appends a newline).
          process.stdout.write(part.text)
          break
        case 'text-end':
          process.stdout.write('\n')
          break
        case 'error':
          console.error(`\n--- STREAM ERROR ---\n${part.error}`)
          break
      }
    }
  } finally {
    if (sandbox) {
      console.log('\nCleaning up sandbox...')
      await sandbox.delete()
    }
  }
}

main().catch((err) => {
  console.error('Error:', err)
  process.exit(1)
})
