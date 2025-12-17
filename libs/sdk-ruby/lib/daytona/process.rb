# frozen_string_literal: true

require 'base64'
require 'json'
require 'uri'

module Daytona
  class Process # rubocop:disable Metrics/ClassLength
    # @return [Daytona::SandboxPythonCodeToolbox,
    attr_reader :code_toolbox

    # @return [String] The ID of the Sandbox
    attr_reader :sandbox_id

    # @return [DaytonaToolboxApiClient::ProcessApi] API client for Sandbox operations
    attr_reader :toolbox_api

    # @return [Proc] Function to get preview link for a port
    attr_reader :get_preview_link

    # Initialize a new Process instance
    #
    # @param code_toolbox [Daytona::SandboxPythonCodeToolbox, Daytona::SandboxTsCodeToolbox]
    # @param sandbox_id [String] The ID of the Sandbox
    # @param toolbox_api [DaytonaToolboxApiClient::ProcessApi] API client for Sandbox operations
    # @param get_preview_link [Proc] Function to get preview link for a port
    def initialize(code_toolbox:, sandbox_id:, toolbox_api:, get_preview_link:)
      @code_toolbox = code_toolbox
      @sandbox_id = sandbox_id
      @toolbox_api = toolbox_api
      @get_preview_link = get_preview_link
    end

    # Execute a shell command in the Sandbox
    #
    # @param command [String] Shell command to execute
    # @param cwd [String, nil] Working directory for command execution. If not specified, uses the sandbox working directory
    # @param env [Hash<String, String>, nil] Environment variables to set for the command
    # @param timeout [Integer, nil] Maximum time in seconds to wait for the command to complete. 0 means wait indefinitely
    # @return [ExecuteResponse] Command execution results containing exit_code, result, and artifacts
    #
    # @example
    #   # Simple command
    #   response = sandbox.process.exec("echo 'Hello'")
    #   puts response.artifacts.stdout
    #   => "Hello\n"
    #
    #   # Command with working directory
    #   result = sandbox.process.exec("ls", cwd: "workspace/src")
    #
    #   # Command with timeout
    #   result = sandbox.process.exec("sleep 10", timeout: 5)
    def exec(command:, cwd: nil, env: nil, timeout: nil) # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
      command = "echo '#{Base64.encode64(command)}' | base64 -d | sh"

      if env && !env.empty?
        safe_env_exports = env.map do |key, value|
          "export #{key}=$(echo '#{Base64.encode64(value)}' | base64 -d)"
        end.join(';')
        command = "#{safe_env_exports}; #{command}"
      end

      command = "sh -c \"#{command}\""

      response = toolbox_api.execute_command(DaytonaToolboxApiClient::ExecuteRequest.new(command:, cwd:, timeout:))
      # Post-process the output to extract ExecutionArtifacts
      artifacts = parse_output(response.result.split("\n"))

      # Create new response with processed output and charts
      ExecuteResponse.new(
        exit_code: response.exit_code,
        result: artifacts.stdout,
        artifacts: artifacts
      )
    end

    # Execute code in the Sandbox using the appropriate language runtime
    #
    # @param code [String] Code to execute
    # @param params [CodeRunParams, nil] Parameters for code execution
    # @param timeout [Integer, nil] Maximum time in seconds to wait for the code to complete. 0 means wait indefinitely
    # @return [ExecuteResponse] Code execution result containing exit_code, result, and artifacts
    #
    # @example
    #   # Run Python code
    #   response = sandbox.process.code_run(<<~CODE)
    #     x = 10
    #     y = 20
    #     print(f"Sum: {x + y}")
    #   CODE
    #   puts response.artifacts.stdout  # Prints: Sum: 30
    def code_run(code:, params: nil, timeout: nil)
      exec(command: code_toolbox.get_run_command(code, params), env: params&.env, timeout:)
    end

    # Creates a new long-running background session in the Sandbox
    #
    # Sessions are background processes that maintain state between commands, making them ideal for
    # scenarios requiring multiple related commands or persistent environment setup.
    #
    # @param session_id [String] Unique identifier for the new session
    # @return [void]
    #
    # @example
    #   # Create a new session
    #   session_id = "my-session"
    #   sandbox.process.create_session(session_id)
    #   session = sandbox.process.get_session(session_id)
    #   # Do work...
    #   sandbox.process.delete_session(session_id)
    def create_session(session_id)
      toolbox_api.create_session(DaytonaToolboxApiClient::CreateSessionRequest.new(session_id:))
    end

    # Gets a session in the Sandbox
    #
    # @param session_id [String] Unique identifier of the session to retrieve
    # @return [DaytonaApiClient::Session] Session information including session_id and commands
    #
    # @example
    #   session = sandbox.process.get_session("my-session")
    #   session.commands.each do |cmd|
    #     puts "Command: #{cmd.command}"
    #   end
    def get_session(session_id) = toolbox_api.get_session(session_id)

    # Gets information about a specific command executed in a session
    #
    # @param session_id [String] Unique identifier of the session
    # @param command_id [String] Unique identifier of the command
    # @return [DaytonaApiClient::Command] Command information including id, command, and exit_code
    #
    # @example
    #   cmd = sandbox.process.get_session_command(session_id: "my-session", command_id: "cmd-123")
    #   if cmd.exit_code == 0
    #     puts "Command #{cmd.command} completed successfully"
    #   end
    def get_session_command(session_id:, command_id:)
      toolbox_api.get_session_command(session_id, command_id)
    end

    # Executes a command in the session
    #
    # @param session_id [String] Unique identifier of the session to use
    # @param req [Daytona::SessionExecuteRequest] Command execution request containing command and run_async
    # @return [Daytona::SessionExecuteResponse] Command execution results containing cmd_id, output, stdout, stderr, and exit_code
    #
    # @example
    #   # Execute commands in sequence, maintaining state
    #   session_id = "my-session"
    #
    #   # Change directory
    #   req = Daytona::SessionExecuteRequest.new(command: "cd /workspace")
    #   sandbox.process.execute_session_command(session_id:, req:)
    #
    #   # Create a file
    #   req = Daytona::SessionExecuteRequest.new(command: "echo 'Hello' > test.txt")
    #   sandbox.process.execute_session_command(session_id:, req:)
    #
    #   # Read the file
    #   req = Daytona::SessionExecuteRequest.new(command: "cat test.txt")
    #   result = sandbox.process.execute_session_command(session_id:, req:)
    #   puts "Command stdout: #{result.stdout}"
    #   puts "Command stderr: #{result.stderr}"
    def execute_session_command(session_id:, req:) # rubocop:disable Metrics/MethodLength
      response = toolbox_api.session_execute_command(
        session_id,
        DaytonaToolboxApiClient::SessionExecuteRequest.new(command: req.command, run_async: req.run_async)
      )

      stdout, stderr = Util.demux(response.output || '')

      SessionExecuteResponse.new(
        cmd_id: response.cmd_id,
        output: response.output,
        stdout:,
        stderr:,
        exit_code: response.exit_code,
        # TODO: DaytonaApiClient::SessionExecuteResponse doesn't have additional_properties attribute
        additional_properties: {}
      )
    end

    # Get the logs for a command executed in a session
    #
    # @param session_id [String] Unique identifier of the session
    # @param command_id [String] Unique identifier of the command
    # @return [Daytona::SessionCommandLogsResponse] Command logs including output, stdout, and stderr
    #
    # @example
    #   logs = sandbox.process.get_session_command_logs(session_id: "my-session", command_id: "cmd-123")
    #   puts "Command stdout: #{logs.stdout}"
    #   puts "Command stderr: #{logs.stderr}"
    def get_session_command_logs(session_id:, command_id:)
      parse_session_command_logs(
        toolbox_api.get_session_command_logs(
          session_id,
          command_id
        )
      )
    end

    # Asynchronously retrieves and processes the logs for a command executed in a session as they become available
    #
    # @param session_id [String] Unique identifier of the session
    # @param command_id [String] Unique identifier of the command
    # @param on_stdout [Proc] Callback function to handle stdout log chunks as they arrive
    # @param on_stderr [Proc] Callback function to handle stderr log chunks as they arrive
    # @return [WebSocket::Client::Simple::Client]
    #
    # @example
    #   sandbox.process.get_session_command_logs_async(
    #     session_id: "my-session",
    #     command_id: "cmd-123",
    #     on_stdout: ->(log) { puts "[STDOUT]: #{log}" },
    #     on_stderr: ->(log) { puts "[STDERR]: #{log}" }
    #   )
    def get_session_command_logs_async(session_id:, command_id:, on_stdout:, on_stderr:) # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
      preview_link = get_preview_link.call(WS_PORT)
      url = URI.parse(preview_link.url)
      url.scheme = url.scheme == 'https' ? 'wss' : 'ws'
      url.path = "/process/session/#{session_id}/command/#{command_id}/logs"
      url.query = 'follow=true'

      WebSocket::Client::Simple.connect(
        url.to_s,
        headers: toolbox_api.api_client.default_headers.dup.merge(
          'X-Daytona-Preview-Token' => preview_link.token,
          'Content-Type' => 'text/plain',
          'Accept' => 'text/plain'
        )
      ) do |ws|
        ws.on(:message) do |message|
          if message.type == :close
            ws.close
            next
          else
            stdout, stderr = Util.demux(message.data.to_s)

            on_stdout.call(stdout) unless stdout.empty?
            on_stderr.call(stderr) unless stderr.empty?
          end
        end
      end
    end

    #
    # @return [Array<DaytonaApiClient::Session>] List of all sessions in the Sandbox
    #
    # @example
    #   sessions = sandbox.process.list_sessions
    #   sessions.each do |session|
    #     puts "Session #{session.session_id}:"
    #     puts "  Commands: #{session.commands.length}"
    #   end
    def list_sessions = toolbox_api.list_sessions

    # Terminates and removes a session from the Sandbox, cleaning up any resources associated with it
    #
    # @param session_id [String] Unique identifier of the session to delete
    #
    # @example
    #   # Create and use a session
    #   sandbox.process.create_session("temp-session")
    #   # ... use the session ...
    #
    #   # Clean up when done
    #   sandbox.process.delete_session("temp-session")
    def delete_session(session_id) = toolbox_api.delete_session(session_id)

    # Creates a new PTY (pseudo-terminal) session in the Sandbox.
    #
    # Creates an interactive terminal session that can execute commands and handle user input.
    # The PTY session behaves like a real terminal, supporting features like command history.
    #
    # @param id [String] Unique identifier for the PTY session. Must be unique within the Sandbox.
    # @param cwd [String, nil] Working directory for the PTY session. Defaults to the sandbox's working directory.
    # @param envs [Hash<String, String>, nil] Environment variables to set in the PTY session. These will be merged with
    #                                        the Sandbox's default environment variables.
    # @param pty_size [PtySize, nil] Terminal size configuration. Defaults to 80x24 if not specified.
    # @return [PtyHandle] Handle for managing the created PTY session. Use this to send input,
    #                     receive output, resize the terminal, and manage the session lifecycle.
    #
    # @example
    #   # Create a basic PTY session
    #   pty_handle = sandbox.process.create_pty_session(id: "my-pty")
    #
    #   # Create a PTY session with specific size and environment
    #   pty_size = Daytona::PtySize.new(rows: 30, cols: 120)
    #   pty_handle = sandbox.process.create_pty_session(
    #     id: "my-pty",
    #     cwd: "/workspace",
    #     envs: {"NODE_ENV" => "development"},
    #     pty_size: pty_size
    #   )
    #
    #   # Use the PTY session
    #   pty_handle.wait_for_connection
    #   pty_handle.send_input("ls -la\n")
    #   result = pty_handle.wait
    #   pty_handle.disconnect
    #
    # @raise [Daytona::Sdk::Error] If the PTY session creation fails or the session ID is already in use.
    def create_pty_session(id:, cwd: nil, envs: nil, pty_size: nil) # rubocop:disable Metrics/MethodLength
      response = toolbox_api.create_pty_session(
        DaytonaToolboxApiClient::PtyCreateRequest.new(
          id:,
          cwd:,
          envs:,
          cols: pty_size&.cols,
          rows: pty_size&.rows,
          lazy_start: true
        )
      )

      connect_pty_session(response.session_id)
    end

    # Connects to an existing PTY session in the Sandbox.
    #
    # Establishes a WebSocket connection to an existing PTY session, allowing you to
    # interact with a previously created terminal session.
    #
    # @param session_id [String] Unique identifier of the PTY session to connect to.
    # @return [PtyHandle] Handle for managing the connected PTY session.
    #
    # @example
    #   # Connect to an existing PTY session
    #   pty_handle = sandbox.process.connect_pty_session("my-pty-session")
    #   pty_handle.wait_for_connection
    #   pty_handle.send_input("echo 'Hello World'\n")
    #   result = pty_handle.wait
    #   pty_handle.disconnect
    #
    # @raise [Daytona::Sdk::Error] If the PTY session doesn't exist or connection fails.
    def connect_pty_session(session_id) # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
      preview_link = get_preview_link.call(WS_PORT)
      url = URI.parse(preview_link.url)
      url.scheme = url.scheme == 'https' ? 'wss' : 'ws'
      url.path = "/process/pty/#{session_id}/connect"

      PtyHandle.new(
        WebSocket::Client::Simple.connect(
          url.to_s,
          headers: toolbox_api.api_client.default_headers.dup.merge(
            'X-Daytona-Preview-Token' => preview_link.token
          )
        ),
        session_id:,

        handle_resize: ->(pty_size) { resize_pty_session(session_id, pty_size) },
        handle_kill: -> { delete_pty_session(session_id) }
      ).tap(&:wait_for_connection)
    end

    # Resizes a PTY session to the specified dimensions
    #
    # @param session_id [String] Unique identifier of the PTY session
    # @param pty_size [PtySize] New terminal size
    # @return [DaytonaApiClient::PtySessionInfo] Updated PTY session information
    #
    # @example
    #   pty_size = Daytona::PtySize.new(rows: 30, cols: 120)
    #   session_info = sandbox.process.resize_pty_session("my-pty", pty_size)
    #   puts "PTY resized to #{session_info.cols}x#{session_info.rows}"
    def resize_pty_session(session_id, pty_size)
      toolbox_api.resize_pty_session(
        session_id,
        DaytonaToolboxApiClient::PtyResizeRequest.new(
          cols: pty_size.cols,
          rows: pty_size.rows
        )
      )
    end

    # Deletes a PTY session, terminating the associated process
    #
    # @param session_id [String] Unique identifier of the PTY session to delete
    # @return [void]
    #
    # @example
    #   sandbox.process.delete_pty_session("my-pty")
    def delete_pty_session(session_id)
      toolbox_api.delete_pty_session(session_id)
    end

    # Lists all PTY sessions in the Sandbox
    #
    # @return [Array<DaytonaApiClient::PtySessionInfo>] List of PTY session information
    #
    # @example
    #   sessions = sandbox.process.list_pty_sessions
    #   sessions.each do |session|
    #     puts "PTY Session #{session.id}: #{session.cols}x#{session.rows}"
    #   end
    def list_pty_sessions
      toolbox_api.list_pty_sessions
    end

    # Gets detailed information about a specific PTY session
    #
    # Retrieves comprehensive information about a PTY session including its current state,
    # configuration, and metadata.
    #
    # @param session_id [String] Unique identifier of the PTY session to retrieve information for
    # @return [DaytonaApiClient::PtySessionInfo] Detailed information about the PTY session including ID, state,
    #                                            creation time, working directory, environment variables, and more
    #
    # @example
    #   # Get details about a specific PTY session
    #   session_info = sandbox.process.get_pty_session_info("my-session")
    #   puts "Session ID: #{session_info.id}"
    #   puts "Active: #{session_info.active}"
    #   puts "Working Directory: #{session_info.cwd}"
    #   puts "Terminal Size: #{session_info.cols}x#{session_info.rows}"
    def get_pty_session_info(session_id)
      toolbox_api.get_pty_session(session_id)
    end

    private

    # Parse the output of a command to extract ExecutionArtifacts
    #
    # @param lines [Array<String>] A list of lines of output from a command
    # @return [Daytona::ExecutionArtifacts] The artifacts from the command execution
    def parse_output(lines)
      artifacts = ExecutionArtifacts.new('', [])

      lines.each do |line|
        if line.start_with?(ARTIFACT_PREFIX)
          parse_json_line(line:, artifacts:)
        else
          artifacts.stdout += "#{line}\n"
        end
      end

      artifacts
    end

    # Parse a JSON line to extract artifacts
    #
    # @param line [String] The line to parse
    # @param artifacts [Daytona::ExecutionArtifacts] The artifacts to add to
    # @return [void]
    def parse_json_line(line:, artifacts:)
      data = JSON.parse(line.sub(ARTIFACT_PREFIX, '').strip, symbolize_names: true)

      case data.fetch(:type, nil)
      when ArtifactType::CHART
        artifacts.charts.append(Charts.parse(data.fetch(:value, {})))
      end
    end

    # Parse combined stdout/stderr output into separate streams
    #
    # @param data [String] Combined log string with STDOUT_PREFIX and STDERR_PREFIX markers
    # @return [SessionCommandLogsResponse] Response with separated stdout and stderr
    def parse_session_command_logs(data)
      stdout, stderr = Util.demux(data)

      SessionCommandLogsResponse.new(
        output: data,
        stdout:,
        stderr:
      )
    end

    ARTIFACT_PREFIX = 'dtn_artifact_k39fd2:'
    private_constant :ARTIFACT_PREFIX

    WS_PORT = 2280
    private_constant :WS_PORT

    module ArtifactType
      ALL = [
        CHART = 'chart'
      ].freeze
    end
  end
end
