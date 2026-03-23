# frozen_string_literal: true

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
      # ts-node does not support reading from stdin via - or /dev/stdin when stdin is a pipe,
      # so write to a temp file, execute it, then clean up
      # Capture the exit code before filtering to preserve ts-node's exit status
      '_f=/tmp/dtn_$$.ts; ' \
        "printf '%s' '#{encoded_code}' | base64 -d > \"$_f\"; " \
        "_dtn_out=$(npx ts-node -O '{\"module\":\"CommonJS\"}' \"$_f\" #{argv} 2>&1); " \
        '_dtn_ec=$?; ' \
        'rm -f "$_f"; ' \
        "printf '%s\\n' \"$_dtn_out\" | grep -v 'npm notice'; " \
        'exit $_dtn_ec'
    end
  end
end
