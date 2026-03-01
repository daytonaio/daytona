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
      "echo '#{base64_code}' | base64 --decode | node /dev/stdin #{argv} 2>&1 | grep -vE 'npm notice'"
    end
  end
end
