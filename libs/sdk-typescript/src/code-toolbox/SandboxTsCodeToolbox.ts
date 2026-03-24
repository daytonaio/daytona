/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { SandboxCodeToolbox } from '../Sandbox'
import { CodeRunParams } from '../Process'
import { Buffer } from 'buffer'

export class SandboxTsCodeToolbox implements SandboxCodeToolbox {
  public getRunCommand(code: string, params?: CodeRunParams): string {
    // Prepend argv fix: ts-node places the script path at argv[1]; splice it out to match legacy node -e behaviour
    const base64Code = Buffer.from('process.argv.splice(1, 1);\n' + code).toString('base64')
    const argv = params?.argv ? params.argv.join(' ') : ''

    // Pipe the base64-encoded code via stdin to avoid OS ARG_MAX limits on large payloads
    // ts-node does not support - for stdin; write to a temp file keyed on shell PID, execute, then clean up
    // Capture output to a second temp file so npm notice lines can be filtered without variable buffering
    return [
      `_f=/tmp/dtn_$$.ts`,
      `_o=/tmp/dtn_o_$$.log`,
      `printf '%s' '${base64Code}' | base64 -d > "$_f"`,
      `npx ts-node -T --ignore-diagnostics 5107 -O '{"module":"CommonJS"}' "$_f" ${argv} > "$_o" 2>&1`,
      `_dtn_ec=$?`,
      `rm -f "$_f"`,
      `grep -v -e 'npm notice' -e 'npm warn exec' "$_o" || true`,
      `rm -f "$_o"`,
      `exit $_dtn_ec`,
    ].join('; ')
  }
}
