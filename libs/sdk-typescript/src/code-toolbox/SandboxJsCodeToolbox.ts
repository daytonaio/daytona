/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { SandboxCodeToolbox } from '../Sandbox'
import { CodeRunParams } from '../Process'
import { Buffer } from 'buffer'

export class SandboxJsCodeToolbox implements SandboxCodeToolbox {
  public getRunCommand(code: string, params?: CodeRunParams): string {
    const base64Code = Buffer.from(code).toString('base64')
    const argv = params?.argv ? params.argv.join(' ') : ''

    return `sh -c 'echo ${base64Code} | base64 --decode | node -e "$(cat)" ${argv} 2>&1 | grep -vE "npm notice"'`
  }
}
