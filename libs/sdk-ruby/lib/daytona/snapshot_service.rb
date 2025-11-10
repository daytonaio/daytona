# frozen_string_literal: true

require 'uri'

module Daytona
  class SnapshotService
    SNAPSHOTS_FETCH_LIMIT = 200

    # @param snapshots_api [DaytonaApiClient::SnapshotsApi] The snapshots API client
    # @param object_storage_api [DaytonaApiClient::ObjectStorageApi] The object storage API client
    def initialize(snapshots_api:, object_storage_api:)
      @snapshots_api = snapshots_api
      @object_storage_api = object_storage_api
    end

    # List all Snapshots.
    #
    # @param page [Integer, Nil]
    # @param limit [Integer, Nil]
    # @return [Daytona::PaginatedResource] Paginated list of all Snapshots
    # @raise [Daytona::Sdk::Error]
    #
    # @example
    #   daytona = Daytona::Daytona.new
    #   response = daytona.snapshot.list(page: 1, limit: 10)
    #   snapshots.items.each { |snapshot| puts "#{snapshot.name} (#{snapshot.image_name})" }
    def list(page: nil, limit: nil)
      raise Sdk::Error, 'page must be positive integer' if page && page < 1

      raise Sdk::Error, 'limit must be positive integer' if limit && limit < 1

      response = snapshots_api.get_all_snapshots(page:, limit:)
      PaginatedResource.new(
        total: response.total,
        page: response.page,
        total_pages: response.total_pages,
        items: response.items.map { |snapshot_dto| Snapshot.from_dto(snapshot_dto) }
      )
    end

    # Delete a Snapshot.
    #
    # @param snapshot [Daytona::Snapshot] Snapshot to delete
    # @return [void]
    #
    # @example
    #   daytona = Daytona::Daytona.new
    #   snapshot = daytona.snapshot.get("demo")
    #   daytona.snapshot.delete(snapshot)
    #   puts "Snapshot deleted"
    def delete(snapshot) = snapshots_api.remove_snapshot(snapshot.id)

    # Get a Snapshot by name.
    #
    # @param name [String] Name of the Snapshot to get
    # @return [Daytona::Snapshot] The Snapshot object
    #
    # @example
    #   daytona = Daytona::Daytona.new
    #   snapshot = daytona.snapshot.get("demo")
    #   puts "#{snapshot.name} (#{snapshot.image_name})"
    def get(name) = Snapshot.from_dto(snapshots_api.get_snapshot(name))

    # Creates and registers a new snapshot from the given Image definition.
    #
    # @param params [Daytona::CreateSnapshotParams] Parameters for snapshot creation
    # @param on_logs [Proc, Nil] Callback proc handling snapshot creation logs
    # @return [Daytona::Snapshot] The created snapshot
    #
    # @example
    #   image = Image.debianSlim('3.12').pipInstall('numpy')
    #   params = CreateSnapshotParams.new(name: 'my-snapshot', image: image)
    #   snapshot = daytona.snapshot.create(params) do |chunk|
    #     print chunk
    #   end
    def create(params, on_logs: nil) # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
      create_snapshot_req = DaytonaApiClient::CreateSnapshot.new(name: params.name)

      if params.image.is_a?(String)
        create_snapshot_req.image_name = params.image
        create_snapshot_req.entrypoint = params.entrypoint
      else
        create_snapshot_req.build_info = DaytonaApiClient::CreateBuildInfo.new(
          context_hashes: self.class.process_image_context(object_storage_api, params.image),
          dockerfile_content: if params.entrypoint
                                params.image.entrypoint(params.entrypoint).dockerfile
                              else
                                params.image.dockerfile
                              end
        )
      end

      if params.resources
        create_snapshot_req.cpu = params.resources.cpu
        create_snapshot_req.gpu = params.resources.gpu
        create_snapshot_req.memory = params.resources.memory
        create_snapshot_req.disk = params.resources.disk
      end

      snapshot = snapshots_api.create_snapshot(create_snapshot_req)

      snapshot = stream_logs(snapshot, on_logs:) if on_logs

      if [DaytonaApiClient::SnapshotState::ERROR, DaytonaApiClient::SnapshotState::BUILD_FAILED].include?(snapshot.state)
        raise Sdk::Error, "Failed to create snapshot #{snapshot.name}, reason: #{snapshot.error_reason}"
      end

      Snapshot.from_dto(snapshot)
    end

    # Activate a snapshot
    #
    # @param snapshot [Daytona::Snapshot] The snapshot instance
    # @return [Daytona::Snapshot]
    def activate(snapshot) = Snapshot.from_dto(snapshots_api.activate_snapshot(snapshot.id))

    # Processes the image context by uploading it to object storage
    #
    # @param image [Daytona::Image] The Image instance
    # @return [Array<String>] List of context hashes stored in object storage
    def self.process_image_context(object_storage_api, image) # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
      return [] unless image.context_list && !image.context_list.empty?

      push_access_creds = object_storage_api.get_push_access

      object_storage = ObjectStorage.new(
        endpoint_url: push_access_creds.storage_url,
        aws_access_key_id: push_access_creds.access_key,
        aws_secret_access_key: push_access_creds.secret,
        aws_session_token: push_access_creds.session_token,
        bucket_name: push_access_creds.bucket
      )

      image.context_list.map do |context|
        object_storage.upload(
          context.source_path,
          push_access_creds.organization_id,
          context.archive_path
        )
      end
    end

    private

    # @return [DaytonaApiClient::SnapshotsApi] The snapshots API client
    attr_reader :snapshots_api

    # @return [DaytonaApiClient::ObjectStorageApi, nil] The object storage API client
    attr_reader :object_storage_api

    # @param snapshot [DaytonaApiClient::SnapshotDto]
    # @param on_logs [Proc]
    # @return [DaytonaApiClient::SnapshotDto]
    def stream_logs(snapshot, on_logs:) # rubocop:disable Metrics/AbcSize, Metrics/MethodLength
      terminal_states = [
        DaytonaApiClient::SnapshotState::ACTIVE,
        DaytonaApiClient::SnapshotState::ERROR,
        DaytonaApiClient::SnapshotState::BUILD_FAILED
      ]

      thread = nil
      previous_state = snapshot.state
      until terminal_states.include?(snapshot.state)
        Sdk.logger.debug("Waiting for snapshot to be created: #{snapshot.state}")
        if thread.nil? && snapshot.state != DaytonaApiClient::SnapshotState::BUILD_PENDING
          thread = start_log_streaming(snapshot, on_logs:)
        end

        on_logs.call("Creating snapshot #{snapshot.name} (#{snapshot.state})") if previous_state != snapshot.state

        sleep(1)
        previous_state = snapshot.state
        snapshot = snapshots_api.get_snapshot(snapshot.id)
      end

      thread&.join

      if snapshot.state == DaytonaApiClient::SnapshotState::ACTIVE
        on_logs.call("Created snapshot #{snapshot.name} (#{snapshot.state})")
      end

      snapshot
    end

    # @param snapshot [DaytonaApiClient::SnapshotDto]
    # @param on_logs [Proc]
    # @return [Thread]
    def start_log_streaming(snapshot, on_logs:)
      uri = URI.parse(snapshots_api.api_client.config.base_url)
      uri.path = "/api/snapshots/#{snapshot.id}/build-logs"
      uri.query = 'follow=true'

      headers = {}
      snapshots_api.api_client.update_params_for_auth!(headers, nil, ['bearer'])
      Util.stream_async(uri:, headers:, on_chunk: on_logs)
    end
  end
end
