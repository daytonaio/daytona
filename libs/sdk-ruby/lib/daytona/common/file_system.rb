# frozen_string_literal: true

module Daytona
  class FileUpload
    # @return [String, IO] File contents as a string/bytes object or a local file path or IO object.
    #   If a string path is provided, the file content will be read from disk.
    #   If a string content is provided, make sure it fits into memory.
    attr_reader :source

    # @return [String] Absolute destination path in the Sandbox. Relative paths are resolved based on
    #   the sandbox working directory.
    attr_reader :destination

    # Initializes a new FileUpload instance.
    #
    # @param source [String, IO] File contents as a string/bytes object or a local file path or IO object.
    # @param destination [String] Absolute destination path in the Sandbox.
    def initialize(source, destination)
      @source = source
      @destination = destination
    end
  end
end
