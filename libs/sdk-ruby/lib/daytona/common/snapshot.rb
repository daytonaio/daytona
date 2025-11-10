# frozen_string_literal: true

module Daytona
  class CreateSnapshotParams
    # @return [String] Name of the snapshot
    attr_reader :name

    # @return [String, Daytona::Image] Image of the snapshot. If a string is provided,
    #   it should be available on some registry. If an Image instance is provided,
    #   it will be used to create a new image in Daytona.
    attr_reader :image

    # @return [Daytona::Resources, nil] Resources of the snapshot
    attr_reader :resources

    # @return [Array<String>, nil] Entrypoint of the snapshot
    attr_reader :entrypoint

    # @param name [String] Name of the snapshot
    # @param image [String, Daytona::Image] Image of the snapshot
    # @param resources [Daytona::Resources, nil] Resources of the snapshot
    # @param entrypoint [Array<String>, nil] Entrypoint of the snapshot
    def initialize(name:, image:, resources: nil, entrypoint: nil)
      @name = name
      @image = image
      @resources = resources
      @entrypoint = entrypoint
    end
  end

  class Snapshot
    # @return [String] Unique identifier for the Snapshot
    attr_reader :id

    # @return [String, nil] Organization ID of the Snapshot
    attr_reader :organization_id

    # @return [Boolean, nil] Whether the Snapshot is general
    attr_reader :general

    # @return [String] Name of the Snapshot
    attr_reader :name

    # @return [String] Name of the Image of the Snapshot
    attr_reader :image_name

    # @return [String] State of the Snapshot
    attr_reader :state

    # @return [Float, nil] Size of the Snapshot
    attr_reader :size

    # @return [Array<String>, nil] Entrypoint of the Snapshot
    attr_reader :entrypoint

    # @return [Float] CPU of the Snapshot
    attr_reader :cpu

    # @return [Float] GPU of the Snapshot
    attr_reader :gpu

    # @return [Float] Memory of the Snapshot in GiB
    attr_reader :mem

    # @return [Float] Disk of the Snapshot in GiB
    attr_reader :disk

    # @return [String, nil] Error reason of the Snapshot
    attr_reader :error_reason

    # @return [String] Timestamp when the Snapshot was created
    attr_reader :created_at

    # @return [String] Timestamp when the Snapshot was last updated
    attr_reader :updated_at

    # @return [String, nil] Timestamp when the Snapshot was last used
    attr_reader :last_used_at

    # @return [DaytonaApiClient::BuildInfo, nil] Build information for the snapshot
    attr_reader :build_info

    # @param snapshot_dto [DaytonaApiClient::SnapshotDto] The snapshot DTO from the API
    def initialize(snapshot_dto) # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
      @id = snapshot_dto.id
      @organization_id = snapshot_dto.organization_id
      @general = snapshot_dto.general
      @name = snapshot_dto.name
      @image_name = snapshot_dto.image_name
      @state = snapshot_dto.state
      @size = snapshot_dto.size
      @entrypoint = snapshot_dto.entrypoint
      @cpu = snapshot_dto.cpu
      @gpu = snapshot_dto.gpu
      @mem = snapshot_dto.mem
      @disk = snapshot_dto.disk
      @error_reason = snapshot_dto.error_reason
      @created_at = snapshot_dto.created_at
      @updated_at = snapshot_dto.updated_at
      @last_used_at = snapshot_dto.last_used_at
      @build_info = snapshot_dto.build_info
    end

    # Creates a Snapshot instance from a SnapshotDto
    #
    # @param dto [DaytonaApiClient::SnapshotDto] The snapshot DTO from the API
    # @return [Daytona::Snapshot] The snapshot instance
    def self.from_dto(dto) = new(dto)
  end
end
