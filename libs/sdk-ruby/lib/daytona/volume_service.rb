# frozen_string_literal: true

module Daytona
  class VolumeService
    # Service for managing Daytona Volumes. Can be used to list, get, create and delete Volumes.
    #
    # @param volumes_api [DaytonaApiClient::VolumesApi]
    def initialize(volumes_api)
      @volumes_api = volumes_api
    end

    # Create new Volume.
    #
    # @param name [String]
    # @return [Daytona::Volume]
    def create(name) = Volume.new(volumes_api.create_volume(DaytonaApiClient::CreateVolume.new(name:)))

    # Delete a Volume.
    #
    # @param volume [Daytona::Volume]
    # @return [void]
    def delete(volume) = volumes_api.delete_volume(volume.id)

    # Get a Volume by name.
    #
    # @param name [String]
    # @param create [Boolean]
    # @return [Daytona::Volume]
    def get(name, create: false)
      Volume.new(volumes_api.get_volume_by_name(name))
    rescue DaytonaApiClient::ApiError => e
      raise unless create && e.code == 404 && e.message.include?("Volume with name #{name} not found")

      create(name)
    end

    # List all Volumes.
    #
    # @return [Array<Daytona::Volume>]
    def list
      volumes_api.list_volumes.map { |volume| Volume.new(volume) }
    end

    private

    # @return [DaytonaApiClient::VolumesApi]
    attr_reader :volumes_api
  end
end
