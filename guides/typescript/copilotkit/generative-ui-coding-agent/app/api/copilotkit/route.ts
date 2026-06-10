/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { CopilotRuntime, copilotRuntimeNextJSAppRouterEndpoint } from '@copilotkit/runtime'
import { BuiltInAgent, defineTool } from '@copilotkit/runtime/v2'
import { Daytona } from '@daytona/sdk'
import type { NextRequest } from 'next/server'
import { z } from 'zod'

const daytona = new Daytona({ apiKey: process.env.DAYTONA_API_KEY })

const SYSTEM_PROMPT = `You are a coding agent with shell access to a fresh Daytona sandbox.

The user can ask you anything a developer might do at a terminal: build apps, debug or analyze code, run scripts, work with data, install packages, write tests, whatever fits the request.

Work under /home/daytona by default. Reuse the same sandboxId across every tool call. The sandbox auto-deletes after a period of inactivity; if a tool call fails because the sandbox no longer exists, call createSandbox again and continue with the new sandboxId.

When the user wants to see a running web app:

1. Prefer a modern, maintained scaffolder. Vite is the safest default for React/TS/SPA work; use \`npm create vite@latest <name> -- --template react-ts --yes\` or similar. Avoid \`create-react-app\`; it is deprecated and has very slow first-compile times.

2. ALWAYS bind the dev server to 0.0.0.0 or the Daytona proxy will not reach it. Cheat sheet:
   - Vite: \`vite --host 0.0.0.0 --port 5173\` (CLI flag) AND write a \`vite.config.ts\` with \`server: { host: '0.0.0.0', port: 5173, strictPort: true, hmr: { clientPort: 443, protocol: 'wss' } }\` so HMR survives the HTTPS proxy.
   - Next.js: \`next dev -H 0.0.0.0 -p 3000\`.
   - Express / Node: \`app.listen(PORT, '0.0.0.0')\`.
   - Flask: \`flask run --host 0.0.0.0 --port 5000\`.
   - FastAPI / Uvicorn: \`uvicorn main:app --host 0.0.0.0 --port 8000\`.

3. Use startWebServer with the dev-server command and its port. It starts the server in the background, waits for the port to be reachable, and returns the preview URL in one shot.

Reply to the user with one short sentence per turn. The tool cards in the chat carry the visual feedback.`

const createSandbox = defineTool({
  name: 'createSandbox',
  description:
    'Create a fresh Daytona sandbox with public preview URLs enabled. Call ONCE at session start; reuse the returned sandboxId for every subsequent tool call. Optionally inject environment variables, labels, or change the auto-stop interval.',
  parameters: z.object({
    envVars: z
      .record(z.string())
      .optional()
      .describe(
        'Environment variables to set inside the sandbox. Use this when the user provides API keys or other secrets the project needs.',
      ),
    labels: z.record(z.string()).optional().describe('Optional labels for organization-level sandbox tracking.'),
    autoStopInterval: z
      .number()
      .int()
      .nonnegative()
      .optional()
      .describe('Minutes of inactivity before the sandbox auto-stops. 0 disables, default 15.'),
  }),
  execute: async ({ envVars, labels, autoStopInterval }) => {
    const sandbox = await daytona.create({
      public: true,
      ephemeral: true,
      envVars,
      labels,
      autoStopInterval,
    })
    return { sandboxId: sandbox.id }
  },
})

const runCommand = defineTool({
  name: 'runCommand',
  description:
    'Execute a shell command in the sandbox. Set background:true for long-lived fire-and-forget processes (test watchers, build watchers, log followers) the agent will not need to interact with again. Use plain commands (rm, mv, mkdir, chmod, ...) for filesystem ops that do not need structured output. For dev servers the user should see in a browser, use startWebServer instead — it returns the preview URL atomically.',
  parameters: z.object({
    sandboxId: z.string(),
    command: z.string().describe('Shell command. Use && to chain. Absolute paths or `cd /home/daytona && ...`.'),
    background: z
      .boolean()
      .optional()
      .describe(
        'Run asynchronously and return immediately. Use for long-lived non-preview processes such as watchers or log tails; for user-visible dev servers, use startWebServer.',
      ),
  }),
  execute: async ({ sandboxId, command, background }) => {
    const sandbox = await daytona.get(sandboxId)
    if (background) {
      const sessionId = `bg-${Date.now()}`
      await sandbox.process.createSession(sessionId)
      const result = await sandbox.process.executeSessionCommand(sessionId, {
        command,
        runAsync: true,
      })
      return { background: true, sessionId, cmdId: result.cmdId, command }
    }
    const result = await sandbox.process.executeCommand(command)
    return { exitCode: result.exitCode, stdout: result.result, command }
  },
})

