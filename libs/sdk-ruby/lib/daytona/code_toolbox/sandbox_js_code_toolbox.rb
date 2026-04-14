# frozen_string_literal: true

# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

require 'base64'

module Daytona
  class SandboxJsCodeToolbox
    def get_run_command(code, params = nil)
      # Prepend argv fix: node - places '-' at argv[1]; splice it out to match legacy node -e behaviour
      # Encode the provided code in base64
      base64_code = Base64.strict_encode64("process.argv.splice(1, 1);\n" + code)

      # Build command-line arguments string
      argv = ''
      argv = params.argv.join(' ') if params&.argv && !params.argv.empty?

      # Pipe the base64-encoded code via stdin to avoid OS ARG_MAX limits on large payloads
      # Use node - to read from stdin (node /dev/stdin does not work when stdin is a pipe)
      "printf '%s' '#{base64_code}' | base64 -d | node - #{argv}"
    end
  end
end
