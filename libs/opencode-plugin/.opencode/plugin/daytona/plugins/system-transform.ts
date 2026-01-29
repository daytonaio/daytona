/**
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Plugin, PluginInput } from '@opencode-ai/plugin'
import type { ExperimentalChatSystemTransformInput, ExperimentalChatSystemTransformOutput } from '../core/types'

/**
 * Creates the system transform plugin for Daytona
 * Adds Daytona-specific instructions to the system prompt
 */
export function createSystemTransformPlugin(repoPath: string): Plugin {
  return async (pluginCtx: PluginInput) => {
    return {
      'experimental.chat.system.transform': async (
        input: ExperimentalChatSystemTransformInput,
        output: ExperimentalChatSystemTransformOutput,
      ) => {
        output.system.push(`
        ## Daytona Sandbox Integration
        This session is integrated with a Daytona sandbox.
        The main project repository is located at: ${repoPath}.
        Bash commands will run in this directory.
        Work in this directory. Do NOT try to use the current working directory of the host system.
        When executing long-running commands, use the 'background' option to run them asynchronously.
        Before showing a preview URL, ensure the server is running in the sandbox on that port.
      `)
      },
    }
  }
}
