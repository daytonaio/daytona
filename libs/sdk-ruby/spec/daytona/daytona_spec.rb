# frozen_string_literal: true

RSpec.describe Daytona::Daytona do
  let(:config) { build_config }
  let(:sandbox_api) { instance_double(DaytonaApiClient::SandboxApi) }
  let(:config_api) { instance_double(DaytonaApiClient::ConfigApi) }
  let(:volumes_api) { instance_double(DaytonaApiClient::VolumesApi) }
  let(:object_storage_api) { instance_double(DaytonaApiClient::ObjectStorageApi) }
  let(:snapshots_api) { instance_double(DaytonaApiClient::SnapshotsApi) }
  let(:sandbox_dto) { build_sandbox_dto }

  before do
    allow(DaytonaApiClient::SandboxApi).to receive(:new).and_return(sandbox_api)
    allow(DaytonaApiClient::ConfigApi).to receive(:new).and_return(config_api)
    allow(DaytonaApiClient::VolumesApi).to receive(:new).and_return(volumes_api)
    allow(DaytonaApiClient::ObjectStorageApi).to receive(:new).and_return(object_storage_api)
    allow(DaytonaApiClient::SnapshotsApi).to receive(:new).and_return(snapshots_api)
  end

  describe '#initialize' do
    it 'creates instance with valid api_key config' do
      daytona = described_class.new(config)
      expect(daytona.config).to eq(config)
    end

    it 'raises error when no api_key or jwt_token provided' do
      bad_config = Daytona::Config.new(api_key: nil, jwt_token: nil)
      expect { described_class.new(bad_config) }.to raise_error(Daytona::Sdk::Error, /API key or JWT token is required/)
    end

    it 'raises error when jwt_token without organization_id' do
      bad_config = Daytona::Config.new(jwt_token: 'jwt', organization_id: nil)
      expect { described_class.new(bad_config) }.to raise_error(Daytona::Sdk::Error, /Organization ID is required/)
    end

    it 'accepts jwt_token with organization_id' do
      jwt_config = Daytona::Config.new(jwt_token: 'jwt', organization_id: 'org-1')
      daytona = described_class.new(jwt_config)
      expect(daytona.config.jwt_token).to eq('jwt')
    end

    it 'exposes volume service' do
      daytona = described_class.new(config)
      expect(daytona.volume).to be_a(Daytona::VolumeService)
    end

    it 'exposes snapshot service' do
      daytona = described_class.new(config)
      expect(daytona.snapshot).to be_a(Daytona::SnapshotService)
    end
  end

  describe '#get' do
    it 'returns a Sandbox for the given id' do
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(sandbox_dto)

      daytona = described_class.new(config)
      sandbox = daytona.get('sandbox-123')

      expect(sandbox).to be_a(Daytona::Sandbox)
      expect(sandbox.id).to eq('sandbox-123')
    end
  end

  describe '#list' do
    let(:paginated_response) do
      instance_double(
        DaytonaApiClient::PaginatedSandboxes,
        total: 1, page: 1, total_pages: 1,
        items: [sandbox_dto]
      )
    end

    it 'returns a PaginatedResource' do
      allow(sandbox_api).to receive(:list_sandboxes_paginated).and_return(paginated_response)

      daytona = described_class.new(config)
      result = daytona.list

      expect(result).to be_a(Daytona::PaginatedResource)
      expect(result.total).to eq(1)
      expect(result.items.first).to be_a(Daytona::Sandbox)
    end

    it 'passes labels and pagination params' do
      allow(sandbox_api).to receive(:list_sandboxes_paginated)
        .with(labels: '{"env":"test"}', page: 2, limit: 10)
        .and_return(paginated_response)

      daytona = described_class.new(config)
      daytona.list({ 'env' => 'test' }, page: 2, limit: 10)
    end

    it 'raises error on invalid page' do
      daytona = described_class.new(config)
      expect { daytona.list({}, page: 0) }.to raise_error(Daytona::Sdk::Error, /page must be positive/)
    end

    it 'raises error on invalid limit' do
      daytona = described_class.new(config)
      expect { daytona.list({}, limit: -1) }.to raise_error(Daytona::Sdk::Error, /limit must be positive/)
    end
  end

  describe '#start' do
    it 'delegates to sandbox.start' do
      sandbox = instance_double(Daytona::Sandbox)
      allow(sandbox).to receive(:start)

      daytona = described_class.new(config)
      daytona.start(sandbox, 30)

      expect(sandbox).to have_received(:start).with(30)
    end
  end

  describe '#stop' do
    it 'delegates to sandbox.stop' do
      sandbox = instance_double(Daytona::Sandbox)
      allow(sandbox).to receive(:stop)

      daytona = described_class.new(config)
      daytona.stop(sandbox, 30)

      expect(sandbox).to have_received(:stop).with(30)
    end
  end

  describe '#delete' do
    it 'delegates to sandbox.delete' do
      sandbox = instance_double(Daytona::Sandbox)
      allow(sandbox).to receive(:delete)

      daytona = described_class.new(config)
      daytona.delete(sandbox)

      expect(sandbox).to have_received(:delete)
    end
  end

  describe '#close' do
    it 'does not raise when otel is disabled' do
      daytona = described_class.new(config)
      expect { daytona.close }.not_to raise_error
    end
  end
end
