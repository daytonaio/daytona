# frozen_string_literal: true

require 'timeout'

module Daytona
  class Sandbox # rubocop:disable Metrics/ClassLength
    DEFAULT_TIMEOUT = 60

    # @return [String] The ID of the sandbox
    attr_reader :id

    # @return [String] The organization ID of the sandbox
    attr_reader :organization_id

    # @return [String] The snapshot used for the sandbox
    attr_reader :snapshot

    # @return [String] The user associated with the project
    attr_reader :user

    # @return [Hash<String, String>] Environment variables for the sandbox
    attr_reader :env

    # @return [Hash<String, String>] Labels for the sandbox
    attr_reader :labels

    # @return [Boolean] Whether the sandbox http preview is public
    attr_reader :public

    # @return [Boolean] Whether to block all network access for the sandbox
    attr_reader :network_block_all

    # @return [String] Comma-separated list of allowed CIDR network addresses for the sandbox
    attr_reader :network_allow_list

    # @return [String] The target environment for the sandbox
    attr_reader :target

    # @return [Float] The CPU quota for the sandbox
    attr_reader :cpu

    # @return [Float] The GPU quota for the sandbox
    attr_reader :gpu

    # @return [Float] The memory quota for the sandbox
    attr_reader :memory

    # @return [Float] The disk quota for the sandbox
    attr_reader :disk

    # @return [DaytonaApiClient::SandboxState] The state of the sandbox
    attr_reader :state

    # @return [DaytonaApiClient::SandboxDesiredState] The desired state of the sandbox
    attr_reader :desired_state

    # @return [String] The error reason of the sandbox
    attr_reader :error_reason

    # @return [String] The state of the backup
    attr_reader :backup_state

    # @return [String] The creation timestamp of the last backup
    attr_reader :backup_created_at

    # @return [Float] Auto-stop interval in minutes (0 means disabled)
    attr_reader :auto_stop_interval

    # @return [Float] Auto-archive interval in minutes
    attr_reader :auto_archive_interval

    # @return [Float] Auto-delete interval in minutes
    # (negative value means disabled, 0 means delete immediately upon stopping)
    attr_reader :auto_delete_interval

    # @return [String] The domain name of the runner
    attr_reader :runner_domain

    # @return [Array<DaytonaApiClient::SandboxVolume>] Array of volumes attached to the sandbox
    attr_reader :volumes

    # @return [DaytonaApiClient::BuildInfo] Build information for the sandbox
    attr_reader :build_info

    # @return [String] The creation timestamp of the sandbox
    attr_reader :created_at

    # @return [String] The last update timestamp of the sandbox
    attr_reader :updated_at

    # @return [String] The version of the daemon running in the sandbox
    attr_reader :daemon_version

    # @return [Daytona::SandboxPythonCodeToolbox, Daytona::SandboxTsCodeToolbox]
    attr_reader :code_toolbox

    # @return [Daytona::Config]
    attr_reader :config

    # @return [DaytonaApiClient::SandboxApi]
    attr_reader :sandbox_api

    # @return [DaytonaApiClient::ToolboxApi]
    attr_reader :toolbox_api

    # @return [Daytona::Process]
    attr_reader :process

    # @return [Daytona::FileSystem]
    attr_reader :fs

    # @return [Daytona::Git]
    attr_reader :git

    # @return [Daytona::ComputerUse]
    attr_reader :computer_use

    # @params code_toolbox [Daytona::SandboxPythonCodeToolbox, Daytona::SandboxTsCodeToolbox]
    # @params config [Daytona::Config]
    # @params sandbox_api [DaytonaApiClient::SandboxApi]
    # @params sandbox_dto [DaytonaApiClient::Sandbox]
    # @params toolbox_api [DaytonaApiClient::ToolboxApi]
    def initialize(code_toolbox:, sandbox_dto:, config:, sandbox_api:, toolbox_api:) # rubocop:disable Metrics/MethodLength
      process_response(sandbox_dto)
      @code_toolbox = code_toolbox
      @config = config
      @sandbox_api = sandbox_api
      @toolbox_api = toolbox_api
      @process = Process.new(
        sandbox_id: id,
        code_toolbox:,
        toolbox_api:,
        get_preview_link: proc { |port| preview_url(port) }
      )
      @fs = FileSystem.new(sandbox_id: id, toolbox_api:)
      @git = Git.new(sandbox_id: id, toolbox_api:)
      @computer_use = ComputerUse.new(sandbox_id: id, toolbox_api:)
    end

    # Archives the sandbox, making it inactive and preserving its state. When sandboxes are
    # archived, the entire filesystem state is moved to cost-effective object storage, making it
    # possible to keep sandboxes available for an extended period. The tradeoff between archived
    # and stopped states is that starting an archived sandbox takes more time, depending on its size.
    # Sandbox must be stopped before archiving.
    #
    # @return [void]
    def archive
      sandbox_api.archive_sandbox(id)
      refresh
    end

    # Sets the auto-archive interval for the Sandbox.
    # The Sandbox will automatically archive after being continuously stopped for the specified interval.
    #
    # @param interval [Integer]
    # @return [Integer]
    # @raise [Daytona:Sdk::Error]
    def auto_archive_interval=(interval)
      raise Sdk::Error, 'Auto-archive interval must be a non-negative integer' if interval.negative?

      sandbox_api.set_auto_archive_interval(id, interval)
      @auto_archive_interval = interval
    end

    # Sets the auto-delete interval for the Sandbox.
    # The Sandbox will automatically delete after being continuously stopped for the specified interval.
    #
    # @param interval [Integer]
    # @return [Integer]
    # @raise [Daytona:Sdk::Error]
    def auto_delete_interval=(interval)
      sandbox_api.set_auto_delete_interval(id, interval)
      @auto_delete_interval = interval
    end

    # Sets the auto-stop interval for the Sandbox.
    # The Sandbox will automatically stop after being idle (no new events) for the specified interval.
    # Events include any state changes or interactions with the Sandbox through the SDK.
    # Interactions using Sandbox Previews are not included.
    #
    # @param interval [Integer]
    # @return [Integer]
    # @raise [Daytona:Sdk::Error]
    def auto_stop_interval=(interval)
      raise Sdk::Error, 'Auto-stop interval must be a non-negative integer' if interval.negative?

      sandbox_api.set_autostop_interval(id, interval)
      @auto_stop_interval = interval
    end

    # Creates an SSH access token for the sandbox.
    #
    # @param expires_in_minutes [Integer] TThe number of minutes the SSH access token will be valid for
    # @return [DaytonaApiClient::SshAccessDto]
    def create_ssh_access(expires_in_minutes) = sandbox_api.create_ssh_access(id, { expires_in_minutes: })

    # @return [void]
    def delete
      sandbox_api.delete_sandbox(id)
      refresh
    end

    # Sets labels for the Sandbox.
    #
    # @param labels [Hash<String, String>]
    # @return [Hash<String, String>]
    def labels=(labels)
      @labels = sandbox_api.replace_labels(id, DaytonaApiClient::SandboxLabels.build_from_hash(labels:)).labels
    end

    # Retrieves the preview link for the sandbox at the specified port. If the port is closed,
    # it will be opened automatically. For private sandboxes, a token is included to grant access
    # to the URL.
    #
    # @param port [Integer]
    # @return [DaytonaApiClient::PortPreviewUrl]
    def preview_url(port) = sandbox_api.get_port_preview_url(id, port)

    # Refresh the Sandbox data from the API.
    #
    # @return [void]
    def refresh = process_response(sandbox_api.get_sandbox(id))

    # Revokes an SSH access token for the sandbox.
    #
    # @param token [String]
    # @return [void]
    def revoke_ssh_access(token) = sandbox_api.revoke_ssh_access(id, token:)

    # Starts the Sandbox and waits for it to be ready.
    #
    # @param timeout [Numeric] Maximum wait time in seconds (defaults to 60 s).
    # @return [void]
    def start(timeout = DEFAULT_TIMEOUT)
      with_timeout(
        timeout:,
        message: "Sandbox #{id} failed to become ready within the #{timeout} seconds timeout period",
        setup: proc { process_response(sandbox_api.start_sandbox(id)) }
      ) { wait_for_states(operation: OPERATION_START, target_states: [DaytonaApiClient::SandboxState::STARTED]) }
    end

    # Stops the Sandbox and waits for it to be stopped.
    #
    # @param timeout [Numeric] Maximum wait time in seconds (defaults to 60 s).
    # @return [void]
    def stop(timeout = DEFAULT_TIMEOUT) # rubocop:disable Metrics/MethodLength
      with_timeout(
        timeout:,
        message: "Sandbox #{id} failed to become stopped within the #{timeout} seconds timeout period",
        setup: proc {
          sandbox_api.stop_sandbox(id)
          refresh
        }
      ) do
        wait_for_states(
          operation: OPERATION_STOP,
          target_states: [DaytonaApiClient::SandboxState::STOPPED, DaytonaApiClient::SandboxState::DESTROYED]
        )
      end
    end

    # Creates a new Language Server Protocol (LSP) server instance.
    # The LSP server provides language-specific features like code completion,
    # diagnostics, and more.
    #
    # @param language_id [Symbol] The language server type (e.g., Daytona::LspServer::Language::PYTHON)
    # @param path_to_project [String] Path to the project root directory. Relative paths are resolved
    #                      based on the sandbox working directory.
    # @return [Daytona::LspServer]
    def create_lsp_server(language_id:, path_to_project:)
      LspServer.new(language_id:, path_to_project:, toolbox_api:, sandbox_id: id)
    end

    #  Validates an SSH access token for the sandbox.
    #
    # @param token [String]
    # @return [DaytonaApiClient::SshAccessValidationDto]
    def validate_ssh_access(token) = sandbox_api.validate_ssh_access(token)

    # Waits for the Sandbox to reach the 'started' state. Polls the Sandbox status until it
    # reaches the 'started' state or encounters an error.
    #
    # @param timeout [Numeric] Maximum wait time in seconds (defaults to 60 s).
    # @return [void]
    def wait_for_sandbox_start(_timeout = DEFAULT_TIMEOUT)
      wait_for_states(operation: OPERATION_START, target_states: [DaytonaApiClient::SandboxState::STARTED])
    end

    private

    # @params sandbox_dto [DaytonaApiClient::Sandbox]
    # @return [void]
    def process_response(sandbox_dto) # rubocop:disable Metrics/MethodLength, Metrics/AbcSize
      @id = sandbox_dto.id
      @organization_id = sandbox_dto.organization_id
      @snapshot = sandbox_dto.snapshot
      @user = sandbox_dto.user
      @env = sandbox_dto.env
      @labels = sandbox_dto.labels
      @public = sandbox_dto.public
      @target = sandbox_dto.target
      @cpu = sandbox_dto.cpu
      @gpu = sandbox_dto.gpu
      @memory = sandbox_dto.memory
      @disk = sandbox_dto.disk
      @state = sandbox_dto.state
      @desired_state = sandbox_dto.desired_state
      @error_reason = sandbox_dto.error_reason
      @backup_state = sandbox_dto.backup_state
      @backup_created_at = sandbox_dto.backup_created_at
      @auto_stop_interval = sandbox_dto.auto_stop_interval
      @auto_archive_interval = sandbox_dto.auto_archive_interval
      @auto_delete_interval = sandbox_dto.auto_delete_interval
      @runner_domain = sandbox_dto.runner_domain
      @volumes = sandbox_dto.volumes
      @build_info = sandbox_dto.build_info
      @created_at = sandbox_dto.created_at
      @updated_at = sandbox_dto.updated_at
      @daemon_version = sandbox_dto.daemon_version
      @network_block_all = sandbox_dto.network_block_all
      @network_allow_list = sandbox_dto.network_allow_list
    end

    # Monitors block not to exceed max execution time.
    #
    # @param setup [#call, Nil] Optional setup block
    # @param timeout [Numeric] Maximum wait time in seconds (defaults to 60 s)
    # @param message [String] Error message
    # @return [void]
    # @raise [Daytona::Sdk::Error]
    def with_timeout(message:, setup:, timeout: DEFAULT_TIMEOUT, &)
      start_at = Time.now
      setup&.call

      Timeout.timeout(
        setup ? [NO_TIMEOUT, timeout - (Time.now - start_at)].max : timeout,
        Sdk::Error,
        message,
        &
      )
    end

    # Waits for the Sandbox to reach the one of the target states. Polls the Sandbox status until it
    # reaches the one of the target states or encounters an error. It will wait up to 60 seconds
    # for the Sandbox to reach one of the target states.
    #
    # @param operation [#to_s] Operation name for error message
    # @param target_states [Array<DaytonaApiClient::SandboxState>] List of the target states
    # @return [void]
    # @raise [Daytona::Sdk::Error]
    def wait_for_states(operation:, target_states:)
      loop do
        case state
        when *target_states then return
        when DaytonaApiClient::SandboxState::ERROR, DaytonaApiClient::SandboxState::BUILD_FAILED
          raise Sdk::Error, "Sandbox #{id} failed to #{operation} with state: #{state}, error reason: #{error_reason}"
        end

        sleep(IDLE_DURATION)
        refresh
      end
    end

    IDLE_DURATION = 0.1
    private_constant :IDLE_DURATION

    NO_TIMEOUT = 0
    private_constant :NO_TIMEOUT

    OPERATION_START = :start
    private_constant :OPERATION_START

    OPERATION_STOP = :stop
    private_constant :OPERATION_STOP
  end
end
