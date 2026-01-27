# frozen_string_literal: true

module Daytona
  class Volume
    # @return [String]
    attr_reader :id

    # @return [String]
    attr_reader :name

    # @return [String]
    attr_reader :organization_id

    # @return [String]
    attr_reader :state

    # @return [String]
    attr_reader :created_at

    # @return [String]
    attr_reader :updated_at

    # @return [String]
    attr_reader :last_used_at

    # @return [String, nil]
    attr_reader :error_reason

    # Initialize volume from DTO
    #
    # @param volume_dto [DaytonaApiClient::SandboxVolume]
    def initialize(volume_dto)
      @id = volume_dto.id
      @name = volume_dto.name
      @organization_id = volume_dto.organization_id
      @state = volume_dto.state
      @created_at = volume_dto.created_at
      @updated_at = volume_dto.updated_at
      @last_used_at = volume_dto.last_used_at
      @error_reason = volume_dto.error_reason
    end
  end
end
