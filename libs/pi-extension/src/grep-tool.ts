/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Remote grep.
 *
 * Pi's built-in grep tool always spawns ripgrep on the LOCAL machine and only
 * uses its GrepOperations to read context lines — so swapping operations does
 * not redirect the search. To grep a sandbox we must run the search inside it.
 *
 * This runs ripgrep in the sandbox (falling back to POSIX grep when rg is not
 * installed) and returns matching `path:line: text` lines, mirroring the shape
 * of Pi's grep output closely enough for the model.
 */

import type { Sandbox } from '@daytona/sdk'
import { execCommand } from './sandbox.ts'
import { shellQuote } from './util.ts'

/** Mirrors Pi's grepSchema. */
export interface GrepParams {
  pattern: string
  path?: string
  glob?: string
  ignoreCase?: boolean
  literal?: boolean
  context?: number
  limit?: number
}

const DEFAULT_LIMIT = 100

export interface RemoteGrepResult {
  content: { type: 'text'; text: string }[]
  details: undefined
}

export async function runRemoteGrep(sandbox: Sandbox, cwd: string, params: GrepParams): Promise<RemoteGrepResult> {
  const { pattern, path: searchDir = '.', glob, ignoreCase, literal, context, limit } = params
  // Guard against malformed input (NaN/Infinity/non-integer) before it reaches
  // `head -n`, which would otherwise become e.g. `head -n NaN`.
  const requested = limit ?? DEFAULT_LIMIT
  const max = Number.isFinite(requested) ? Math.max(1, Math.floor(requested)) : DEFAULT_LIMIT
  // Same guard as `max`: context is interpolated into `--context`/`-C`, so reject
  // Infinity/non-integer before it becomes e.g. `--context Infinity`. Default to 0
  // first so the value narrows to a number (mirrors find-tool.ts).
  const ctxRequested = context ?? 0
  const ctxLines = Number.isFinite(ctxRequested) && ctxRequested > 0 ? Math.floor(ctxRequested) : 0

  const rg = ['rg', '--line-number', '--no-heading', '--color=never', '--hidden']
  if (ignoreCase) rg.push('--ignore-case')
  if (literal) rg.push('--fixed-strings')
  if (ctxLines) rg.push('--context', String(ctxLines))
  if (glob) rg.push('--glob', shellQuote(glob))
  rg.push('--', shellQuote(pattern), shellQuote(searchDir))

  // POSIX grep fallback — always present even on minimal images.
  const gp = ['grep', '-rnI']
  if (ignoreCase) gp.push('-i')
  if (literal) gp.push('-F')
  if (ctxLines) gp.push('-C', String(ctxLines))
  if (glob) gp.push(`--include=${shellQuote(glob)}`)
  gp.push('--', shellQuote(pattern), shellQuote(searchDir))

  const command =
    `if command -v rg >/dev/null 2>&1; then ${rg.join(' ')}; ` + `else ${gp.join(' ')}; fi | head -n ${max}`

  const res = await execCommand(sandbox, command, cwd)
  const text = (res.result ?? res.artifacts?.stdout ?? '').replace(/\s+$/, '')
  const body = text.length > 0 ? text : `No matches found for /${pattern}/ in ${searchDir}`
  return { content: [{ type: 'text', text: body }], details: undefined }
}
