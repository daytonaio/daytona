# frozen_string_literal: true

# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

require 'base64'

module Daytona
  class SandboxTsCodeToolbox
    # Get the run command for executing TypeScript code
    #
    # @param code [String] The TypeScript code to execute
    # @param params [Daytona::CodeRunParams, nil] Optional parameters for code execution
    # @return [String] The command to run the TypeScript code
    def get_run_command(code, params = nil)
      # Prepend argv fix: ts-node places the script path at argv[1]; splice it out to match legacy node -e behaviour
      encoded_code = Base64.strict_encode64("process.argv.splice(1, 1);\n" + code)

      argv = params&.argv&.join(' ') || ''

      # Pipe the base64-encoded code via stdin to avoid OS ARG_MAX limits on large payloads
      # ts-node does not support - for stdin; use shell PID ($$) for the temp file — each code_run spawns its own
      # shell process so $$ is unique across concurrent calls; cleaned up before exit
      # npm_config_loglevel=error suppresses npm notice/warn output at source, preserving streaming and real errors
      '_f=/tmp/dtn_$$.ts; ' \
        "printf '%s' '#{encoded_code}' | base64 -d > \"$_f\"; " \
        "npm_config_loglevel=error npx ts-node -T --ignore-diagnostics 5107 -O '{\"module\":\"CommonJS\"}' \"$_f\" #{argv}; " \
        '_dtn_ec=$?; ' \
        'rm -f "$_f"; ' \
        'exit $_dtn_ec'
    end
  end
end