const writeFile = defineTool({
  name: 'writeFile',
  description: 'Write a file with the FULL new content. Overwrites if it exists.',
  parameters: z.object({
    sandboxId: z.string(),
    path: z.string().describe('Absolute path, e.g. "/home/daytona/app/src/App.tsx".'),
    content: z.string().describe('Complete new file content.'),
  }),
  execute: async ({ sandboxId, path, content }) => {
    const sandbox = await daytona.get(sandboxId)
    await sandbox.fs.uploadFile(Buffer.from(content), path)
    return { path, bytesWritten: Buffer.byteLength(content) }
  },
})

const readFile = defineTool({
  name: 'readFile',
  description: 'Read a file from the sandbox and return its text content.',
  parameters: z.object({
    sandboxId: z.string(),
    path: z.string().describe('Absolute path to the file in the sandbox.'),
  }),
  execute: async ({ sandboxId, path }) => {
    const sandbox = await daytona.get(sandboxId)
    const buf = await sandbox.fs.downloadFile(path)
    return { path, content: buf.toString('utf-8'), bytes: buf.length }
  },
})

const listFiles = defineTool({
  name: 'listFiles',
  description: 'List the contents of a directory in the sandbox.',
  parameters: z.object({
    sandboxId: z.string(),
    path: z.string().describe('Absolute directory path, e.g. "/home/daytona/app".'),
  }),
  execute: async ({ sandboxId, path }) => {
    const sandbox = await daytona.get(sandboxId)
    const files = await sandbox.fs.listFiles(path)
    return {
      path,
      entries: files.map((f) => ({
        name: f.name,
        isDir: f.isDir,
        size: f.size,
        permissions: f.permissions,
      })),
    }
  },
})

const findFiles = defineTool({
  name: 'findFiles',
  description: 'Search file CONTENTS for a regex pattern (grep-like). Returns file:line:content matches.',
  parameters: z.object({
    sandboxId: z.string(),
    path: z.string().describe('Directory to search under.'),
    pattern: z.string().describe('Regex pattern to match against file contents.'),
  }),
  execute: async ({ sandboxId, path, pattern }) => {
    const sandbox = await daytona.get(sandboxId)
    const matches = await sandbox.fs.findFiles(path, pattern)
    return { pattern, matches }
  },
})

const searchFiles = defineTool({
  name: 'searchFiles',
  description: 'Search file NAMES with a glob pattern (find-like). Returns matching paths.',
  parameters: z.object({
    sandboxId: z.string(),
    path: z.string().describe('Directory to search under.'),
    pattern: z.string().describe('Glob pattern for file names, e.g. "**/*.tsx".'),
  }),
  execute: async ({ sandboxId, path, pattern }) => {
    const sandbox = await daytona.get(sandboxId)
    const result = await sandbox.fs.searchFiles(path, pattern)
    return { pattern, files: result.files }
  },
})

const replaceInFiles = defineTool({
  name: 'replaceInFiles',
  description:
    'Codemod-style find-and-replace across multiple files in one call. Use instead of readFile + writeFile for bulk renames.',
  parameters: z.object({
    sandboxId: z.string(),
    files: z.array(z.string()).min(1).describe('Absolute file paths to apply the replacement in.'),
    pattern: z.string().describe('Text or regex pattern to find.'),
    newValue: z.string().describe('Replacement text.'),
  }),
  execute: async ({ sandboxId, files, pattern, newValue }) => {
    const sandbox = await daytona.get(sandboxId)
    const results = await sandbox.fs.replaceInFiles(files, pattern, newValue)
    return {
      pattern,
      newValue,
      results: results.map((r) => ({
        file: r.file,
        success: r.success,
        error: r.error,
      })),
    }
  },
})

