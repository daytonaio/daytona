/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { SandboxCodeToolbox } from '../Sandbox'
import { CodeRunParams } from '../Process'
import { Buffer } from 'buffer'

export class SandboxJsCodeToolbox implements SandboxCodeToolbox {
  public getRunCommand(code: string, params?: CodeRunParams): string {
    // Prepend argv fix: node - places '-' at argv[1]; splice it out to match legacy node -e behaviour
    const base64Code = Buffer.from('process.argv.splice(1, 1);\n' + code).toString('base64')
    const argv = params?.argv ? params.argv.join(' ') : ''

    // Pipe the base64-encoded code via stdin to avoid OS ARG_MAX limits on large payloads
    // Use node - to read from stdin (node /dev/stdin does not work when stdin is a pipe)
    return `printf '%s' '${base64Code}' | base64 -d | node - ${argv}`
  }
}
