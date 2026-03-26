# frozen_string_literal: true

module Daytona
  # Represents stdout or stderr output from code execution
  class OutputMessage
    # @return [String] The output content
    attr_reader :output

    # @param output [String]
    def initialize(output:)
      @output = output
    end
  end

  # Represents an error that occurred during code execution
  class ExecutionError
    # @return [String] The error type/class name (e.g., "ValueError", "SyntaxError")
    attr_reader :name

    # @return [String] The error value
    attr_reader :value

    # @return [String] Full traceback of the error
    attr_reader :traceback

    # @param name [String]
    # @param value [String]
    # @param traceback [String]
    def initialize(name:, value:, traceback: '')
      @name = name
      @value = value
      @traceback = traceback
    end
  end

  # Result of code execution
  class ExecutionResult
    # @return [String] Standard output from the code execution
    attr_accessor :stdout

    # @return [String] Standard error output from the code execution
    attr_accessor :stderr

    # @return [ExecutionError, nil] Error details if execution failed, nil otherwise
    attr_accessor :error

    def initialize(stdout: '', stderr: '', error: nil)
      @stdout = stdout
      @stderr = stderr
      @error = error
    end
  end
end
