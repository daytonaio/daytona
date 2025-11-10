# frozen_string_literal: true

module Daytona
  class ExecuteResponse
    # @return [Integer] The exit code from the command execution
    attr_reader :exit_code

    # @return [String] The output from the command execution
    attr_reader :result

    # @return [ExecutionArtifacts, nil] Artifacts from the command execution
    attr_reader :artifacts

    # @return [Hash] Additional properties from the response
    attr_reader :additional_properties

    # Initialize a new ExecuteResponse
    #
    # @param exit_code [Integer] The exit code from the command execution
    # @param result [String] The output from the command execution
    # @param artifacts [ExecutionArtifacts, nil] Artifacts from the command execution
    # @param additional_properties [Hash] Additional properties from the response
    def initialize(exit_code:, result:, artifacts: nil, additional_properties: {})
      @exit_code = exit_code
      @result = result
      @artifacts = artifacts
      @additional_properties = additional_properties
    end
  end

  class ExecutionArtifacts
    # @return [String] Standard output from the command, same as `result` in `ExecuteResponse`
    attr_accessor :stdout

    # @return [Array] List of chart metadata from matplotlib
    attr_accessor :charts

    # Initialize a new ExecutionArtifacts
    #
    # @param stdout [String] Standard output from the command
    # @param charts [Array] List of chart metadata from matplotlib
    def initialize(stdout = '', charts = [])
      @stdout = stdout
      @charts = charts
    end
  end

  class CodeRunParams
    # @return [Array<String>, nil] Command line arguments
    attr_accessor :argv

    # @return [Hash<String, String>, nil] Environment variables
    attr_accessor :env

    # Initialize a new CodeRunParams
    #
    # @param argv [Array<String>, nil] Command line arguments
    # @param env [Hash<String, String>, nil] Environment variables
    def initialize(argv: nil, env: nil)
      @argv = argv
      @env = env
    end
  end

  class SessionExecuteRequest
    # @return [String] The command to execute
    attr_accessor :command

    # @return [Boolean] Whether to execute the command asynchronously
    attr_accessor :run_async

    # Initialize a new SessionExecuteRequest
    #
    # @param command [String] The command to execute
    # @param run_async [Boolean] Whether to execute the command asynchronously
    def initialize(command:, run_async: false)
      @command = command
      @run_async = run_async
    end
  end

  class SessionExecuteResponse
    # @return [String, nil] Unique identifier for the executed command
    attr_reader :cmd_id

    # @return [String, nil] The output from the command execution
    attr_reader :output

    # @return [String, nil] Standard output from the command
    attr_reader :stdout

    # @return [String, nil] Standard error from the command
    attr_reader :stderr

    # @return [Integer, nil] The exit code from the command execution
    attr_reader :exit_code

    # @return [Hash] Additional properties from the response
    attr_reader :additional_properties

    # Initialize a new SessionExecuteResponse
    #
    # @param opts [Hash] Options for the SessionExecuteResponse
    # @param cmd_id [String, nil] Unique identifier for the executed command
    # @param output [String, nil] The output from the command execution
    # @param stdout [String, nil] Standard output from the command
    # @param stderr [String, nil] Standard error from the command
    # @param exit_code [Integer, nil] The exit code from the command execution
    # @param additional_properties [Hash] Additional properties from the response
    def initialize(opts = {})
      @cmd_id = opts.fetch(:cmd_id, nil)
      @output = opts.fetch(:output, nil)
      @stdout = opts.fetch(:stdout, nil)
      @stderr = opts.fetch(:stderr, nil)
      @exit_code = opts.fetch(:exit_code, nil)
      @additional_properties = opts.fetch(:additional_properties, {})
    end
  end

  class SessionCommandLogsResponse
    # @return [String, nil] The combined output from the command
    attr_reader :output

    # @return [String, nil] The stdout from the command
    attr_reader :stdout

    # @return [String, nil] The stderr from the command
    attr_reader :stderr

    # Initialize a new SessionCommandLogsResponse
    #
    # @param output [String, nil] The combined output from the command
    # @param stdout [String, nil] The stdout from the command
    # @param stderr [String, nil] The stderr from the command
    def initialize(output: nil, stdout: nil, stderr: nil)
      @output = output
      @stdout = stdout
      @stderr = stderr
    end
  end
end
