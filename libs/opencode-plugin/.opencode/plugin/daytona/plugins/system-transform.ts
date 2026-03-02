/**
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { PluginInput } from '@opencode-ai/plugin'
import type { ExperimentalChatSystemTransformInput, ExperimentalChatSystemTransformOutput } from '../core/types'

/**
 * Adds Daytona-specific instructions to the system prompt.
 */
export async function systemPromptTransform(ctx: PluginInput, repoPath: string) {
  return async (input: ExperimentalChatSystemTransformInput, output: ExperimentalChatSystemTransformOutput) => {
    output.system.push(
      [
        '## Daytona Sandbox Integration',
        'This session is integrated with a Daytona sandbox.',
        `The main project repository is located at: ${repoPath}.`,
        'Bash commands will run in this directory.',
        'Put all projects in the project directory. Do NOT try to use the current working directory of the host system.',
        "When executing long-running commands, use the 'background' option to run them asynchronously.",
        'Before showing a preview URL, ensure the server is running in the sandbox on that port.',
      ].join('\n'),
    )
  }
}
