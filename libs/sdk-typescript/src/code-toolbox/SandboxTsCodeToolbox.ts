/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { SandboxCodeToolbox } from '../Sandbox'
import { CodeRunParams } from '../Process'
import { Buffer } from 'buffer'

export class SandboxTsCodeToolbox implements SandboxCodeToolbox {
  public getRunCommand(code: string, params?: CodeRunParams): string {
    const base64Code = Buffer.from(code).toString('base64')
    const argv = params?.argv ? params.argv.join(' ') : ''

    // Pipe the base64-encoded code via stdin to avoid OS ARG_MAX limits on large payloads
    // Use /dev/stdin instead of -e "$(cat)" which would expand as a process arg and hit ARG_MAX
    return `echo '${base64Code}' | base64 --decode | npx ts-node -O '{"module":"CommonJS"}' /dev/stdin ${argv} 2>&1 | grep -vE "npm notice"`
  }
}
