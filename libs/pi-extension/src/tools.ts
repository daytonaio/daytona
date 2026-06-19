/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Tool registration.
 *
 * Each of Pi's built-in tools is replaced with a sandbox-backed variant:
 * - a sandbox is active        -> the tool runs inside it
 * - `--daytona` set, no sandbox -> the call FAILS (never run on your host)
 * - `--daytona` off            -> the extension is dormant, Pi's local tool runs
 *
 * The operation-backed tools (bash/read/write/edit/ls) share one wrapper;
 * find/grep run a dedicated in-sandbox search; preview_url is a custom tool.
 */

import type { Sandbox } from '@daytona/sdk'
import type { ExtensionAPI } from '@earendil-works/pi-coding-agent'
import {
  createBashTool,
  createEditTool,
  createFindTool,
  createGrepTool,
  createLsTool,
  createReadTool,
  createWriteTool,
} from '@earendil-works/pi-coding-agent'
import { Type } from 'typebox'
import { type FindParams, runRemoteFind } from './find-tool.ts'
import { type GrepParams, runRemoteGrep } from './grep-tool.ts'
import { createBashOps, createEditOps, createLsOps, createReadOps, createWriteOps } from './ops.ts'
import { withRecovery } from './sandbox.ts'

/** The bits of the active sandbox the tools need. */
export interface ToolSandbox {
  sandbox: Sandbox
  cwd: string
}

/**
 * Register all tools. `getActive` returns the sandbox bound to the current
 * session, or null when running locally.
 */
export function registerTools(pi: ExtensionAPI, getActive: () => ToolSandbox | null): void {
  const localCwd = process.cwd()
  const localBash = createBashTool(localCwd)
  const localRead = createReadTool(localCwd)
  const localWrite = createWriteTool(localCwd)
  const localEdit = createEditTool(localCwd)
  const localLs = createLsTool(localCwd)
  const localFind = createFindTool(localCwd)
  const localGrep = createGrepTool(localCwd)

  /**
   * Resolve the sandbox for a tool call. With `--daytona` on but no sandbox, we
   * throw rather than run on the host. With `--daytona` off, return null so the
   * local tool runs (the extension is dormant — it never broke normal Pi usage).
   */
  function requireSandbox(): ToolSandbox | null {
    const active = getActive()
    if (active) return active
    if (pi.getFlag('daytona') === true) {
      throw new Error('Daytona sandbox is unavailable — the tool was NOT run on your host. Restart Pi.')
    }
    return null
  }

  /** Wrap a tool so it runs against the sandbox (built per call) when one is active. */
  function sandboxTool<T extends { execute: (...args: never[]) => unknown }>(
    local: T,
    makeRemote: (cwd: string, sandbox: Sandbox) => T,
  ): T {
    return {
      ...local,
      execute: (...args: Parameters<T['execute']>) => {
        const active = requireSandbox()
        const tool = active ? makeRemote(active.cwd, active.sandbox) : local
        return tool.execute(...args)
      },
    } as T
  }

  pi.registerTool(sandboxTool(localBash, (cwd, sb) => createBashTool(cwd, { operations: createBashOps(sb) })))
  pi.registerTool(sandboxTool(localRead, (cwd, sb) => createReadTool(cwd, { operations: createReadOps(sb) })))
  pi.registerTool(sandboxTool(localWrite, (cwd, sb) => createWriteTool(cwd, { operations: createWriteOps(sb) })))
  pi.registerTool(sandboxTool(localEdit, (cwd, sb) => createEditTool(cwd, { operations: createEditOps(sb) })))
  pi.registerTool(sandboxTool(localLs, (cwd, sb) => createLsTool(cwd, { operations: createLsOps(sb) })))

  // find and grep can't be redirected via operations: Pi runs fd/ripgrep
  // locally, and Daytona's searchFiles only does basename matching. So we run
  // the search inside the sandbox via dedicated tools.
  pi.registerTool({
    ...localFind,
    async execute(id, params, signal, onUpdate) {
      const active = requireSandbox()
      if (active) {
        // We can only honor a pre-aborted signal here: Daytona's exec is a
        // single blocking, non-streaming call, so there's no mid-flight cancel
        // and no incremental output for onUpdate to stream.
        if (signal?.aborted) throw new Error('aborted')
        return runRemoteFind(active.sandbox, active.cwd, params as FindParams)
      }
      return localFind.execute(id, params, signal, onUpdate)
    },
  })

  pi.registerTool({
    ...localGrep,
    async execute(id, params, signal, onUpdate) {
      const active = requireSandbox()
      if (active) {
        // See find above: only a pre-aborted signal is honorable; Daytona's
        // exec is blocking and non-streaming, so no mid-flight cancel / onUpdate.
        if (signal?.aborted) throw new Error('aborted')
        return runRemoteGrep(active.sandbox, active.cwd, params as GrepParams)
      }
      return localGrep.execute(id, params, signal, onUpdate)
    },
  })

  // Custom tool: let the agent fetch a port's preview URL itself, so after it
  // starts a server (e.g. `npm run dev &`) it can hand the user a clickable
  // link without them running /sandbox url.
  pi.registerTool({
    name: 'preview_url',
    label: 'Preview URL',
    description:
      'Get the public preview URL for a port served inside the Daytona sandbox. ' +
      'Use this after starting a server (e.g. a dev server on port 3000) to give the user a link.',
    promptSnippet: 'Get a browser-openable preview URL for a port served in the sandbox',
    parameters: Type.Object({
      port: Type.Number({ description: 'The port the server listens on inside the sandbox' }),
    }),
    async execute(_id, { port }) {
      const active = getActive()
      if (!active) {
        return { content: [{ type: 'text', text: 'No active Daytona sandbox.' }], details: undefined }
      }
      const { sandbox } = active
      const link = await withRecovery(sandbox, () => sandbox.getPreviewLink(port))
      const text = sandbox.public
        ? `Preview URL for port ${port}: ${link.url}`
        : `Preview URL for port ${port}: ${link.url}\n` +
          `This is a private sandbox, so the URL needs an auth header:\n` +
          `  curl -H "x-daytona-preview-token: ${link.token}" ${link.url}`
      return { content: [{ type: 'text', text }], details: undefined }
    },
  })

  // Route user `!` bash commands to the sandbox. When --daytona is set but no
  // sandbox is available, return an error result so the command is NOT run on the
  // host. With --daytona off, return undefined to let Pi run it locally.
  pi.on('user_bash', () => {
    const active = getActive()
    // Route user `!` commands through sandbox operations. Pi runs these with the
    // HOST working directory (sessionManager.getCwd()), which doesn't exist in the
    // sandbox — passing it to Daytona's exec fails its chdir and surfaces as a
    // misleading "fork/exec <shell>: no such file" error. So pin the cwd to the
    // sandbox working dir (active.cwd), where the user expects `!` to run.
    if (active) return { operations: createBashOps(active.sandbox, active.cwd) }
    if (pi.getFlag('daytona') === true) {
      return {
        result: {
          output: 'Daytona sandbox is unavailable — the command was NOT run on your host. Restart Pi.',
          exitCode: 1,
          cancelled: false,
          truncated: false,
        },
      }
    }
    return
  })
}
