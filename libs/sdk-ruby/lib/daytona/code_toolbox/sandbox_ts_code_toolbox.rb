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
      encoded_code = Base64.encode64(code)

      argv = params&.argv&.join(' ') || ''

      # Execute TypeScript code using ts-node with ESM support
      " sh -c 'echo #{encoded_code} | base64 --decode | npx ts-node -O " \
        "\"{\\\"module\\\":\\\"CommonJS\\\"}\" -e \"$(cat)\" x #{argv} 2>&1 | grep -vE " \
        "\"npm notice\"' "
    end
  end
end
