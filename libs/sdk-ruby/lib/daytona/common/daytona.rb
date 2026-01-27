# frozen_string_literal: true

module Daytona
  class CreateSandboxBaseParams
    # @return [Symbol, nil] Programming language for the Sandbox
    attr_accessor :language

    # @return [String, nil] OS user for the Sandbox
    attr_accessor :os_user

    # @return [Hash<String, String>, nil] Environment variables to set in the Sandbox
    attr_accessor :env_vars

    # @return [Hash<String, String>, nil] Custom labels for the Sandbox
    attr_accessor :labels

    # @return [Boolean, nil] Whether the Sandbox should be public
    attr_accessor :public

    # @return [Float, nil] Timeout in seconds for Sandbox to be created and started
    attr_accessor :timeout

    # @return [Integer, nil] Auto-stop interval in minutes
    attr_accessor :auto_stop_interval

    # @return [Integer, nil] Auto-archive interval in minutes
    attr_accessor :auto_archive_interval

    # @return [Integer, nil] Auto-delete interval in minutes
    attr_accessor :auto_delete_interval

    # @return [Array<DaytonaApiClient::SandboxVolume>, nil] List of volumes mounts to attach to the Sandbox
    attr_accessor :volumes

    # @return [Boolean, nil] Whether to block all network access for the Sandbox
    attr_accessor :network_block_all

    # @return [String, nil] Comma-separated list of allowed CIDR network addresses for the Sandbox
    attr_accessor :network_allow_list

    # @return [Boolean, nil] Whether the Sandbox should be ephemeral
    attr_accessor :ephemeral

    # Initialize CreateSandboxBaseParams
    #
    # @param language [Symbol, nil] Programming language for the Sandbox
    # @param os_user [String, nil] OS user for the Sandbox
    # @param env_vars [Hash<String, String>, nil] Environment variables to set in the Sandbox
    # @param labels [Hash<String, String>, nil] Custom labels for the Sandbox
    # @param public [Boolean, nil] Whether the Sandbox should be public
    # @param timeout [Float, nil] Timeout in seconds for Sandbox to be created and started
    # @param auto_stop_interval [Integer, nil] Auto-stop interval in minutes
    # @param auto_archive_interval [Integer, nil] Auto-archive interval in minutes
    # @param auto_delete_interval [Integer, nil] Auto-delete interval in minutes
    # @param volumes [Array<DaytonaApiClient::SandboxVolume>, nil] List of volumes mounts to attach to the Sandbox
    # @param network_block_all [Boolean, nil] Whether to block all network access for the Sandbox
    # @param network_allow_list [String, nil] Comma-separated list of allowed CIDR network addresses for the Sandbox
    # @param ephemeral [Boolean, nil] Whether the Sandbox should be ephemeral
    def initialize( # rubocop:disable Metrics/MethodLength, Metrics/ParameterLists
      language: nil,
      os_user: nil,
      env_vars: nil,
      labels: nil,
      public: nil,
      timeout: nil,
      auto_stop_interval: nil,
      auto_archive_interval: nil,
      auto_delete_interval: nil,
      volumes: nil,
      network_block_all: nil,
      network_allow_list: nil,
      ephemeral: nil
    )
      @language = language
      @os_user = os_user
      @env_vars = env_vars
      @labels = labels
      @public = public
      @timeout = timeout
      @auto_stop_interval = auto_stop_interval
      @auto_archive_interval = auto_archive_interval
      @auto_delete_interval = auto_delete_interval
      @volumes = volumes
      @network_block_all = network_block_all
      @network_allow_list = network_allow_list
      @ephemeral = ephemeral

      # Handle ephemeral and auto_delete_interval conflict
      handle_ephemeral_auto_delete_conflict
    end

    # Convert to hash representation
    #
    # @return [Hash<Symbol, Object>] Hash representation of the parameters
    def to_h # rubocop:disable Metrics/MethodLength
      {
        language:,
        os_user:,
        env_vars:,
        labels:,
        public:,
        timeout:,
        auto_stop_interval:,
        auto_archive_interval:,
        auto_delete_interval:,
        volumes:,
        network_block_all:,
        network_allow_list:,
        ephemeral:
      }.compact
    end

    private

    # Handle the conflict between ephemeral and auto_delete_interval
    #
    # @return [void]
    def handle_ephemeral_auto_delete_conflict
      return unless ephemeral && auto_delete_interval && !auto_delete_interval.zero?

      warn(
        "'ephemeral' and 'auto_delete_interval' cannot be used together. " \
        'If ephemeral is true, auto_delete_interval will be ignored and set to 0.'
      )
      @auto_delete_interval = 0
    end
  end

  class CreateSandboxFromImageParams < CreateSandboxBaseParams
    # @return [String, Image] Custom Docker image to use for the Sandbox. If an Image object is provided,
    #   the image will be dynamically built.
    attr_accessor :image

    # @return [Daytona::Resources, nil] Resource configuration for the Sandbox. If not provided, sandbox will
    #   have default resources.
    attr_accessor :resources

    # Initialize CreateSandboxFromImageParams
    #
    # @param image [String, Image] Custom Docker image to use for the Sandbox
    # @param resources [Daytona::Resources, nil] Resource configuration for the Sandbox
    # @param language [Symbol, nil] Programming language for the Sandbox
    # @param os_user [String, nil] OS user for the Sandbox
    # @param env_vars [Hash<String, String>, nil] Environment variables to set in the Sandbox
    # @param labels [Hash<String, String>, nil] Custom labels for the Sandbox
    # @param public [Boolean, nil] Whether the Sandbox should be public
    # @param timeout [Float, nil] Timeout in seconds for Sandbox to be created and started
    # @param auto_stop_interval [Integer, nil] Auto-stop interval in minutes
    # @param auto_archive_interval [Integer, nil] Auto-archive interval in minutes
    # @param auto_delete_interval [Integer, nil] Auto-delete interval in minutes
    # @param volumes [Array<DaytonaApiClient::SandboxVolume>, nil] List of volumes mounts to attach to the Sandbox
    # @param network_block_all [Boolean, nil] Whether to block all network access for the Sandbox
    # @param network_allow_list [String, nil] Comma-separated list of allowed CIDR network addresses for the Sandbox
    # @param ephemeral [Boolean, nil] Whether the Sandbox should be ephemeral
    def initialize(image:, resources: nil, **args)
      @image = image
      @resources = resources

      super(**args)
    end

    # Convert to hash representation
    #
    # @return [Hash<Symbol, Object>] Hash representation of the parameters
    def to_h
      super.merge(
        image:,
        resources: resources&.to_h
      ).compact
    end
  end

  class CreateSandboxFromSnapshotParams < CreateSandboxBaseParams
    # @return [String, nil] Name of the snapshot to use for the Sandbox
    attr_accessor :snapshot

    # Initialize CreateSandboxFromSnapshotParams
    #
    # @param snapshot [String, nil] Name of the snapshot to use for the Sandbox
    # @param language [Symbol, nil] Programming language for the Sandbox
    # @param os_user [String, nil] OS user for the Sandbox
    # @param env_vars [Hash<String, String>, nil] Environment variables to set in the Sandbox
    # @param labels [Hash<String, String>, nil] Custom labels for the Sandbox
    # @param public [Boolean, nil] Whether the Sandbox should be public
    # @param timeout [Float, nil] Timeout in seconds for Sandbox to be created and started
    # @param auto_stop_interval [Integer, nil] Auto-stop interval in minutes
    # @param auto_archive_interval [Integer, nil] Auto-archive interval in minutes
    # @param auto_delete_interval [Integer, nil] Auto-delete interval in minutes
    # @param volumes [Array<DaytonaApiClient::SandboxVolume>, nil] List of volumes mounts to attach to the Sandbox
    # @param network_block_all [Boolean, nil] Whether to block all network access for the Sandbox
    # @param network_allow_list [String, nil] Comma-separated list of allowed CIDR network addresses for the Sandbox
    # @param ephemeral [Boolean, nil] Whether the Sandbox should be ephemeral
    def initialize(snapshot: nil, **args)
      @snapshot = snapshot

      super(**args)
    end

    # Convert to hash representation
    #
    # @return [Hash<Symbol, Object>] Hash representation of the parameters
    def to_h
      super.merge(snapshot:).compact
    end
  end
end
