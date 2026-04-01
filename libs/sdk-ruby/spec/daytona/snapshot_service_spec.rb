# frozen_string_literal: true

RSpec.describe Daytona::SnapshotService do
  let(:snapshots_api) { instance_double(DaytonaApiClient::SnapshotsApi) }
  let(:object_storage_api) { instance_double(DaytonaApiClient::ObjectStorageApi) }
  let(:service) do
    described_class.new(
      snapshots_api: snapshots_api,
      object_storage_api: object_storage_api,
      default_region_id: 'us'
    )
  end

  describe '#list' do
    it 'returns PaginatedResource of Snapshots' do
      dto = build_snapshot_dto
      paginated = instance_double(
        DaytonaApiClient::PaginatedSnapshots,
        total: 1, page: 1, total_pages: 1, items: [dto]
      )
      allow(snapshots_api).to receive(:get_all_snapshots).with(page: nil, limit: nil).and_return(paginated)

      result = service.list
      expect(result).to be_a(Daytona::PaginatedResource)
      expect(result.items.first).to be_a(Daytona::Snapshot)
      expect(result.total).to eq(1)
    end

    it 'passes pagination params' do
      paginated = instance_double(
        DaytonaApiClient::PaginatedSnapshots,
        total: 5, page: 2, total_pages: 3, items: []
      )
      allow(snapshots_api).to receive(:get_all_snapshots).with(page: 2, limit: 10).and_return(paginated)

      result = service.list(page: 2, limit: 10)
      expect(result.page).to eq(2)
    end

    it 'raises on invalid page' do
      expect { service.list(page: 0) }.to raise_error(Daytona::Sdk::Error, /page must be positive/)
    end

    it 'raises on invalid limit' do
      expect { service.list(limit: -1) }.to raise_error(Daytona::Sdk::Error, /limit must be positive/)
    end
  end

  describe '#get' do
    it 'returns Snapshot by name' do
      dto = build_snapshot_dto(name: 'my-snap')
      allow(snapshots_api).to receive(:get_snapshot).with('my-snap').and_return(dto)

      snapshot = service.get('my-snap')
      expect(snapshot).to be_a(Daytona::Snapshot)
      expect(snapshot.name).to eq('my-snap')
    end
  end

  describe '#delete' do
    it 'removes snapshot by id' do
      snapshot = Daytona::Snapshot.from_dto(build_snapshot_dto)
      allow(snapshots_api).to receive(:remove_snapshot).with('snap-123')

      service.delete(snapshot)
      expect(snapshots_api).to have_received(:remove_snapshot).with('snap-123')
    end
  end

  describe '#activate' do
    it 'activates a snapshot' do
      snapshot = Daytona::Snapshot.from_dto(build_snapshot_dto)
      activated_dto = build_snapshot_dto(state: 'active')
      allow(snapshots_api).to receive(:activate_snapshot).with('snap-123').and_return(activated_dto)

      result = service.activate(snapshot)
      expect(result).to be_a(Daytona::Snapshot)
      expect(result.state).to eq('active')
    end
  end

  describe '#create' do
    it 'creates a snapshot from a string image and waits' do
      created_dto = build_snapshot_dto(state: DaytonaApiClient::SnapshotState::ACTIVE)
      allow(snapshots_api).to receive(:create_snapshot).and_return(created_dto)
      allow(snapshots_api).to receive(:get_snapshot).and_return(created_dto)

      params = Daytona::CreateSnapshotParams.new(name: 'my-snap', image: 'ubuntu:22.04')
      result = service.create(params)
      expect(result).to be_a(Daytona::Snapshot)
      expect(result.name).to eq('test-snapshot')
    end

    it 'raises on failed snapshot creation' do
      failed_dto = build_snapshot_dto(
        state: DaytonaApiClient::SnapshotState::BUILD_FAILED,
        error_reason: 'dockerfile error'
      )
      allow(snapshots_api).to receive(:create_snapshot).and_return(failed_dto)
      allow(snapshots_api).to receive(:get_snapshot).and_return(failed_dto)

      params = Daytona::CreateSnapshotParams.new(name: 'fail-snap', image: 'bad:image')
      expect { service.create(params) }.to raise_error(Daytona::Sdk::Error, /Failed to create snapshot/)
    end
  end
end
