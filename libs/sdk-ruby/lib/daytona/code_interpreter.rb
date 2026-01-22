# frozen_string_literal: true

require 'json'
require 'websocket-client-simple'

module Daytona
  # Handles code interpretation and execution within a Sandbox. Currently supports only Python.
  #
  # This class provides methods to execute code in isolated interpreter contexts,
  # manage contexts, and stream execution output via callbacks. If subsequent code executions
  # are performed in the same context, the variables, imports, and functions defined in
  # the previous execution will be available.
  #
  # For other languages, use the `code_run` method from the `Process` interface,
  # or execute the appropriate command directly in the sandbox terminal.
  class CodeInterpreter
    WEBSOCKET_TIMEOUT_CODE = 4008

    # @param sandbox_id [String]
    # @param toolbox_api [DaytonaToolboxApiClient::InterpreterApi]
    # @param get_preview_link [Proc]
    def initialize(sandbox_id:, toolbox_api:, get_preview_link:)
      @sandbox_id = sandbox_id
      @toolbox_api = toolbox_api
      @get_preview_link = get_preview_link
    end

    # Execute Python code in the sandbox.
    #
    # By default, code runs in the default shared context which persists variables,
    # imports, and functions across executions. To run in an isolated context,
    # create a new context with `create_context` and pass it as the `context` argument.
    #
    # @param code [String] Code to execute
    # @param context [DaytonaToolboxApiClient::InterpreterContext, nil] Context to run code in
    # @param on_stdout [Proc, nil] Callback for stdout messages (receives OutputMessage)
    # @param on_stderr [Proc, nil] Callback for stderr messages (receives OutputMessage)
    # @param on_error [Proc, nil] Callback for execution errors (receives ExecutionError)
    # @param envs [Hash<String, String>, nil] Environment variables for this execution
    # @param timeout [Integer, nil] Timeout in seconds. 0 means no timeout. Default is 10 minutes.
    # @return [Daytona::ExecutionResult]
    # @raise [Daytona::Sdk::Error]
    #
    # @example
    #   def handle_stdout(msg)
    #     print "STDOUT: #{msg.output}"
    #   end
    #
    #   def handle_stderr(msg)
    #     print "STDERR: #{msg.output}"
    #   end
    #
    #   def handle_error(err)
    #     puts "ERROR: #{err.name}: #{err.value}"
    #   end
    #
    #   code = <<~PYTHON
    #     import sys
    #     import time
    #     for i in range(5):
    #         print(i)
    #         time.sleep(1)
    #     sys.stderr.write("Counting done!")
    #   PYTHON
    #
    #   result = sandbox.code_interpreter.run_code(
    #     code,
    #     on_stdout: method(:handle_stdout),
    #     on_stderr: method(:handle_stderr),
    #     on_error: method(:handle_error),
    #     timeout: 10
    #   )
    def run_code(code, context: nil, on_stdout: nil, on_stderr: nil, on_error: nil, envs: nil, timeout: nil) # rubocop:disable Metrics/AbcSize, Metrics/CyclomaticComplexity, Metrics/MethodLength, Metrics/ParameterLists, Metrics/PerceivedComplexity
      # Get WebSocket URL
      base_url = @toolbox_api.api_client.config.base_url
      ws_url = base_url.sub(%r{^http}, 'ws') + '/process/interpreter/execute'

      result = ExecutionResult.new

      # Create request payload
      request = { code: }
      request[:contextId] = context.id if context
      request[:envs] = envs if envs
      request[:timeout] = timeout if timeout

      # Connect to WebSocket and execute
      ws = WebSocket::Client::Simple.connect(ws_url, headers: build_headers)

      ws.on :open do
        ws.send(JSON.dump(request))
      end

      ws.on :message do |msg|
        handle_message(msg.data, result, on_stdout, on_stderr, on_error)
      end

      ws.on :error do |e|
        raise Sdk::Error, "WebSocket error: #{e.message}"
      end

      ws.on :close do |e|
        handle_close(e)
      end

      # Wait for completion
      sleep 0.01 until ws.close?

      result
    rescue StandardError => e
      raise Sdk::Error, "Failed to run code: #{e.message}"
    end

    # Create a new isolated interpreter context.
    #
    # Contexts provide isolated execution environments with their own global namespace.
    # Variables, imports, and functions defined in one context don't affect others.
    #
    # @param cwd [String, nil] Working directory for the context
    # @return [DaytonaToolboxApiClient::InterpreterContext]
    # @raise [Daytona::Sdk::Error]
    #
    # @example
    #   # Create isolated context
    #   ctx = sandbox.code_interpreter.create_context
    #
    #   # Execute code in this context
    #   sandbox.code_interpreter.run_code("x = 100", context: ctx)
    #
    #   # Variable only exists in this context
    #   result = sandbox.code_interpreter.run_code("print(x)", context: ctx)  # OK
    #
    #   # Won't see the variable in default context
    #   result = sandbox.code_interpreter.run_code("print(x)")  # NameError
    #
    #   # Clean up
    #   sandbox.code_interpreter.delete_context(ctx)
    def create_context(cwd: nil)
      request = DaytonaToolboxApiClient::CreateContextRequest.new(cwd:)
      @toolbox_api.create_interpreter_context(request)
    rescue StandardError => e
      raise Sdk::Error, "Failed to create interpreter context: #{e.message}"
    end

    # List all user-created interpreter contexts.
    #
    # The default context is not included in this list. Only contexts created
    # via `create_context` are returned.
    #
    # @return [Array<DaytonaToolboxApiClient::InterpreterContext>]
    # @raise [Daytona::Sdk::Error]
    #
    # @example
    #   contexts = sandbox.code_interpreter.list_contexts
    #   contexts.each do |ctx|
    #     puts "Context #{ctx.id}: #{ctx.language} at #{ctx.cwd}"
    #   end
    def list_contexts
      response = @toolbox_api.list_interpreter_contexts
      response.contexts || []
    rescue StandardError => e
      raise Sdk::Error, "Failed to list interpreter contexts: #{e.message}"
    end

    # Delete an interpreter context and shut down all associated processes.
    #
    # This permanently removes the context and all its state (variables, imports, etc.).
    # The default context cannot be deleted.
    #
    # @param context [DaytonaToolboxApiClient::InterpreterContext]
    # @return [void]
    # @raise [Daytona::Sdk::Error]
    #
    # @example
    #   ctx = sandbox.code_interpreter.create_context
    #   # ... use context ...
    #   sandbox.code_interpreter.delete_context(ctx)
    def delete_context(context)
      @toolbox_api.delete_interpreter_context(context.id)
      nil
    rescue StandardError => e
      raise Sdk::Error, "Failed to delete interpreter context: #{e.message}"
    end

    private

    # @return [Hash<String, String>]
    def build_headers
      headers = {}
      @toolbox_api.api_client.update_params_for_auth!(headers, nil, ['bearer'])
      headers
    end

    # @param data [String]
    # @param result [Daytona::ExecutionResult]
    # @param on_stdout [Proc, nil]
    # @param on_stderr [Proc, nil]
    # @param on_error [Proc, nil]
    # @return [void]
    def handle_message(data, result, on_stdout, on_stderr, on_error) # rubocop:disable Metrics/AbcSize
      chunk = JSON.parse(data)
      chunk_type = chunk['type']

      case chunk_type
      when 'stdout'
        stdout = chunk['text'] || ''
        result.stdout += stdout
        on_stdout&.call(OutputMessage.new(output: stdout))
      when 'stderr'
        stderr = chunk['text'] || ''
        result.stderr += stderr
        on_stderr&.call(OutputMessage.new(output: stderr))
      when 'error'
        error = ExecutionError.new(
          name: chunk['name'] || '',
          value: chunk['value'] || '',
          traceback: chunk['traceback'] || ''
        )
        result.error = error
        on_error&.call(error)
      end
    end

    # @param event [Object]
    # @return [void]
    def handle_close(event)
      code = event.code
      reason = event.reason

      if code == WEBSOCKET_TIMEOUT_CODE
        raise Sdk::Error,
              'Execution timed out: operation exceeded the configured `timeout`. Provide a larger value if needed.'
      end

      return if code == 1000 # Normal closure

      detail = reason.to_s.empty? ? 'WebSocket connection closed unexpectedly' : reason.to_s
      detail = "#{detail} (close code #{code})" if code
      raise Sdk::Error, detail
    end
  end
end
