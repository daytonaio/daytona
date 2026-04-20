# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::Daytona do
  let(:config) { build_config }
  let(:api_client) do
    double('ApiClient', default_headers: {}).tap do |client|
      allow(client).to receive(:user_agent=)
      allow(client).to receive(:update_params_for_auth!)
    end
  end
  let(:sandbox_api) { instance_double(DaytonaApiClient::SandboxApi) }
  let(:config_api) { instance_double(DaytonaApiClient::ConfigApi) }
  let(:volumes_api) { instance_double(DaytonaApiClient::VolumesApi) }
  let(:object_storage_api) { instance_double(DaytonaApiClient::ObjectStorageApi) }
  let(:snapshots_api) { instance_double(DaytonaApiClient::SnapshotsApi) }
  let(:sandbox_dto) { build_sandbox_dto }
  let(:sandbox) { instance_double(Daytona::Sandbox, id: 'sandbox-123', state: DaytonaApiClient::SandboxState::STARTED) }

  before do
    allow(DaytonaApiClient::ApiClient).to receive(:new).and_return(api_client)
    allow(DaytonaApiClient::SandboxApi).to receive(:new).and_return(sandbox_api)
    allow(DaytonaApiClient::ConfigApi).to receive(:new).and_return(config_api)
    allow(DaytonaApiClient::VolumesApi).to receive(:new).and_return(volumes_api)
    allow(DaytonaApiClient::ObjectStorageApi).to receive(:new).and_return(object_storage_api)
    allow(DaytonaApiClient::SnapshotsApi).to receive(:new).and_return(snapshots_api)
    allow(Daytona::Sandbox).to receive(:new).and_return(sandbox)
  end

  describe '#initialize' do
    it 'creates instance with valid api_key config' do
      daytona = described_class.new(config)

      expect(daytona.config).to eq(config)
      expect(daytona.volume).to be_a(Daytona::VolumeService)
      expect(daytona.snapshot).to be_a(Daytona::SnapshotService)
    end

    it 'configures API client headers and user agent' do
      described_class.new(config)

      expect(api_client.default_headers['X-Daytona-Source']).to eq('sdk-ruby')
      expect(api_client.default_headers['X-Daytona-SDK-Version']).to eq(Daytona::Sdk::VERSION)
      expect(api_client).to have_received(:user_agent=).with("sdk-ruby/#{Daytona::Sdk::VERSION}")
    end

    it 'adds organization header when using a JWT token' do
      jwt_config = Daytona::Config.new(jwt_token: 'jwt', organization_id: 'org-1', api_url: 'https://api.example.com')

      described_class.new(jwt_config)

      expect(api_client.default_headers['X-Daytona-Organization-ID']).to eq('org-1')
    end

    it 'initializes otel when the otel_enabled config option is set' do
      otel_state = double('OtelState')
      otel_config = build_config(otel_enabled: true)
      allow(Daytona).to receive(:init_otel).and_return(otel_state)

      described_class.new(otel_config)

      expect(Daytona).to have_received(:init_otel).with(Daytona::Sdk::VERSION)
    end

    it 'initializes otel when experimental config enables it' do
      otel_state = double('OtelState')
      experimental_config = build_config(_experimental: { 'otel_enabled' => true })
      allow(Daytona).to receive(:init_otel).and_return(otel_state)

      described_class.new(experimental_config)

      expect(Daytona).to have_received(:init_otel).with(Daytona::Sdk::VERSION)
    end

    it 'initializes otel when DAYTONA_OTEL_ENABLED is set' do
      env_config = build_config
      allow(env_config).to receive(:read_env).with('DAYTONA_OTEL_ENABLED').and_return('true')
      allow(env_config).to receive(:read_env).with('DAYTONA_EXPERIMENTAL_OTEL_ENABLED').and_return(nil)
      allow(Daytona).to receive(:init_otel).and_return(double('OtelState'))

      described_class.new(env_config)

      expect(Daytona).to have_received(:init_otel)
    end

    it 'initializes otel when DAYTONA_EXPERIMENTAL_OTEL_ENABLED is set' do
      env_config = build_config
      allow(env_config).to receive(:read_env).with('DAYTONA_OTEL_ENABLED').and_return(nil)
      allow(env_config).to receive(:read_env).with('DAYTONA_EXPERIMENTAL_OTEL_ENABLED').and_return('true')
      allow(Daytona).to receive(:init_otel).and_return(double('OtelState'))

      described_class.new(env_config)

      expect(Daytona).to have_received(:init_otel)
    end

    it 'raises error when no api_key or jwt_token provided' do
      bad_config = Daytona::Config.new(api_key: nil, jwt_token: nil, api_url: 'https://api.example.com')

      expect { described_class.new(bad_config) }
        .to raise_error(Daytona::Sdk::Error, /Authentication credentials not found. Set DAYTONA_API_KEY/)
    end

    it 'raises error when jwt_token without organization_id' do
      bad_config = Daytona::Config.new(jwt_token: 'jwt', organization_id: nil, api_url: 'https://api.example.com')

      expect { described_class.new(bad_config) }
        .to raise_error(Daytona::Sdk::Error, /DAYTONA_ORGANIZATION_ID is required/)
    end
  end

  describe '#create' do
    it 'creates a sandbox with default snapshot params and python language when params are nil' do
      allow(sandbox_api).to receive(:create_sandbox).and_return(sandbox_dto)

      daytona = described_class.new(config)
      result = daytona.create

      expect(result).to eq(sandbox)
      expect(sandbox_api).to have_received(:create_sandbox) do |request|
        expect(request.labels[Daytona::CODE_TOOLBOX_LANGUAGE_LABEL]).to eq('python')
      end
    end

    it 'fills in a default language when params.language is nil' do
      params = Daytona::CreateSandboxFromSnapshotParams.new(snapshot: 'snap-1')
      allow(sandbox_api).to receive(:create_sandbox).and_return(sandbox_dto)

      described_class.new(config).create(params)

      expect(params.language).to eq(:python)
    end

    it 'raises on invalid language values' do
      params = Daytona::CreateSandboxFromSnapshotParams.new(snapshot: 'snap-1', language: :ruby)

      expect { described_class.new(config).create(params) }
        .to raise_error(ArgumentError, /Invalid code-toolbox-language: ruby/)
    end

    it 'raises on negative timeout through the private create helper' do
      params = Daytona::CreateSandboxFromSnapshotParams.new(snapshot: 'snap-1', language: :python)

      expect { described_class.new(config).send(:_create, params, timeout: -1) }
        .to raise_error(Daytona::Sdk::Error, /Timeout must be a non-negative number/)
    end

    it 'raises on negative auto stop interval' do
      params = Daytona::CreateSandboxFromSnapshotParams.new(snapshot: 'snap-1', language: :python,
                                                            auto_stop_interval: -1)

      expect { described_class.new(config).create(params) }
        .to raise_error(Daytona::Sdk::Error, /auto_stop_interval must be a non-negative integer/)
    end

    it 'raises on negative auto archive interval' do
      params = Daytona::CreateSandboxFromSnapshotParams.new(snapshot: 'snap-1', language: :python,
                                                            auto_archive_interval: -1)

      expect { described_class.new(config).create(params) }
        .to raise_error(Daytona::Sdk::Error, /auto_archive_interval must be a non-negative integer/)
    end

    it 'creates a sandbox from a string image and merges labels' do
      params = Daytona::CreateSandboxFromImageParams.new(
        image: 'ruby:3.4',
        language: :typescript,
        labels: { 'env' => 'test' },
        env_vars: { 'A' => '1' }
      )
      allow(sandbox_api).to receive(:create_sandbox).and_return(sandbox_dto)

      described_class.new(config).create(params)

      expect(sandbox_api).to have_received(:create_sandbox) do |request|
        expect(request.env).to eq({ 'A' => '1' })
        expect(request.labels).to eq({ 'env' => 'test', Daytona::CODE_TOOLBOX_LANGUAGE_LABEL => 'typescript' })
        expect(request.build_info.dockerfile_content).to eq("FROM ruby:3.4\n")
      end
    end

    it 'creates a sandbox from an image object and passes resources and network settings' do
      image = Daytona::Image.base('python:3.12').workdir('/workspace')
      params = Daytona::CreateSandboxFromImageParams.new(
        image: image,
        language: :python,
        resources: Daytona::Resources.new(cpu: 2, memory: 4, disk: 8, gpu: 1),
        network_block_all: true,
        network_allow_list: '10.0.0.0/8'
      )
      allow(Daytona::SnapshotService).to receive(:process_image_context).and_return(['hash-1'])
      allow(sandbox_api).to receive(:create_sandbox).and_return(sandbox_dto)

      described_class.new(config).create(params)

      expect(Daytona::SnapshotService).to have_received(:process_image_context).with(object_storage_api, image)
      expect(sandbox_api).to have_received(:create_sandbox) do |request|
        expect(request.build_info.context_hashes).to eq(['hash-1'])
        expect(request.cpu).to eq(2)
        expect(request.memory).to eq(4)
        expect(request.disk).to eq(8)
        expect(request.gpu).to eq(1)
        expect(request.network_block_all).to be(true)
        expect(request.network_allow_list).to eq('10.0.0.0/8')
      end
    end

    it 'waits for the sandbox to start when the API returns a non-started state' do
      pending_sandbox = instance_double(Daytona::Sandbox, state: 'pending')
      allow(Daytona::Sandbox).to receive(:new).and_return(pending_sandbox)
      allow(pending_sandbox).to receive(:wait_for_sandbox_start)
      allow(sandbox_api).to receive(:create_sandbox).and_return(build_sandbox_dto(state: 'pending'))

      described_class.new(config).create(Daytona::CreateSandboxFromSnapshotParams.new(snapshot: 'snap-1',
                                                                                      language: :python))

      expect(pending_sandbox).to have_received(:wait_for_sandbox_start)
    end

    it 'streams build logs for pending builds when a callback is provided' do
      build_response = build_sandbox_dto(id: 'sb-1', state: DaytonaApiClient::SandboxState::PENDING_BUILD)
      started_response = build_sandbox_dto(id: 'sb-1', state: DaytonaApiClient::SandboxState::STARTED)
      build_logs = double('BuildLogsResponse', url: 'https://logs.example.com/build')
      callback = proc { |_chunk| }

      allow(sandbox_api).to receive(:create_sandbox).and_return(build_response)
      allow(sandbox_api).to receive(:get_sandbox).with('sb-1').and_return(started_response)
      allow(sandbox_api).to receive(:get_build_logs_url).with('sb-1').and_return(build_logs)
      allow(sandbox_api).to receive(:api_client).and_return(api_client)
      allow(api_client).to receive(:update_params_for_auth!)
      allow(Daytona::Util).to receive(:stream_async)
      allow_any_instance_of(described_class).to receive(:sleep)

      described_class.new(config).create(
        Daytona::CreateSandboxFromSnapshotParams.new(snapshot: 'snap-1', language: :python),
        on_snapshot_create_logs: callback
      )

      expect(api_client).to have_received(:update_params_for_auth!).with({}, nil, ['bearer'])
      expect(Daytona::Util).to have_received(:stream_async) do |uri:, headers:, on_chunk:|
        expect(uri.to_s).to eq('https://logs.example.com/build?follow=true')
        expect(headers).to eq({})
        expect(on_chunk).to eq(callback)
      end
    end
  end

  describe '#get' do
    it 'returns a Sandbox for the given id' do
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(sandbox_dto)

      result = described_class.new(config).get('sandbox-123')

      expect(result).to eq(sandbox)
      expect(Daytona::Sandbox).to have_received(:new).with(
        sandbox_dto: sandbox_dto,
        config: config,
        sandbox_api: sandbox_api,
        otel_state: nil
      )
    end
  end

  describe '#list' do
    let(:single_page_response) do
      instance_double(
        DaytonaApiClient::ListSandboxesResponse,
        items: [sandbox_dto],
        next_cursor: nil
      )
    end

    it 'returns a lazy Enumerator without performing the API call' do
      # No stub on sandbox_api — if list eagerly fetched, this would raise.
      result = described_class.new(config).list

      expect(result).to be_a(Enumerator)
    end

    it 'yields sandboxes from a single page' do
      allow(sandbox_api).to receive(:list_sandboxes).and_return(single_page_response)

      collected = described_class.new(config).list.to_a

      expect(collected).to eq([sandbox])
      expect(sandbox_api).to have_received(:list_sandboxes).once
    end

    it 'passes label filters and limit hint to the API client' do
      query = Daytona::ListSandboxesQuery.new(labels: { 'env' => 'test' }, limit: 10)
      allow(sandbox_api).to receive(:list_sandboxes)
        .and_return(single_page_response)

      described_class.new(config).list(query).to_a

      expect(sandbox_api).to have_received(:list_sandboxes) do |opts|
        expect(opts[:labels]).to eq('{"env":"test"}')
        expect(opts[:limit]).to eq(10)
        # cursor is internal: first page fetch must omit it (or pass nil).
        expect(opts[:cursor]).to be_nil
      end
    end

    it 'paginates across multiple pages using next_cursor' do
      page1 = instance_double(
        DaytonaApiClient::ListSandboxesResponse,
        items: [sandbox_dto, sandbox_dto],
        next_cursor: 'cursor-2'
      )
      page2 = instance_double(
        DaytonaApiClient::ListSandboxesResponse,
        items: [sandbox_dto],
        next_cursor: nil
      )
      allow(sandbox_api).to receive(:list_sandboxes).and_return(page1, page2)

      collected = described_class.new(config).list.to_a

      expect(collected.size).to eq(3)
      expect(sandbox_api).to have_received(:list_sandboxes).twice
    end

    it 'stops fetching when iteration terminates early' do
      page1 = instance_double(
        DaytonaApiClient::ListSandboxesResponse,
        items: [sandbox_dto, sandbox_dto],
        next_cursor: 'cursor-2'
      )
      allow(sandbox_api).to receive(:list_sandboxes).and_return(page1)

      first = described_class.new(config).list.first

      expect(first).to eq(sandbox)
      # Only page 1 was fetched; we never advanced past the first item.
      expect(sandbox_api).to have_received(:list_sandboxes).once
    end
  end

  describe '#start' do
    it 'delegates to sandbox.start' do
      allow(sandbox).to receive(:start)

      described_class.new(config).start(sandbox, 30)

      expect(sandbox).to have_received(:start).with(30)
    end
  end

  describe '#stop' do
    it 'delegates to sandbox.stop' do
      allow(sandbox).to receive(:stop)

      described_class.new(config).stop(sandbox, 30)

      expect(sandbox).to have_received(:stop).with(30)
    end
  end

  describe '#delete' do
    it 'delegates to sandbox.delete' do
      allow(sandbox).to receive(:delete)

      described_class.new(config).delete(sandbox)

      expect(sandbox).to have_received(:delete)
    end
  end

  describe '#close' do
    it 'shuts down otel and clears the state' do
      otel_state = double('OtelState')
      allow(Daytona).to receive(:init_otel).and_return(otel_state)
      allow(Daytona).to receive(:shutdown_otel)
      daytona = described_class.new(build_config(_experimental: { 'otel_enabled' => true }))

      daytona.close
      daytona.close

      expect(Daytona).to have_received(:shutdown_otel).with(otel_state).once
    end
  end
end
