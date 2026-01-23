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

      # Combine everything into the final command for JavaScript
      " sh -c 'echo #{base64_code} | base64 --decode | node -e \"$(cat)\" #{argv} 2>&1 | grep -vE \"npm notice\"' "
    end
  end
end
