# frozen_string_literal: true

require 'json'
require 'websocket-client-simple'
require 'timeout'

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
    include Instrumentation

    WEBSOCKET_TIMEOUT_CODE = 4008
    WS_PORT = 2280
    private_constant :WS_PORT

    # @param sandbox_id [String]
    # @param toolbox_api [DaytonaToolboxApiClient::InterpreterApi]
    # @param get_preview_link [Proc]
    # @param otel_state [Daytona::OtelState, nil]
    def initialize(sandbox_id:, toolbox_api:, get_preview_link:, otel_state: nil)
      @sandbox_id = sandbox_id
      @toolbox_api = toolbox_api
      @get_preview_link = get_preview_link
      @otel_state = otel_state
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
    def run_code(code, context: nil, on_stdout: nil, on_stderr: nil, on_error: nil, envs: nil, timeout: nil) # rubocop:disable Metrics/AbcSize, Metrics/MethodLength, Metrics/ParameterLists
      # Get WebSocket URL via preview link
      preview_link = @get_preview_link.call(WS_PORT)
      url = URI.parse(preview_link.url)
      url.scheme = url.scheme == 'https' ? 'wss' : 'ws'
      url.path = '/process/interpreter/execute'
      ws_url = url.to_s

      result = ExecutionResult.new

      # Create request payload
      request = { code: }
      request[:contextId] = context.id if context
      request[:envs] = envs if envs
      request[:timeout] = timeout if timeout

      # Build headers with preview token
      headers = @toolbox_api.api_client.default_headers.dup.merge(
        'X-Daytona-Preview-Token' => preview_link.token,
        'Content-Type' => 'application/json',
        'Accept' => 'application/json'
      )

      # Use queue for synchronization
      completion_queue = Queue.new
      interpreter = self # Capture self for use in blocks
      last_message_time = Time.now
      message_mutex = Mutex.new

      puts "[DEBUG] Connecting to WebSocket: #{ws_url}" if ENV['DEBUG']

      # Connect to WebSocket and execute
      ws = WebSocket::Client::Simple.connect(ws_url, headers:)

      ws.on :open do
        puts '[DEBUG] WebSocket opened, sending request' if ENV['DEBUG']
        ws.send(JSON.dump(request))
      end

      ws.on :message do |msg|
        message_mutex.synchronize { last_message_time = Time.now }

        puts "[DEBUG] Received message (length=#{msg.data.length}): #{msg.data.inspect[0..200]}" if ENV['DEBUG']

        interpreter.send(:handle_message, msg.data, result, on_stdout, on_stderr, on_error, completion_queue)
      end

      ws.on :error do |e|
        puts "[DEBUG] WebSocket error: #{e.message}" if ENV['DEBUG']
        completion_queue.push({ type: :error, error: e })
      end

      ws.on :close do |e|
        if ENV['DEBUG']
          code = e&.code || 'nil'
          reason = e&.reason || 'nil'
          puts "[DEBUG] WebSocket closed: code=#{code}, reason=#{reason}"
        end
        error_info = interpreter.send(:handle_close, e)
        if error_info
          completion_queue.push({ type: :error_from_close, error: error_info })
        else
          completion_queue.push({ type: :close })
        end
      end

      # Wait for completion signal with idle timeout
      # If timeout is specified, wait longer to detect actual timeout errors
      # Otherwise use short idle timeout for normal completion
      idle_timeout = timeout ? (timeout + 2.0) : 1.0
      max_wait = (timeout || 300) + 3 # Add buffer to configured timeout
      start_time = Time.now
      completion_reason = nil

      # Wait for completion or close event
      loop do
        begin
          completion = completion_queue.pop(true) # non-blocking
          puts "[DEBUG] Got completion signal: #{completion[:type]}" if ENV['DEBUG']

          # Control message (completed/interrupted) = normal completion
          if completion[:type] == :completed
            completion_reason = :completed
            break
          # If it's an error from close event (like timeout), raise it
          elsif completion[:type] == :error_from_close
            error_msg = completion[:error]
            # Raise TimeoutError for timeout cases, regular Error for others
            if error_msg.include?('timed out') || error_msg.include?('Execution timed out')
              raise Sdk::TimeoutError, error_msg
            end

            raise Sdk::Error, error_msg

          # Close event during execution (before control message) = likely timeout or error
          elsif completion[:type] == :close
            elapsed = Time.now - start_time
            # If we got close near the timeout, it's likely a timeout
            if timeout && elapsed >= timeout && elapsed < (timeout + 2)
              raise Sdk::TimeoutError,
                    'Execution timed out: operation exceeded the configured `timeout`. Provide a larger value if needed.'
            end
            # Otherwise normal close
            completion_reason = :close
            break
          # WebSocket errors
          elsif completion[:type] == :error && !completion[:error].message.include?('stream closed')
            raise Sdk::Error, "WebSocket error: #{completion[:error].message}"
          end
        rescue ThreadError
          # Queue is empty, check idle timeout
        end

        # Check idle timeout (no messages for N seconds = completion)
        time_since_last_message = message_mutex.synchronize { Time.now - last_message_time }
        if time_since_last_message > idle_timeout
          puts "[DEBUG] Idle timeout reached (#{idle_timeout}s), assuming completion" if ENV['DEBUG']
          completion_reason = :idle_complete
          break
        end

        # Check for absolute timeout (safety net)
        if Time.now - start_time > max_wait
          ws.close
          raise Sdk::TimeoutError,
                'Execution timed out: operation exceeded the configured `timeout`. Provide a larger value if needed.'
        end

        sleep 0.05 # Check every 50ms
      end

      # Close WebSocket if not already closed
      ws.close if completion_reason != :close
      sleep 0.05

      result
    rescue Sdk::Error
      # Re-raise SDK errors as-is
      raise
    rescue StandardError => e
      # Wrap unexpected errors
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

    instrument :run_code, :create_context, :list_contexts, :delete_context,
               component: 'CodeInterpreter'

    private

    # @return [Daytona::OtelState, nil]
    attr_reader :otel_state

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
    # @param completion_queue [Queue, nil] Queue to signal completion
    # @return [void]
    def handle_message(data, result, on_stdout, on_stderr, on_error, completion_queue = nil) # rubocop:disable Metrics/AbcSize, Metrics/ParameterLists
      # Empty messages are just keepalives or noise, ignore them
      if data.nil? || data.empty?
        puts '[DEBUG] Received empty message, ignoring' if ENV['DEBUG']
        return
      end

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
      when 'control'
        control_text = chunk['text'] || ''
        if %w[completed interrupted].include?(control_text)
          puts "[DEBUG] Received control message: #{control_text}" if ENV['DEBUG']
          completion_queue&.push({ type: :completed })
        end
      end
    rescue JSON::ParserError => e
      # Skip malformed messages
      warn "Warning: Failed to parse message: #{e.message}" if ENV['DEBUG']
    end

    # @param event [Object]
    # @return [void]
    def handle_close(event)
      return nil unless event # Skip if event is nil (manual close)

      code = event.respond_to?(:code) ? event.code : nil
      reason = event.respond_to?(:reason) ? event.reason : nil

      if code == WEBSOCKET_TIMEOUT_CODE
        return 'Execution timed out: operation exceeded the configured `timeout`. Provide a larger value if needed.'
      end

      return nil if code == 1000 || code.nil? # Normal closure or no code

      detail = reason.to_s.empty? ? 'WebSocket connection closed unexpectedly' : reason.to_s
      detail = "#{detail} (close code #{code})" if code
      detail
    end
  end
end
