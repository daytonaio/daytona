# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::Sandbox do
  let(:config) { build_config }
  let(:sandbox_api) { instance_double(DaytonaApiClient::SandboxApi) }
  let(:sandbox_dto) { build_sandbox_dto }
  let(:toolbox_client) do
    double('ToolboxApiClient', default_headers: {}).tap do |client|
      allow(client).to receive(:user_agent=)
    end
  end
  let(:process_api) { instance_double(DaytonaToolboxApiClient::ProcessApi) }
  let(:fs_api) { instance_double(DaytonaToolboxApiClient::FileSystemApi) }
  let(:git_api) { instance_double(DaytonaToolboxApiClient::GitApi) }
  let(:lsp_api) { instance_double(DaytonaToolboxApiClient::LspApi) }
  let(:computer_use_api) { instance_double(DaytonaToolboxApiClient::ComputerUseApi) }
  let(:interpreter_api) { instance_double(DaytonaToolboxApiClient::InterpreterApi) }
  let(:info_api) { instance_double(DaytonaToolboxApiClient::InfoApi) }

  before do
    allow(DaytonaToolboxApiClient::ApiClient).to receive(:new).and_return(toolbox_client)
    allow(DaytonaToolboxApiClient::ProcessApi).to receive(:new).and_return(process_api)
    allow(DaytonaToolboxApiClient::FileSystemApi).to receive(:new).and_return(fs_api)
    allow(DaytonaToolboxApiClient::GitApi).to receive(:new).and_return(git_api)
    allow(DaytonaToolboxApiClient::LspApi).to receive(:new).and_return(lsp_api)
    allow(DaytonaToolboxApiClient::ComputerUseApi).to receive(:new).and_return(computer_use_api)
    allow(DaytonaToolboxApiClient::InterpreterApi).to receive(:new).and_return(interpreter_api)
    allow(DaytonaToolboxApiClient::InfoApi).to receive(:new).and_return(info_api)
  end

  let(:sandbox) do
    described_class.new(
      sandbox_dto: sandbox_dto,
      config: config,
      sandbox_api: sandbox_api
    )
  end

  describe '#initialize' do
    it 'populates attributes from DTO' do
      expect(sandbox.id).to eq('sandbox-123')
      expect(sandbox.organization_id).to eq('org-1')
      expect(sandbox.user).to eq('daytona')
      expect(sandbox.state).to eq('started')
      expect(sandbox.cpu).to eq(4)
      expect(sandbox.memory).to eq(8)
      expect(sandbox.disk).to eq(30)
      expect(sandbox.last_activity_at).to eq('2025-01-01T00:00:00Z')
      expect(sandbox.network_block_all).to be(false)
      expect(sandbox.network_allow_list).to be_nil
    end

    it 'creates process, fs, git, computer_use, and code_interpreter helpers' do
      expect(sandbox.process).to be_a(Daytona::Process)
      expect(sandbox.fs).to be_a(Daytona::FileSystem)
      expect(sandbox.git).to be_a(Daytona::Git)
      expect(sandbox.computer_use).to be_a(Daytona::ComputerUse)
      expect(sandbox.code_interpreter).to be_a(Daytona::CodeInterpreter)
    end

    it 'configures toolbox client authorization and sdk headers' do
      described_class.new(sandbox_dto: sandbox_dto, config: config, sandbox_api: sandbox_api)

      expect(toolbox_client.default_headers['Authorization']).to eq('Bearer test-api-key')
      expect(toolbox_client.default_headers['X-Daytona-Source']).to eq('sdk-ruby')
      expect(toolbox_client.default_headers['X-Daytona-SDK-Version']).to eq(Daytona::Sdk::VERSION)
      expect(toolbox_client).to have_received(:user_agent=).with("sdk-ruby/#{Daytona::Sdk::VERSION}").at_least(:once)
    end

    it 'adds organization header when using JWT auth' do
      jwt_config = Daytona::Config.new(jwt_token: 'jwt', organization_id: 'org-9', api_url: 'https://api.example.com')

      described_class.new(sandbox_dto: sandbox_dto, config: jwt_config, sandbox_api: sandbox_api)

      expect(toolbox_client.default_headers['X-Daytona-Organization-ID']).to eq('org-9')
    end
  end

  describe '#archive' do
    it 'calls archive_sandbox and refreshes' do
      allow(sandbox_api).to receive(:archive_sandbox).with('sandbox-123')
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(build_sandbox_dto(state: 'archived'))

      sandbox.archive

      expect(sandbox_api).to have_received(:archive_sandbox).with('sandbox-123')
      expect(sandbox.state).to eq('archived')
    end
  end

  describe '#auto_stop_interval=' do
    it 'sets interval via API' do
      allow(sandbox_api).to receive(:set_autostop_interval).with('sandbox-123', 30)

      sandbox.auto_stop_interval = 30

      expect(sandbox.auto_stop_interval).to eq(30)
    end

    it 'raises on negative interval' do
      expect { sandbox.auto_stop_interval = -1 }
        .to raise_error(Daytona::Sdk::Error, /Auto-stop interval must be a non-negative integer/)
    end
  end

  describe '#auto_archive_interval=' do
    it 'sets interval via API' do
      allow(sandbox_api).to receive(:set_auto_archive_interval).with('sandbox-123', 60)

      sandbox.auto_archive_interval = 60

      expect(sandbox.auto_archive_interval).to eq(60)
    end

    it 'raises on negative interval' do
      expect { sandbox.auto_archive_interval = -1 }
        .to raise_error(Daytona::Sdk::Error, /Auto-archive interval must be a non-negative integer/)
    end
  end

  describe '#auto_delete_interval=' do
    it 'sets interval via API' do
      allow(sandbox_api).to receive(:set_auto_delete_interval).with('sandbox-123', 120)

      sandbox.auto_delete_interval = 120

      expect(sandbox.auto_delete_interval).to eq(120)
    end
  end

  describe '#create_ssh_access' do
    it 'delegates to sandbox_api' do
      ssh_dto = instance_double(DaytonaApiClient::SshAccessDto)
      allow(sandbox_api).to receive(:create_ssh_access)
        .with('sandbox-123', { expires_in_minutes: 60 })
        .and_return(ssh_dto)

      expect(sandbox.create_ssh_access(60)).to eq(ssh_dto)
    end
  end

  describe '#delete' do
    it 'calls sandbox_api.delete_sandbox and refreshes' do
      allow(sandbox_api).to receive(:delete_sandbox).with('sandbox-123')
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(build_sandbox_dto(state: 'destroyed'))

      sandbox.delete

      expect(sandbox_api).to have_received(:delete_sandbox).with('sandbox-123')
      expect(sandbox.state).to eq('destroyed')
    end

    it 'sets state to destroyed on 404' do
      allow(sandbox_api).to receive(:delete_sandbox).and_raise(DaytonaApiClient::ApiError.new(code: 404,
                                                                                              message: 'Not found'))

      sandbox.delete

      expect(sandbox.state).to eq('destroyed')
    end

    it 're-raises non-404 API errors' do
      allow(sandbox_api).to receive(:delete_sandbox).and_raise(DaytonaApiClient::ApiError.new(code: 500,
                                                                                              message: 'boom'))

      expect { sandbox.delete }.to raise_error(DaytonaApiClient::ApiError)
    end
  end

  describe '#get_user_home_dir' do
    it 'returns the user home directory' do
      allow(info_api).to receive(:get_user_home_dir).and_return(double('DirResponse', dir: '/home/daytona'))

      expect(sandbox.get_user_home_dir).to eq('/home/daytona')
    end

    it 'wraps errors in Sdk::Error' do
      allow(info_api).to receive(:get_user_home_dir).and_raise(StandardError, 'connection refused')

      expect { sandbox.get_user_home_dir }.to raise_error(Daytona::Sdk::Error, /Failed to get user home directory/)
    end
  end

  describe '#get_work_dir' do
    it 'returns the working directory' do
      allow(info_api).to receive(:get_work_dir).and_return(double('DirResponse', dir: '/workspace'))

      expect(sandbox.get_work_dir).to eq('/workspace')
    end

    it 'wraps errors in Sdk::Error' do
      allow(info_api).to receive(:get_work_dir).and_raise(StandardError, 'timeout')

      expect { sandbox.get_work_dir }.to raise_error(Daytona::Sdk::Error, /Failed to get working directory path/)
    end
  end

  describe '#labels=' do
    it 'replaces labels via API' do
      label_request = instance_double(DaytonaApiClient::SandboxLabels)
      label_response = instance_double(DaytonaApiClient::SandboxLabels, labels: { 'env' => 'prod' })
      allow(DaytonaApiClient::SandboxLabels).to receive(:build_from_hash).with(labels: { 'env' => 'prod' }).and_return(label_request)
      allow(sandbox_api).to receive(:replace_labels).with('sandbox-123', label_request).and_return(label_response)

      sandbox.labels = { 'env' => 'prod' }

      expect(sandbox.labels).to eq({ 'env' => 'prod' })
    end
  end

  describe '#preview_url' do
    it 'delegates to sandbox_api.get_port_preview_url' do
      preview = instance_double(DaytonaApiClient::PortPreviewUrl)
      allow(sandbox_api).to receive(:get_port_preview_url).with('sandbox-123', 3000).and_return(preview)

      expect(sandbox.preview_url(3000)).to eq(preview)
    end
  end

  describe '#create_signed_preview_url' do
    it 'delegates to sandbox_api with optional expiration' do
      signed_url = instance_double(DaytonaApiClient::SignedPortPreviewUrl)
      allow(sandbox_api).to receive(:get_signed_port_preview_url)
        .with('sandbox-123', 3000, { expires_in_seconds: 120 })
        .and_return(signed_url)

      expect(sandbox.create_signed_preview_url(3000, 120)).to eq(signed_url)
    end

    it 'passes nil expiration when not provided' do
      allow(sandbox_api).to receive(:get_signed_port_preview_url)
        .with('sandbox-123', 3000, { expires_in_seconds: nil })

      sandbox.create_signed_preview_url(3000)

      expect(sandbox_api).to have_received(:get_signed_port_preview_url)
    end
  end

  describe '#expire_signed_preview_url' do
    it 'delegates to sandbox_api' do
      allow(sandbox_api).to receive(:expire_signed_port_preview_url).with('sandbox-123', 3000, 'token-val')

      sandbox.expire_signed_preview_url(3000, 'token-val')

      expect(sandbox_api).to have_received(:expire_signed_port_preview_url).with('sandbox-123', 3000, 'token-val')
    end
  end

  describe '#refresh' do
    it 'updates state from API' do
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(build_sandbox_dto(state: 'stopped'))

      sandbox.refresh

      expect(sandbox.state).to eq('stopped')
    end
  end

  describe '#refresh_activity' do
    it 'calls update_last_activity' do
      allow(sandbox_api).to receive(:update_last_activity).with('sandbox-123')

      expect(sandbox.refresh_activity).to be_nil
    end

    it 'wraps errors in Sdk::Error' do
      allow(sandbox_api).to receive(:update_last_activity).and_raise(StandardError, 'offline')

      expect { sandbox.refresh_activity }.to raise_error(Daytona::Sdk::Error, /Failed to refresh sandbox activity/)
    end
  end

  describe '#revoke_ssh_access' do
    it 'delegates to sandbox_api' do
      allow(sandbox_api).to receive(:revoke_ssh_access).with('sandbox-123', token: 'token-1')

      sandbox.revoke_ssh_access('token-1')

      expect(sandbox_api).to have_received(:revoke_ssh_access).with('sandbox-123', token: 'token-1')
    end
  end

  describe '#start' do
    it 'calls start_sandbox and waits for started state' do
      allow(sandbox_api).to receive(:start_sandbox).with('sandbox-123').and_return(build_sandbox_dto(state: DaytonaApiClient::SandboxState::STARTED))

      sandbox.start(5)

      expect(sandbox_api).to have_received(:start_sandbox).with('sandbox-123')
    end
  end

  describe '#recover' do
    it 'calls recover_sandbox and waits for started state' do
      allow(sandbox_api).to receive(:recover_sandbox).with('sandbox-123').and_return(build_sandbox_dto(state: DaytonaApiClient::SandboxState::STARTED))

      sandbox.recover(5)

      expect(sandbox_api).to have_received(:recover_sandbox).with('sandbox-123')
    end

    it 'wraps errors in Sdk::Error' do
      allow(sandbox_api).to receive(:recover_sandbox).and_raise(StandardError, 'bad state')

      expect { sandbox.recover(5) }.to raise_error(Daytona::Sdk::Error, /Failed to recover sandbox/)
    end
  end

  describe '#stop' do
    it 'calls stop_sandbox and waits for stopped state' do
      allow(sandbox_api).to receive(:stop_sandbox).with('sandbox-123', { force: false })
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(build_sandbox_dto(state: DaytonaApiClient::SandboxState::STOPPED))

      sandbox.stop(5)

      expect(sandbox_api).to have_received(:stop_sandbox).with('sandbox-123', { force: false })
    end

    it 'passes force stop options through to the API' do
      allow(sandbox_api).to receive(:stop_sandbox).with('sandbox-123', { force: true })
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(build_sandbox_dto(state: DaytonaApiClient::SandboxState::DESTROYED))

      sandbox.stop(5, force: true)

      expect(sandbox_api).to have_received(:stop_sandbox).with('sandbox-123', { force: true })
    end
  end

  describe '#resize' do
    it 'raises error when resources is nil' do
      expect { sandbox.resize(nil) }.to raise_error(Daytona::Sdk::Error, /Resources must not be nil/)
    end

    it 'calls resize_sandbox with resource params' do
      allow(sandbox_api).to receive(:resize_sandbox).and_return(build_sandbox_dto(
                                                                  state: DaytonaApiClient::SandboxState::STARTED, cpu: 8
                                                                ))

      sandbox.resize(Daytona::Resources.new(cpu: 8, memory: 16, disk: 32), 5)

      expect(sandbox_api).to have_received(:resize_sandbox) do |_id, request|
        expect(request.cpu).to eq(8)
        expect(request.memory).to eq(16)
        expect(request.disk).to eq(32)
      end
    end
  end

  describe '#create_lsp_server' do
    it 'returns an LspServer instance' do
      lsp = sandbox.create_lsp_server(language_id: :python, path_to_project: '/workspace')

      expect(lsp).to be_a(Daytona::LspServer)
      expect(lsp.language_id).to eq(:python)
      expect(lsp.path_to_project).to eq('/workspace')
    end
  end

  describe '#validate_ssh_access' do
    it 'delegates to the sandbox api' do
      validation = double('Validation')
      allow(sandbox_api).to receive(:validate_ssh_access).with('token-1').and_return(validation)

      expect(sandbox.validate_ssh_access('token-1')).to eq(validation)
    end
  end

  describe '#wait_for_sandbox_start' do
    it 'raises when the sandbox enters an error state' do
      errored = described_class.new(
        sandbox_dto: build_sandbox_dto(state: DaytonaApiClient::SandboxState::ERROR, error_reason: 'boom'),
        config: config,
        sandbox_api: sandbox_api
      )

      expect do
        errored.wait_for_sandbox_start
      end.to raise_error(Daytona::Sdk::Error,
                         /failed to start with state: error, error reason: boom/i)
    end
  end

  describe '#wait_for_sandbox_stop' do
    it 'returns immediately when already destroyed' do
      destroyed = described_class.new(
        sandbox_dto: build_sandbox_dto(state: DaytonaApiClient::SandboxState::DESTROYED),
        config: config,
        sandbox_api: sandbox_api
      )

      expect { destroyed.wait_for_sandbox_stop }.not_to raise_error
    end
  end

  describe '#experimental_create_snapshot' do
    it 'creates a snapshot and waits for completion' do
      allow(sandbox_api).to receive(:create_sandbox_snapshot).with('sandbox-123', anything)
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(build_sandbox_dto(state: DaytonaApiClient::SandboxState::STARTED))

      sandbox.experimental_create_snapshot(name: 'snap-1', timeout: 5)

      expect(sandbox_api).to have_received(:create_sandbox_snapshot) do |_id, request|
        expect(request.name).to eq('snap-1')
      end
    end
  end
end
