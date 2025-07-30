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

    // eslint-disable-next-line no-useless-escape
    return `sh -c 'echo ${base64Code} | base64 --decode | npx ts-node -O "{\\\"module\\\":\\\"CommonJS\\\"}" -e "$(cat)" x ${argv} 2>&1 | grep -vE "npm notice"'`
  }
}
