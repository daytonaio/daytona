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
      # ts-node does not support - for stdin; write to a temp file keyed on shell PID, execute, then clean up
      # Capture output to a second temp file so npm notice lines can be filtered without variable buffering
      '_f=/tmp/dtn_$$.ts; ' \
        '_o=/tmp/dtn_o_$$.log; ' \
        "printf '%s' '#{encoded_code}' | base64 -d > \"$_f\"; " \
        "npx ts-node -T --ignore-diagnostics 5107 -O '{\"module\":\"CommonJS\"}' \"$_f\" #{argv} > \"$_o\" 2>&1; " \
        '_dtn_ec=$?; ' \
        'rm -f "$_f"; ' \
        "grep -v -e 'npm notice' -e 'npm warn exec' \"$_o\" || true; " \
        'rm -f "$_o"; ' \
        'exit $_dtn_ec'
    end
  end
end
