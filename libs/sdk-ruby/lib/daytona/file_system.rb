# frozen_string_literal: true

require 'tempfile'
require 'fileutils'

module Daytona
  class FileSystem
    # @return [String] The Sandbox ID
    attr_reader :sandbox_id

    # @return [DaytonaApiClient::ToolboxApi] API client for Sandbox operations
    attr_reader :toolbox_api

    # Initializes a new FileSystem instance.
    #
    # @param sandbox_id [String] The Sandbox ID
    # @param toolbox_api [DaytonaApiClient::ToolboxApi] API client for Sandbox operations
    def initialize(sandbox_id:, toolbox_api:)
      @sandbox_id = sandbox_id
      @toolbox_api = toolbox_api
    end

    # Creates a new directory in the Sandbox at the specified path with the given
    # permissions.
    #
    # @param path [String] Path where the folder should be created. Relative paths are resolved based
    #   on the sandbox working directory.
    # @param mode [String] Folder permissions in octal format (e.g., "755" for rwxr-xr-x).
    # @return [void]
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # Create a directory with standard permissions
    #   sandbox.fs.create_folder("workspace/data", "755")
    #
    #   # Create a private directory
    #   sandbox.fs.create_folder("workspace/secrets", "700")
    def create_folder(path, mode)
      Sdk.logger.debug("Creating folder #{path} with mode #{mode}")
      toolbox_api.create_folder(sandbox_id, path, mode)
    rescue StandardError => e
      raise Sdk::Error, "Failed to create folder: #{e.message}"
    end

    # Deletes a file from the Sandbox.
    #
    # @param path [String] Path to the file to delete. Relative paths are resolved based on the sandbox working directory.
    # @param recursive [Boolean] If the file is a directory, this must be true to delete it.
    # @return [void]
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # Delete a file
    #   sandbox.fs.delete_file("workspace/data/old_file.txt")
    #
    #   # Delete a directory recursively
    #   sandbox.fs.delete_file("workspace/old_dir", recursive: true)
    def delete_file(path, recursive: false)
      toolbox_api.delete_file(sandbox_id, path, { recursive: })
    rescue StandardError => e
      raise Sdk::Error, "Failed to delete file: #{e.message}"
    end

    # Gets detailed information about a file or directory, including its
    # size, permissions, and timestamps.
    #
    # @param path [String] Path to the file or directory. Relative paths are resolved based
    #   on the sandbox working directory.
    # @return [DaytonaApiClient::FileInfo] Detailed file information
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # Get file metadata
    #   info = sandbox.fs.get_file_info("workspace/data/file.txt")
    #   puts "Size: #{info.size} bytes"
    #   puts "Modified: #{info.mod_time}"
    #   puts "Mode: #{info.mode}"
    #
    #   # Check if path is a directory
    #   info = sandbox.fs.get_file_info("workspace/data")
    #   puts "Path is a directory" if info.is_dir
    def get_file_info(path)
      toolbox_api.get_file_info(sandbox_id, path)
    rescue StandardError => e
      raise Sdk::Error, "Failed to get file info: #{e.message}"
    end

    # Lists files and directories in a given path and returns their information, similar to the ls -l command.
    #
    # @param path [String] Path to the directory to list contents from. Relative paths are resolved
    #   based on the sandbox working directory.
    # @return [Array<DaytonaApiClient::FileInfo>] List of file and directory information
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # List directory contents
    #   files = sandbox.fs.list_files("workspace/data")
    #
    #   # Print files and their sizes
    #   files.each do |file|
    #     puts "#{file.name}: #{file.size} bytes" unless file.is_dir
    #   end
    #
    #   # List only directories
    #   dirs = files.select(&:is_dir)
    #   puts "Subdirectories: #{dirs.map(&:name).join(', ')}"
    def list_files(path)
      toolbox_api.list_files(sandbox_id, { path: })
    rescue StandardError => e
      raise Sdk::Error, "Failed to list files: #{e.message}"
    end

    # Downloads a file from the Sandbox. Returns the file contents as a string.
    # This method is useful when you want to load the file into memory without saving it to disk.
    # It can only be used for smaller files.
    #
    # @param remote_path [String] Path to the file in the Sandbox. Relative paths are resolved based
    #   on the sandbox working directory.
    # @param local_path [String, nil] Optional path to save the file locally. If provided, the file will be saved to disk.
    # @return [File, nil] The file if local_path is nil, otherwise nil
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # Download and get file content
    #   content = sandbox.fs.download_file("workspace/data/file.txt")
    #   puts content
    #
    #   # Download and save a file locally
    #   sandbox.fs.download_file("workspace/data/file.txt", "local_copy.txt")
    #   size_mb = File.size("local_copy.txt") / 1024.0 / 1024.0
    #   puts "Size of the downloaded file: #{size_mb} MB"
    def download_file(remote_path, local_path = nil) # rubocop:disable Metrics/MethodLength
      file = toolbox_api.download_file(sandbox_id, remote_path)

      if local_path

        parent_dir = File.dirname(local_path)
        FileUtils.mkdir_p(parent_dir) unless parent_dir == '.'

        File.binwrite(local_path, file.open.read)
        nil
      else
        file
      end
    rescue StandardError => e
      raise Sdk::Error, "Failed to download file: #{e.message}"
    end

    # Uploads a file to the specified path in the Sandbox. If a file already exists at
    # the destination path, it will be overwritten.
    #
    # @param source [String, IO] File contents as a string/bytes or a local file path or IO object.
    # @param remote_path [String] Path to the destination file. Relative paths are resolved based on
    #   the sandbox working directory.
    # @return [void]
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # Upload a text file from string content
    #   content = "Hello, World!"
    #   sandbox.fs.upload_file(content, "tmp/hello.txt")
    #
    #   # Upload a local file
    #   sandbox.fs.upload_file("local_file.txt", "tmp/file.txt")
    #
    #   # Upload binary data
    #   data = { key: "value" }.to_json
    #   sandbox.fs.upload_file(data, "tmp/config.json")
    def upload_file(source, remote_path) # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
      if source.is_a?(String) && File.exist?(source)
        # Source is a file path
        File.open(source, 'rb') { |file| toolbox_api.upload_file(sandbox_id, remote_path, { file: }) }
      elsif source.respond_to?(:read)
        # Source is an IO object
        toolbox_api.upload_file(sandbox_id, remote_path, { file: source })
      else
        # Source is string content - create a temporary file
        Tempfile.create('daytona_upload') do |file|
          file.binmode
          file.write(source)
          file.rewind
          toolbox_api.upload_file(sandbox_id, remote_path, { file: })
        end
      end
    rescue StandardError => e
      raise Sdk::Error, "Failed to upload file: #{e.message}"
    end

    # Uploads multiple files to the Sandbox. If files already exist at the destination paths,
    # they will be overwritten.
    #
    # @param files [Array<FileUpload>] List of files to upload.
    # @return [void]
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # Upload multiple files
    #   files = [
    #     FileUpload.new("Content of file 1", "/tmp/file1.txt"),
    #     FileUpload.new("workspace/data/file2.txt", "/tmp/file2.txt"),
    #     FileUpload.new('{"key": "value"}', "/tmp/config.json")
    #   ]
    #   sandbox.fs.upload_files(files)
    def upload_files(files)
      files.each { |file_upload| upload_file(file_upload.source, file_upload.destination) }
    rescue StandardError => e
      raise Sdk::Error, "Failed to upload files: #{e.message}"
    end

    # Searches for files containing a pattern, similar to the grep command.
    #
    # @param path [String] Path to the file or directory to search. If the path is a directory,
    #   the search will be performed recursively. Relative paths are resolved based
    #   on the sandbox working directory.
    # @param pattern [String] Search pattern to match against file contents.
    # @return [Array<DaytonaApiClient::Match>] List of matches found in files
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # Search for TODOs in Ruby files
    #   matches = sandbox.fs.find_files("workspace/src", "TODO:")
    #   matches.each do |match|
    #     puts "#{match.file}:#{match.line}: #{match.content.strip}"
    #   end
    def find_files(path, pattern)
      toolbox_api.find_in_files(sandbox_id, path, pattern)
    rescue StandardError => e
      raise Sdk::Error, "Failed to find files: #{e.message}"
    end

    # Searches for files and directories whose names match the specified pattern.
    # The pattern can be a simple string or a glob pattern.
    #
    # @param path [String] Path to the root directory to start search from. Relative paths are resolved
    #   based on the sandbox working directory.
    # @param pattern [String] Pattern to match against file names. Supports glob
    #   patterns (e.g., "*.rb" for Ruby files).
    # @return [DaytonaApiClient::SearchFilesResponse]
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # Find all Ruby files
    #   result = sandbox.fs.search_files("workspace", "*.rb")
    #   result.files.each { |file| puts file }
    #
    #   # Find files with specific prefix
    #   result = sandbox.fs.search_files("workspace/data", "test_*")
    #   puts "Found #{result.files.length} test files"
    def search_files(path, pattern)
      toolbox_api.search_files(sandbox_id, path, pattern)
    rescue StandardError => e
      raise Sdk::Error, "Failed to search files: #{e.message}"
    end

    # Moves or renames a file or directory. The parent directory of the destination must exist.
    #
    # @param source [String] Path to the source file or directory. Relative paths are resolved
    #   based on the sandbox working directory.
    # @param destination [String] Path to the destination. Relative paths are resolved based on
    #   the sandbox working directory.
    # @return [void]
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # Rename a file
    #   sandbox.fs.move_files(
    #     "workspace/data/old_name.txt",
    #     "workspace/data/new_name.txt"
    #   )
    #
    #   # Move a file to a different directory
    #   sandbox.fs.move_files(
    #     "workspace/data/file.txt",
    #     "workspace/archive/file.txt"
    #   )
    #
    #   # Move a directory
    #   sandbox.fs.move_files(
    #     "workspace/old_dir",
    #     "workspace/new_dir"
    #   )
    def move_files(source, destination)
      toolbox_api.move_file(sandbox_id, source, destination)
    rescue StandardError => e
      raise Sdk::Error, "Failed to move files: #{e.message}"
    end

    # Performs search and replace operations across multiple files.
    #
    # @param files [Array<String>] List of file paths to perform replacements in. Relative paths are
    #   resolved based on the sandbox working directory.
    # @param pattern [String] Pattern to search for.
    # @param new_value [String] Text to replace matches with.
    # @return [Array<DaytonaApiClient::ReplaceResult>] List of results indicating replacements made in each file
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # Replace in specific files
    #   results = sandbox.fs.replace_in_files(
    #     files: ["workspace/src/file1.rb", "workspace/src/file2.rb"],
    #     pattern: "old_function",
    #     new_value: "new_function"
    #   )
    #
    #   # Print results
    #   results.each do |result|
    #     if result.success
    #       puts "#{result.file}: #{result.success}"
    #     else
    #       puts "#{result.file}: #{result.error}"
    #     end
    #   end
    def replace_in_files(files:, pattern:, new_value:)
      replace_request = DaytonaApiClient::ReplaceRequest.new(
        files: files,
        pattern: pattern,
        new_value: new_value
      )
      toolbox_api.replace_in_files(sandbox_id, replace_request)
    rescue StandardError => e
      raise Sdk::Error, "Failed to replace in files: #{e.message}"
    end

    # Sets permissions and ownership for a file or directory. Any of the parameters can be nil
    # to leave that attribute unchanged.
    #
    # @param path [String] Path to the file or directory. Relative paths are resolved based on
    #   the sandbox working directory.
    # @param mode [String, nil] File mode/permissions in octal format (e.g., "644" for rw-r--r--).
    # @param owner [String, nil] User owner of the file.
    # @param group [String, nil] Group owner of the file.
    # @return [void]
    # @raise [Daytona::Sdk::Error] If the operation fails
    #
    # @example
    #   # Make a file executable
    #   sandbox.fs.set_file_permissions(
    #     path: "workspace/scripts/run.sh",
    #     mode: "755"  # rwxr-xr-x
    #   )
    #
    #   # Change file owner
    #   sandbox.fs.set_file_permissions(
    #     path: "workspace/data/file.txt",
    #     owner: "daytona",
    #     group: "daytona"
    #   )
    def set_file_permissions(path:, mode: nil, owner: nil, group: nil)
      opts = {}
      opts[:mode] = mode if mode
      opts[:owner] = owner if owner
      opts[:group] = group if group

      toolbox_api.set_file_permissions(sandbox_id, path, opts)
    rescue StandardError => e
      raise Sdk::Error, "Failed to set file permissions: #{e.message}"
    end
  end
end
