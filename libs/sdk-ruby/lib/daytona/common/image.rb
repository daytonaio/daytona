# frozen_string_literal: true

require 'digest'
require 'fileutils'
require 'pathname'
require 'shellwords'

module Daytona
  class Context
    attr_reader :source_path
    attr_reader :archive_path

    # @param source_path [String] The path to the source file or directory
    # @param archive_path [String, nil] The path inside the archive file in object storage
    def initialize(source_path:, archive_path: nil)
      @source_path = source_path
      @archive_path = archive_path
    end
  end

  # Represents an image definition for a Daytona sandbox.
  # Do not construct this class directly. Instead use one of its static factory methods,
  # such as `Image.base()`, `Image.debian_slim()`, or `Image.from_dockerfile()`.
  class Image # rubocop:disable Metrics/ClassLength
    # @return [String, nil] The generated Dockerfile for the image
    attr_reader :dockerfile

    # @return [Array<Context>] List of context files for the image
    attr_reader :context_list

    # Supported Python series
    SUPPORTED_PYTHON_SERIES = %w[3.9 3.10 3.11 3.12 3.13].freeze
    LATEST_PYTHON_MICRO_VERSIONS = %w[3.9.22 3.10.17 3.11.12 3.12.10 3.13.3].freeze

    # @param dockerfile [String, nil] The Dockerfile content
    # @param context_list [Array<Context>] List of context files
    def initialize(dockerfile: nil, context_list: [])
      @dockerfile = dockerfile || ''
      @context_list = context_list
    end

    # Adds commands to install packages using pip
    #
    # @param packages [Array<String>] The packages to install
    # @param find_links [Array<String>, nil] The find-links to use
    # @param index_url [String, nil] The index URL to use
    # @param extra_index_urls [Array<String>, nil] The extra index URLs to use
    # @param pre [Boolean] Whether to install pre-release packages
    # @param extra_options [String] Additional options to pass to pip
    # @return [Image] The image with the pip install commands added
    #
    # @example
    #   image = Image.debian_slim("3.12").pip_install("requests", "pandas")
    def pip_install(*packages, find_links: nil, index_url: nil, extra_index_urls: nil, pre: false, extra_options: '') # rubocop:disable Metrics/ParameterLists
      pkgs = flatten_str_args('pip_install', 'packages', packages)
      return self if pkgs.empty?

      extra_args = format_pip_install_args(find_links:, index_url:, extra_index_urls:, pre:, extra_options:)
      @dockerfile += "RUN python -m pip install #{Shellwords.join(pkgs.sort)}#{extra_args}\n"

      self
    end

    # Installs dependencies from a requirements.txt file
    #
    # @param requirements_txt [String] The path to the requirements.txt file
    # @param find_links [Array<String>, nil] The find-links to use
    # @param index_url [String, nil] The index URL to use
    # @param extra_index_urls [Array<String>, nil] The extra index URLs to use
    # @param pre [Boolean] Whether to install pre-release packages
    # @param extra_options [String] Additional options to pass to pip
    # @return [Image] The image with the pip install commands added
    # @raise [Sdk::Error] If the requirements file does not exist
    #
    # @example
    #   image = Image.debian_slim("3.12").pip_install_from_requirements("requirements.txt")
    def pip_install_from_requirements(requirements_txt, find_links: nil, index_url: nil, extra_index_urls: nil, # rubocop:disable Metrics/ParameterLists
                                      pre: false, extra_options: '')
      requirements_txt = File.expand_path(requirements_txt)
      raise Sdk::Error, "Requirements file #{requirements_txt} does not exist" unless File.exist?(requirements_txt)

      extra_args = format_pip_install_args(find_links:, index_url:, extra_index_urls:, pre:, extra_options:)

      archive_path = ObjectStorage.compute_archive_base_path(requirements_txt)
      @context_list << Context.new(source_path: requirements_txt, archive_path:)
      @dockerfile += "COPY #{archive_path} /.requirements.txt\n"
      @dockerfile += "RUN python -m pip install -r /.requirements.txt#{extra_args}\n"

      self
    end

    # Installs dependencies from a pyproject.toml file
    #
    # @param pyproject_toml [String] The path to the pyproject.toml file
    # @param optional_dependencies [Array<String>] The optional dependencies to install
    # @param find_links [String, nil] The find-links to use
    # @param index_url [String, nil] The index URL to use
    # @param extra_index_url [String, nil] The extra index URL to use
    # @param pre [Boolean] Whether to install pre-release packages
    # @param extra_options [String] Additional options to pass to pip
    # @return [Image] The image with the pip install commands added
    # @raise [Sdk::Error] If pyproject.toml parsing is not supported
    #
    # @example
    #   image = Image.debian_slim("3.12").pip_install_from_pyproject("pyproject.toml", optional_dependencies: ["dev"])
    def pip_install_from_pyproject(pyproject_toml, optional_dependencies: [], find_links: nil, index_url: nil, # rubocop:disable Metrics/MethodLength, Metrics/ParameterLists
                                   extra_index_url: nil, pre: false, extra_options: '')
      data = TOML.load_file(pyproject_toml)
      dependencies = data.dig('project', 'dependencies')

      unless dependencies
        raise Sdk::Error, 'No [project.dependencies] section in pyproject.toml file. ' \
                          'See https://packaging.python.org/en/latest/guides/writing-pyproject-toml ' \
                          'for further file format guidelines.'
      end

      return unless optional_dependencies

      optionals = data.dig('project', 'optional-dependencies')
      optional_dependencies.each do |group|
        dependencies.concat(optionals.fetch(group, []))
      end

      pip_install(*dependencies, find_links:, index_url:, extra_index_urls: extra_index_url, pre:, extra_options:)
    end

    # Adds a local file to the image
    #
    # @param local_path [String] The path to the local file
    # @param remote_path [String] The path to the file in the image
    # @return [Image] The image with the local file added
    #
    # @example
    #   image = Image.debian_slim("3.12").add_local_file("package.json", "/home/daytona/package.json")
    def add_local_file(local_path, remote_path)
      remote_path = "#{remote_path}/#{File.basename(local_path)}" if remote_path.end_with?('/')

      local_path = File.expand_path(local_path)
      archive_path = ObjectStorage.compute_archive_base_path(local_path)
      @context_list << Context.new(source_path: local_path, archive_path: archive_path)
      @dockerfile += "COPY #{archive_path} #{remote_path}\n"

      self
    end

    # Adds a local directory to the image
    #
    # @param local_path [String] The path to the local directory
    # @param remote_path [String] The path to the directory in the image
    # @return [Image] The image with the local directory added
    #
    # @example
    #   image = Image.debian_slim("3.12").add_local_dir("src", "/home/daytona/src")
    def add_local_dir(local_path, remote_path)
      local_path = File.expand_path(local_path)
      archive_path = ObjectStorage.compute_archive_base_path(local_path)
      @context_list << Context.new(source_path: local_path, archive_path: archive_path)
      @dockerfile += "COPY #{archive_path} #{remote_path}\n"

      self
    end

    # Runs commands in the image
    #
    # @param commands [Array<String>] The commands to run
    # @return [Image] The image with the commands added
    #
    # @example
    #   image = Image.debian_slim("3.12").run_commands('echo "Hello, world!"', 'echo "Hello again!"')
    def run_commands(*commands)
      commands.each do |command|
        if command.is_a?(Array)
          escaped = command.map { |c| c.gsub('"', '\\"').gsub("'", "\\'") }
          @dockerfile += "RUN #{escaped.map { |c| "\"#{c}\"" }.join(' ')}\n"
        else
          @dockerfile += "RUN #{command}\n"
        end
      end

      self
    end

    # Sets environment variables in the image
    #
    # @param env_vars [Hash<String, String>] The environment variables to set
    # @return [Image] The image with the environment variables added
    #
    # @example
    #   image = Image.debian_slim("3.12").env({"PROJECT_ROOT" => "/home/daytona"})
    def env(env_vars)
      non_str_keys = env_vars.reject { |_key, val| val.is_a?(String) }.keys
      raise Sdk::Error, "Image ENV variables must be strings. Invalid keys: #{non_str_keys}" unless non_str_keys.empty?

      env_vars.each do |key, val|
        @dockerfile += "ENV #{key}=#{Shellwords.escape(val)}\n"
      end

      self
    end

    # Sets the working directory in the image
    #
    # @param path [String] The path to the working directory
    # @return [Image] The image with the working directory added
    #
    # @example
    #   image = Image.debian_slim("3.12").workdir("/home/daytona")
    def workdir(path)
      @dockerfile += "WORKDIR #{Shellwords.escape(path.to_s)}\n"
      self
    end

    # Sets the entrypoint for the image
    #
    # @param entrypoint_commands [Array<String>] The commands to set as the entrypoint
    # @return [Image] The image with the entrypoint added
    #
    # @example
    #   image = Image.debian_slim("3.12").entrypoint(["/bin/bash"])
    def entrypoint(entrypoint_commands)
      unless entrypoint_commands.is_a?(Array) && entrypoint_commands.all? { |x| x.is_a?(String) }
        raise Sdk::Error, 'entrypoint_commands must be a list of strings.'
      end

      args_str = flatten_str_args('entrypoint', 'entrypoint_commands', entrypoint_commands)
      args_str = args_str.map { |arg| "\"#{arg}\"" }.join(', ') if args_str.any?
      @dockerfile += "ENTRYPOINT [#{args_str}]\n"

      self
    end

    # Sets the default command for the image
    #
    # @param cmd [Array<String>] The commands to set as the default command
    # @return [Image] The image with the default command added
    #
    # @example
    #   image = Image.debian_slim("3.12").cmd(["/bin/bash"])
    def cmd(cmd)
      unless cmd.is_a?(Array) && cmd.all? { |x| x.is_a?(String) }
        raise Sdk::Error, 'Image CMD must be a list of strings.'
      end

      cmd_str = flatten_str_args('cmd', 'cmd', cmd)
      cmd_str = cmd_str.map { |arg| "\"#{arg}\"" }.join(', ') if cmd_str.any?
      @dockerfile += "CMD [#{cmd_str}]\n"
      self
    end

    # Adds arbitrary Dockerfile-like commands to the image
    #
    # @param dockerfile_commands [Array<String>] The commands to add to the Dockerfile
    # @param context_dir [String, nil] The path to the context directory
    # @return [Image] The image with the Dockerfile commands added
    #
    # @example
    #   image = Image.debian_slim("3.12").dockerfile_commands(["RUN echo 'Hello, world!'"])
    def dockerfile_commands(dockerfile_commands, context_dir: nil) # rubocop:disable Metrics/MethodLength
      if context_dir
        context_dir = File.expand_path(context_dir)
        raise Sdk::Error, "Context directory #{context_dir} does not exist" unless Dir.exist?(context_dir)
      end

      # Extract copy sources from dockerfile commands
      extract_copy_sources(dockerfile_commands.join("\n"), context_dir || '').each do |context_path, original_path|
        archive_base_path = context_path
        if context_dir && !original_path.start_with?(context_dir)
          archive_base_path = context_path.delete_prefix(context_dir)
        end
        @context_list << Context.new(source_path: context_path, archive_path: archive_base_path)
      end

      @dockerfile += "#{dockerfile_commands.join("\n")}\n"
      self
    end

    class << self
      # Creates an Image from an existing Dockerfile
      #
      # @param path [String] The path to the Dockerfile
      # @return [Image] The image with the Dockerfile added
      #
      # @example
      #   image = Image.from_dockerfile("Dockerfile")
      def from_dockerfile(path) # rubocop:disable Metrics/AbcSize
        path = Pathname.new(File.expand_path(path))
        dockerfile = path.read
        img = new(dockerfile: dockerfile)

        # Remove dockerfile filename from path
        path_prefix = path.to_s.delete_suffix(path.basename.to_s)

        extract_copy_sources(dockerfile, path_prefix).each do |context_path, original_path|
          archive_base_path = context_path
          archive_base_path = context_path.delete_prefix(path_prefix) unless original_path.start_with?(path_prefix)
          img.context_list << Context.new(source_path: context_path, archive_path: archive_base_path)
        end

        img
      end

      # Creates an Image from an existing base image
      #
      # @param image [String] The base image to use
      # @return [Image] The image with the base image added
      #
      # @example
      #   image = Image.base("python:3.12-slim-bookworm")
      def base(image)
        img = new
        img.instance_variable_set(:@dockerfile, "FROM #{image}\n")
        img
      end

      # Creates a Debian slim image based on the official Python Docker image
      #
      # @param python_version [String, nil] The Python version to use
      # @return [Image] The image with the Debian slim image added
      #
      # @example
      #   image = Image.debian_slim("3.12")
      def debian_slim(python_version = nil) # rubocop:disable Metrics/MethodLength
        python_version = process_python_version(python_version)
        img = new
        commands = [
          "FROM python:#{python_version}-slim-bookworm",
          'RUN apt-get update',
          'RUN apt-get install -y gcc gfortran build-essential',
          'RUN pip install --upgrade pip',
          # Set debian front-end to non-interactive to avoid users getting stuck with input prompts.
          "RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections"
        ]
        img.instance_variable_set(:@dockerfile, "#{commands.join("\n")}\n")
        img
      end

      private

      # Processes the Python version
      #
      # @param python_version [String, nil] The Python version to process
      # @param allow_micro_granularity [Boolean] Whether to allow micro-level granularity
      # @return [String] The processed Python version
      def process_python_version(python_version = nil)
        python_version ||= SUPPORTED_PYTHON_SERIES.last

        unless SUPPORTED_PYTHON_SERIES.include?(python_version)
          raise Sdk::Error, "Unsupported Python version: #{python_version}"
        end

        LATEST_PYTHON_MICRO_VERSIONS.select { |v| v.start_with?(python_version) }.last
      end

      # Extracts source files from COPY commands in a Dockerfile
      #
      # @param dockerfile_content [String] The content of the Dockerfile
      # @param path_prefix [String] The path prefix to use for the sources
      # @return [Array<Array<String>>] The list of the actual file path and its corresponding COPY-command source path
      def extract_copy_sources(dockerfile_content, path_prefix = '') # rubocop:disable Metrics/AbcSize, Metrics/CyclomaticComplexity, Metrics/PerceivedComplexity, Metrics/MethodLength
        sources = []
        lines = dockerfile_content.split("\n")

        lines.each do |line|
          # Skip empty lines and comments
          next if line.strip.empty? || line.strip.start_with?('#')

          # Check if the line contains a COPY command (at the beginning of the line)
          next unless line.match?(/^\s*COPY\s+(?!.*--from=)/i)

          # Extract the sources from the COPY command
          command_parts = parse_copy_command(line)
          next unless command_parts

          # Get source paths from the parsed command parts
          command_parts['sources'].each do |source|
            # Handle absolute and relative paths differently
            full_path_pattern = if Pathname.new(source).absolute?
                                  # Absolute path - use as is
                                  source
                                else
                                  # Relative path - add prefix
                                  File.join(path_prefix, source)
                                end

            # Handle glob patterns
            matching_files = Dir.glob(full_path_pattern)

            if matching_files.any?
              matching_files.each { |matching_file| sources << [matching_file, source] }
            else
              # If no files match, include the pattern anyway
              sources << [full_path_pattern, source]
            end
          end
        end

        sources
      end

      # Parses a COPY command to extract sources and destination
      #
      # @param line [String] The line to parse
      # @return [Hash, nil] A hash containing the sources and destination, or nil if parsing fails
      def parse_copy_command(line) # rubocop:disable Metrics/AbcSize, Metrics/CyclomaticComplexity, Metrics/PerceivedComplexity, Metrics/MethodLength
        # Remove initial "COPY" and strip whitespace
        parts = line.strip[4..].strip

        # Handle JSON array format: COPY ["src1", "src2", "dest"]
        if parts.start_with?('[')
          begin
            # Parse the JSON-like array format
            elements = Shellwords.split(parts.delete('[]'))
            return nil if elements.length < 2

            { 'sources' => elements[0..-2], 'dest' => elements[-1] }
          rescue StandardError
            nil
          end
        end

        # Handle regular format with possible flags
        parts = Shellwords.split(parts)

        # Extract flags like --chown, --chmod, --from
        sources_start_idx = 0
        parts.each_with_index do |part, i|
          break unless part.start_with?('--')

          # Skip the flag and its value if it has one
          sources_start_idx = if !part.include?('=') && i + 1 < parts.length && !parts[i + 1].start_with?('--')
                                i + 2
                              else
                                i + 1
                              end
        end

        # After skipping flags, we need at least one source and one destination
        return nil if parts.length - sources_start_idx < 2

        { 'sources' => parts[sources_start_idx..-2], 'dest' => parts[-1] }
      end
    end

    private

    # Flattens a list of strings and arrays of strings into a single array of strings
    #
    # @param function_name [String] The name of the function that is being called
    # @param arg_name [String] The name of the argument that is being passed
    # @param args [Array] The list of arguments to flatten
    # @return [Array<String>] A list of strings
    def flatten_str_args(function_name, arg_name, args) # rubocop:disable Metrics/MethodLength
      ret = []
      args.each do |x|
        case x
        when String
          ret << x
        when Array
          unless x.all? { |y| y.is_a?(String) }
            raise Sdk::Error, "#{function_name}: #{arg_name} must only contain strings"
          end

          ret.concat(x)

        else
          raise Sdk::Error, "#{function_name}: #{arg_name} must only contain strings"
        end
      end
      ret
    end

    # Formats the arguments in a single string
    #
    # @param find_links [Array<String>, nil] The find-links to use
    # @param index_url [String, nil] The index URL to use
    # @param extra_index_urls [Array<String>, nil] The extra index URLs to use
    # @param pre [Boolean] Whether to install pre-release packages
    # @param extra_options [String] Additional options to pass to pip
    # @return [String] The formatted arguments
    def format_pip_install_args(find_links: nil, index_url: nil, extra_index_urls: nil, pre: false, extra_options: '') # rubocop:disable Metrics/CyclomaticComplexity, Metrics/PerceivedComplexity
      extra_args = ''
      find_links&.each { |find_link| extra_args += " --find-links #{Shellwords.escape(find_link)}" }
      extra_args += " --index-url #{Shellwords.escape(index_url)}" if index_url
      extra_index_urls&.each do |extra_index_url|
        extra_args += " --extra-index-url #{Shellwords.escape(extra_index_url)}"
      end
      extra_args += ' --pre' if pre
      extra_args += " #{extra_options.strip}" if extra_options && !extra_options.strip.empty?

      extra_args
    end
  end
end
