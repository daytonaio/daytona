# frozen_string_literal: true

require 'base64'

module Daytona
  class SandboxJsCodeToolbox
    def get_run_command(code, params = nil)
      # Encode the provided code in base64
      base64_code = Base64.strict_encode64(code)

      # Build command-line arguments string
      argv = ''
      argv = params.argv.join(' ') if params&.argv && !params.argv.empty?

      # Pipe the base64-encoded code via stdin to avoid OS ARG_MAX limits on large payloads
      # Use /dev/stdin instead of -e "$(cat)" which would expand as a process arg and hit ARG_MAX
      # Capture the exit code before filtering to preserve node's exit status
      "_dtn_out=$(echo '#{base64_code}' | base64 -d | node /dev/stdin #{argv} 2>&1); _dtn_ec=$?; " \
        "printf '%s\\n' \"$_dtn_out\" | grep -v 'npm notice'; exit $_dtn_ec"
    end
  end
end
