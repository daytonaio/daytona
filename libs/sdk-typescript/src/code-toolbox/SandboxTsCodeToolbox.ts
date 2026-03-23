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
    // ts-node does not support reading from stdin via - or /dev/stdin when stdin is a pipe,
    // so write to a temp file, execute it, then clean up
    // Capture the exit code before filtering to preserve ts-node's exit status
    return [
      `_f=/tmp/dtn_$$.ts`,
      `printf '%s' '${base64Code}' | base64 -d > "$_f"`,
      `_dtn_out=$(npx ts-node -O '{"module":"CommonJS"}' "$_f" ${argv} 2>&1)`,
      `_dtn_ec=$?`,
      `rm -f "$_f"`,
      `printf '%s\\n' "$_dtn_out" | grep -v 'npm notice'`,
      `exit $_dtn_ec`,
    ].join('; ')
  }
}
