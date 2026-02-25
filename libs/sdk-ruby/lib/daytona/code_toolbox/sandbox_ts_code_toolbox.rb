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
      encoded_code = Base64.strict_encode64(code)

      argv = params&.argv&.join(' ') || ''

      # Pipe the base64-encoded code via stdin to avoid OS ARG_MAX limits on large payloads
      # Use /dev/stdin instead of -e "$(cat)" which would expand as a process arg and hit ARG_MAX
      # Capture the exit code before filtering to preserve ts-node's exit status
      "_dtn_out=$(echo '#{encoded_code}' | base64 -d | npx ts-node -O '{\"module\":\"CommonJS\"}' /dev/stdin #{argv} 2>&1); _dtn_ec=$?; " \
        "printf '%s\\n' \"$_dtn_out\" | grep -v 'npm notice'; exit $_dtn_ec"
    end
  end
end