const getFileDetails = defineTool({
  name: 'getFileDetails',
  description: 'Get metadata (size, mode, owner, modifiedAt) for a file or directory.',
  parameters: z.object({
    sandboxId: z.string(),
    path: z.string(),
  }),
  execute: async ({ sandboxId, path }) => {
    const sandbox = await daytona.get(sandboxId)
    const info = await sandbox.fs.getFileDetails(path)
    return {
      path,
      name: info.name,
      isDir: info.isDir,
      size: info.size,
      mode: info.mode,
      permissions: info.permissions,
      owner: info.owner,
      group: info.group,
      modifiedAt: info.modifiedAt,
    }
  },
})

const startWebServer = defineTool({
  name: 'startWebServer',
  description:
    'Start a dev server in the background AND return its public preview URL in ONE call. Blocks for up to 90 seconds while polling the dev server logs for a ready signal, so the iframe in the chat works as soon as this tool returns. Use for ANY hosted process the user should see in a browser (Vite, Next.js, Express, Flask, FastAPI, ...). Do NOT use runCommand({background: true}) for this; that will not return the URL and may end the turn before the server is ready.',
  parameters: z.object({
    sandboxId: z.string(),
    command: z
      .string()
      .describe(
        "Shell command that starts the dev server, e.g. 'cd /home/daytona/app && npm run dev'. Bind to 0.0.0.0 so the Daytona proxy can reach it.",
      ),
    port: z
      .number()
      .describe(
        'Port the dev server will listen on. The command you pass should pin the port explicitly (CLI flag like `--port 5173`, env var like `PORT=3000`, or an `app.listen(N)` call) so this value is deterministic. Defaults if you do not pin: Vite 5173, Next.js 3000, Flask 5000, FastAPI/Uvicorn 8000, CRA 3000.',
      ),
  }),
  execute: async ({ sandboxId, command, port }) => {
    const sandbox = await daytona.get(sandboxId)
    const sessionId = `web-${Date.now()}`
    await sandbox.process.createSession(sessionId)
    const bg = await sandbox.process.executeSessionCommand(sessionId, {
      command,
      runAsync: true,
    })

    if (bg.cmdId) {
      const urlPattern = new RegExp(`https?:\\/\\/[\\w\\.\\-\\[\\]:]+:${port}\\b`, 'i')
      const phrasePattern = new RegExp(
        `\\b(?:ready|listening|listen|started|running|serving)\\b[^\\n]{0,80}\\b${port}\\b|\\b${port}\\b[^\\n]{0,80}\\b(?:ready|listening|listen|started|running|serving)\\b`,
        'i',
      )
      const deadline = Date.now() + 90_000
      while (Date.now() < deadline) {
        const logs = await sandbox.process.getSessionCommandLogs(sessionId, bg.cmdId)
        const output = logs.output ?? ''
        if (urlPattern.test(output) || phrasePattern.test(output)) break
        await new Promise((r) => setTimeout(r, 750))
      }
    }

    const preview = await sandbox.getPreviewLink(port)
    return {
      sandboxId,
      port,
      url: preview.url,
      sessionId,
      cmdId: bg.cmdId,
      command,
    }
  },
})

const getPreviewUrl = defineTool({
  name: 'getPreviewUrl',
  description:
    'Get the public preview URL for a port on the sandbox. The port is opened automatically if it was closed. Call after starting a hosted process the user should see in a browser.',
  parameters: z.object({
    sandboxId: z.string(),
    port: z.number().describe('Port the hosted process is listening on.'),
  }),
  execute: async ({ sandboxId, port }) => {
    const sandbox = await daytona.get(sandboxId)
    const preview = await sandbox.getPreviewLink(port)
    return { url: preview.url, port }
  },
})

const agent = new BuiltInAgent({
  model: 'openai:gpt-5.4',
  prompt: SYSTEM_PROMPT,
  tools: [
    createSandbox,
    runCommand,
    writeFile,
    readFile,
    listFiles,
    findFiles,
    searchFiles,
    replaceInFiles,
    getFileDetails,
    startWebServer,
    getPreviewUrl,
  ],
  maxSteps: 30,
})

const runtime = new CopilotRuntime({
  agents: { default: agent },
})

export const POST = async (req: NextRequest) => {
  const { handleRequest } = copilotRuntimeNextJSAppRouterEndpoint({
    runtime,
    endpoint: '/api/copilotkit',
  })
  return handleRequest(req)
}
