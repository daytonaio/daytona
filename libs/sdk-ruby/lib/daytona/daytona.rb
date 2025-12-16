# frozen_string_literal: true

require 'json'
require 'uri'

module Daytona
  class Daytona # rubocop:disable Metrics/ClassLength
    # @return [Daytona::Config]
    attr_reader :config

    # @return [DaytonaApiClient]
    attr_reader :api_client

    # @return [DaytonaApiClient::SandboxApi]
    attr_reader :sandbox_api

    # @return [DaytonaApiClient::ToolboxApi]
    attr_reader :toolbox_api

    # @return [Daytona::VolumeService]
    attr_reader :volume

    # @return [DaytonaApiClient::ObjectStorageApi]
    attr_reader :object_storage_api

    # @return [DaytonaApiClient::SnapshotsApi]
    attr_reader :snapshots_api

    # @return [Daytona::SnapshotService]
    attr_reader :snapshot

    # @param config [Daytona::Config] Configuration options. Defaults to Daytona::Config.new
    def initialize(config = Config.new) # rubocop:disable Metrics/AbcSize
      @config = config
      ensure_access_token_defined
      @api_client = build_api_client
      @sandbox_api = DaytonaApiClient::SandboxApi.new(api_client)
      @toolbox_api = DaytonaApiClient::ToolboxApi.new(api_client)
      @volume = VolumeService.new(DaytonaApiClient::VolumesApi.new(api_client))
      @object_storage_api = DaytonaApiClient::ObjectStorageApi.new(api_client)
      @snapshots_api = DaytonaApiClient::SnapshotsApi.new(api_client)
      @snapshot = SnapshotService.new(snapshots_api:, object_storage_api:)
    end

    # Creates a sandbox with the specified parameters
    #
    # @param params [Daytona::CreateSandboxFromSnapshotParams, Daytona::CreateSandboxFromImageParams, Nil] Sandbox creation parameters
    # @return [Daytona::Sandbox] The created sandbox
    # @raise [Daytona::Sdk::Error] If auto_stop_interval or auto_archive_interval is negative
    def create(params = nil, on_snapshot_create_logs: nil)
      if params.nil?
        params = CreateSandboxFromSnapshotParams.new(language: CodeLanguage::PYTHON)
      elsif params.language.nil?
        params.language = CodeLanguage::PYTHON
      end

      _create(params, on_snapshot_create_logs:)
    end

    # Deletes a Sandbox.
    #
    # @param sandbox [Daytona::Sandbox]
    # @return [void]
    def delete(sandbox) = sandbox.delete

    # Gets a Sandbox by its ID.
    #
    # @param id [String]
    # @return [Daytona::Sandbox]
    def get(id)
      sandbox_dto = sandbox_api.get_sandbox(id)
      to_sandbox(sandbox_dto:, code_toolbox: code_toolbox_from_labels(sandbox_dto.labels))
    end

    # Finds a Sandbox by its ID or labels.
    #
    # @param id [String, Nil]
    # @param labels [Hash<String, String>]
    # @return [Daytona::Sandbox]
    # @raise [Daytona::Sdk::Error]
    def find_one(id: nil, labels: nil)
      return get(id) if id

      response = list(labels)
      raise Sdk::Error, "No sandbox found with labels #{labels}" if response.items.empty?

      response.items.first
    end

    # Lists Sandboxes filtered by labels.
    #
    # @param labels [Hash<String, String>]
    # @param page [Integer, Nil]
    # @param limit [Integer, Nil]
    # @return [Daytona::PaginatedResource]
    # @raise [Daytona::Sdk::Error]
    def list(labels = {}, page: nil, limit: nil) # rubocop:disable Metrics/MethodLength
      raise Sdk::Error, 'page must be positive integer' if page && page < 1

      raise Sdk::Error, 'limit must be positive integer' if limit && limit < 1

      response = sandbox_api.list_sandboxes_paginated(labels: JSON.dump(labels), page:, limit:)

      PaginatedResource.new(
        total: response.total,
        page: response.page,
        total_pages: response.total_pages,
        items: response
          .items
          .map { |sandbox_dto| to_sandbox(sandbox_dto:, code_toolbox: code_toolbox_from_labels(sandbox_dto.labels)) }
      )
    end

    # Starts a Sandbox and waits for it to be ready.
    #
    # @param sandbox [Daytona::Sandbox]
    # @param timeout [Numeric] Maximum wait time in seconds (defaults to 60 s).
    # @return [void]
    def start(sandbox, timeout = Sandbox::DEFAULT_TIMEOUT) = sandbox.start(timeout)

    # Stops a Sandbox and waits for it to be stopped.
    #
    # @param sandbox [Daytona::Sandbox]
    # @param timeout [Numeric] Maximum wait time in seconds (defaults to 60 s).
    # @return [void]
    def stop(sandbox, timeout = Sandbox::DEFAULT_TIMEOUT) = sandbox.stop(timeout)

    private

    # Creates a sandbox with the specified parameters
    #
    # @param params [Daytona::CreateSandboxFromSnapshotParams, Daytona::CreateSandboxFromImageParams] Sandbox creation parameters
    # @param timeout [Numeric] Maximum wait time in seconds (defaults to 60 s).
    # @param on_snapshot_create_logs [Proc]
    # @return [Daytona::Sandbox] The created sandbox
    # @raise [Daytona::Sdk::Error] If auto_stop_interval or auto_archive_interval is negative
    def _create(params, timeout: 60, on_snapshot_create_logs: nil) # rubocop:disable Metrics/AbcSize, Metrics/CyclomaticComplexity, Metrics/MethodLength, Metrics/PerceivedComplexity
      raise Sdk::Error, 'Timeout must be a non-negative number' if timeout.negative?

      start_time = Time.now

      raise Sdk::Error, 'auto_stop_interval must be a non-negative integer' if params.auto_stop_interval&.negative?

      if params.auto_archive_interval&.negative?
        raise Sdk::Error, 'auto_archive_interval must be a non-negative integer'
      end

      create_sandbox = DaytonaApiClient::CreateSandbox.new(
        user: params.os_user,
        env: params.env_vars || {},
        labels: params.labels,
        public: params.public,
        target: config.target,
        auto_stop_interval: params.auto_stop_interval,
        auto_archive_interval: params.auto_archive_interval,
        auto_delete_interval: params.auto_delete_interval,
        volumes: params.volumes,
        network_block_all: params.network_block_all,
        network_allow_list: params.network_allow_list
      )

      create_sandbox.snapshot = params.snapshot if params.respond_to?(:snapshot)

      if params.respond_to?(:image) && params.image.is_a?(String)
        create_sandbox.build_info = DaytonaApiClient::CreateBuildInfo.new(
          dockerfile_content: Image.base(params.image).dockerfile
        )
      elsif params.respond_to?(:image) && params.image.is_a?(Image)
        create_sandbox.build_info = DaytonaApiClient::CreateBuildInfo.new(
          context_hashes: SnapshotService.process_image_context(object_storage_api, params.image),
          dockerfile_content: params.image.dockerfile
        )
      end

      if params.respond_to?(:resources)
        create_sandbox.cpu = params.resources&.cpu
        create_sandbox.memory = params.resources&.memory
        create_sandbox.disk = params.resources&.disk
        create_sandbox.gpu = params.resources&.gpu
      end

      response = sandbox_api.create_sandbox(create_sandbox)

      if response.state == DaytonaApiClient::SandboxState::PENDING_BUILD && on_snapshot_create_logs
        uri = URI.parse(sandbox_api.api_client.config.base_url)
        uri.path = "/api/sandbox/#{response.id}/build-logs"
        uri.query = 'follow=true'

        headers = {}
        sandbox_api.api_client.update_params_for_auth!(headers, nil, ['bearer'])
        Util.stream_async(uri:, headers:, on_chunk: on_snapshot_create_logs)
      end

      sandbox = to_sandbox(sandbox_dto: response, code_toolbox: code_toolbox_from_labels(response.labels))

      if sandbox.state != DaytonaApiClient::SandboxState::STARTED
        sandbox.wait_for_sandbox_start([0.001, timeout - (Time.now - start_time)].max)
      end

      sandbox
    end

    # @return [void]
    # @raise [Daytona::Sdk::Error]
    def ensure_access_token_defined
      return if config.api_key

      raise Sdk::Error, 'API key or JWT token is required' unless config.jwt_token
      raise Sdk::Error, 'Organization ID is required when using JWT token' unless config.organization_id
    end

    # @return [DaytonaApiClient::ApiClient]
    def build_api_client
      DaytonaApiClient::ApiClient.new(api_client_config).tap do |client|
        client.default_headers[HEADER_SOURCE] = SOURCE_RUBY
        client.default_headers[HEADER_SDK_VERSION] = Sdk::VERSION
        client.default_headers[HEADER_ORGANIZATION_ID] = config.organization_id if config.jwt_token
      end
    end

    # @return [DaytonaApiClient::Configuration]
    def api_client_config
      DaytonaApiClient::Configuration.new.configure do |api_config|
        uri = URI(config.api_url)
        api_config.scheme = uri.scheme
        api_config.host = uri.host
        api_config.base_path = uri.path

        api_config.access_token_getter = proc { config.api_key || config.jwt_token }
        api_config
      end
    end

    # @param sandbox_dto [DaytonaApiClient::Sandbox]
    # @param code_toolbox [Daytona::SandboxPythonCodeToolbox, Daytona::SandboxTsCodeToolbox]
    # @return [Daytona::Sandbox]
    def to_sandbox(sandbox_dto:, code_toolbox:)
      Sandbox.new(sandbox_dto:, config:, sandbox_api:, toolbox_api:, code_toolbox:)
    end

    # Converts a language to a code toolbox
    #
    # @param language [Symbol]
    # @return [Daytona::CodeToolbox]
    # @raise [Daytona::Sdk::Error] If the language is not supported
    def code_toolbox_from_language(language)
      case language
      when CodeLanguage::PYTHON, nil
        SandboxPythonCodeToolbox.new
      when SandboxTsCodeToolbox, CodeLanguage::TYPESCRIPT
        SandboxTsCodeToolbox.new
      else
        raise Sdk::Error, "Unsupported language: #{language}"
      end
    end

    # Get code toolbox from Sandbox labels
    #
    # @param labels [Hash<String, String>]
    # @return [Daytona::CodeToolbox]
    def code_toolbox_from_labels(labels) = code_toolbox_from_language(labels[LABEL_CODE_TOOLBOX_LANGUAGE]&.to_sym)

    SOURCE_RUBY = 'ruby-sdk'
    private_constant :SOURCE_RUBY

    HEADER_SOURCE = 'X-Daytona-Source'
    private_constant :HEADER_SOURCE

    HEADER_SDK_VERSION = 'X-Daytona-SDK-Version'
    private_constant :HEADER_SDK_VERSION

    HEADER_ORGANIZATION_ID = 'X-Daytona-Organization-ID'
    private_constant :HEADER_ORGANIZATION_ID

    LABEL_CODE_TOOLBOX_LANGUAGE = 'code-toolbox-language'
    private_constant :LABEL_CODE_TOOLBOX_LANGUAGE
  end
end
