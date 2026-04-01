# frozen_string_literal: true

RSpec.describe Daytona::Sandbox do
  let(:config) { build_config }
  let(:sandbox_api) { instance_double(DaytonaApiClient::SandboxApi) }
  let(:sandbox_dto) { build_sandbox_dto }
  let(:code_toolbox) { Daytona::SandboxPythonCodeToolbox.new }

  let(:sandbox) do
    described_class.new(
      sandbox_dto: sandbox_dto,
      config: config,
      sandbox_api: sandbox_api,
      code_toolbox: code_toolbox
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
    end

    it 'creates process, fs, git, computer_use, code_interpreter sub-objects' do
      expect(sandbox.process).to be_a(Daytona::Process)
      expect(sandbox.fs).to be_a(Daytona::FileSystem)
      expect(sandbox.git).to be_a(Daytona::Git)
      expect(sandbox.computer_use).to be_a(Daytona::ComputerUse)
      expect(sandbox.code_interpreter).to be_a(Daytona::CodeInterpreter)
    end
  end

  describe '#delete' do
    it 'calls sandbox_api.delete_sandbox and refreshes' do
      refreshed_dto = build_sandbox_dto(state: 'destroyed')
      allow(sandbox_api).to receive(:delete_sandbox).with('sandbox-123')
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(refreshed_dto)

      sandbox.delete
      expect(sandbox_api).to have_received(:delete_sandbox).with('sandbox-123')
    end

    it 'sets state to destroyed on 404' do
      error = DaytonaApiClient::ApiError.new(code: 404, message: 'Not found')
      allow(sandbox_api).to receive(:delete_sandbox).and_raise(error)

      sandbox.delete
      expect(sandbox.state).to eq('destroyed')
    end
  end

  describe '#archive' do
    it 'calls archive_sandbox and refreshes' do
      refreshed_dto = build_sandbox_dto(state: 'archived')
      allow(sandbox_api).to receive(:archive_sandbox).with('sandbox-123')
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(refreshed_dto)

      sandbox.archive
      expect(sandbox_api).to have_received(:archive_sandbox).with('sandbox-123')
    end
  end

  describe '#auto_stop_interval=' do
    it 'sets interval via API' do
      allow(sandbox_api).to receive(:set_autostop_interval).with('sandbox-123', 30)

      sandbox.auto_stop_interval = 30
      expect(sandbox.auto_stop_interval).to eq(30)
    end

    it 'raises on negative interval' do
      expect { sandbox.auto_stop_interval = -1 }.to raise_error(Daytona::Sdk::Error, /non-negative/)
    end
  end

  describe '#auto_archive_interval=' do
    it 'sets interval via API' do
      allow(sandbox_api).to receive(:set_auto_archive_interval).with('sandbox-123', 60)

      sandbox.auto_archive_interval = 60
      expect(sandbox.auto_archive_interval).to eq(60)
    end

    it 'raises on negative interval' do
      expect { sandbox.auto_archive_interval = -1 }.to raise_error(Daytona::Sdk::Error, /non-negative/)
    end
  end

  describe '#auto_delete_interval=' do
    it 'sets interval via API' do
      allow(sandbox_api).to receive(:set_auto_delete_interval).with('sandbox-123', 120)

      sandbox.auto_delete_interval = 120
      expect(sandbox.auto_delete_interval).to eq(120)
    end
  end

  describe '#labels=' do
    it 'replaces labels via API' do
      label_response = instance_double(DaytonaApiClient::SandboxLabels, labels: { 'env' => 'prod' })
      allow(sandbox_api).to receive(:replace_labels).and_return(label_response)

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

  describe '#refresh' do
    it 'updates state from API' do
      updated_dto = build_sandbox_dto(state: 'stopped')
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(updated_dto)

      sandbox.refresh
      expect(sandbox.state).to eq('stopped')
    end
  end

  describe '#refresh_activity' do
    it 'calls update_last_activity' do
      allow(sandbox_api).to receive(:update_last_activity).with('sandbox-123')

      expect(sandbox.refresh_activity).to be_nil
    end
  end

  describe '#get_user_home_dir' do
    it 'returns the user home directory' do
      info_api = sandbox.instance_variable_get(:@info_api)
      dir_response = double('DirResponse', dir: '/home/daytona')
      allow(info_api).to receive(:get_user_home_dir).and_return(dir_response)

      expect(sandbox.get_user_home_dir).to eq('/home/daytona')
    end

    it 'wraps errors in Sdk::Error' do
      info_api = sandbox.instance_variable_get(:@info_api)
      allow(info_api).to receive(:get_user_home_dir).and_raise(StandardError, 'connection refused')

      expect { sandbox.get_user_home_dir }.to raise_error(Daytona::Sdk::Error, /Failed to get user home/)
    end
  end

  describe '#get_work_dir' do
    it 'returns the working directory' do
      info_api = sandbox.instance_variable_get(:@info_api)
      dir_response = double('DirResponse', dir: '/workspace')
      allow(info_api).to receive(:get_work_dir).and_return(dir_response)

      expect(sandbox.get_work_dir).to eq('/workspace')
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

  describe '#start' do
    it 'calls start_sandbox and waits for started state' do
      started_dto = build_sandbox_dto(state: DaytonaApiClient::SandboxState::STARTED)
      allow(sandbox_api).to receive(:start_sandbox).with('sandbox-123').and_return(started_dto)
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(started_dto)

      sandbox.start(5)
      expect(sandbox_api).to have_received(:start_sandbox).with('sandbox-123')
    end
  end

  describe '#stop' do
    it 'calls stop_sandbox and waits for stopped state' do
      stopped_dto = build_sandbox_dto(state: DaytonaApiClient::SandboxState::STOPPED)
      allow(sandbox_api).to receive(:stop_sandbox).with('sandbox-123', { force: false })
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(stopped_dto)

      sandbox.stop(5)
      expect(sandbox_api).to have_received(:stop_sandbox).with('sandbox-123', { force: false })
    end
  end

  describe '#resize' do
    it 'raises error when resources is nil' do
      expect { sandbox.resize(nil) }.to raise_error(Daytona::Sdk::Error, /Resources must not be nil/)
    end

    it 'calls resize_sandbox with resource params' do
      resized_dto = build_sandbox_dto(state: DaytonaApiClient::SandboxState::STARTED, cpu: 8)
      allow(sandbox_api).to receive(:resize_sandbox).and_return(resized_dto)
      allow(sandbox_api).to receive(:get_sandbox).with('sandbox-123').and_return(resized_dto)

      resources = Daytona::Resources.new(cpu: 8, memory: 16)
      sandbox.resize(resources, 5)
      expect(sandbox_api).to have_received(:resize_sandbox)
    end
  end

  describe '#create_signed_preview_url' do
    it 'delegates to sandbox_api' do
      signed_url = instance_double(DaytonaApiClient::SignedPortPreviewUrl)
      allow(sandbox_api).to receive(:get_signed_port_preview_url)
        .with('sandbox-123', 3000, { expires_in_seconds: 120 })
        .and_return(signed_url)

      expect(sandbox.create_signed_preview_url(3000, 120)).to eq(signed_url)
    end
  end

  describe '#expire_signed_preview_url' do
    it 'delegates to sandbox_api' do
      allow(sandbox_api).to receive(:expire_signed_port_preview_url)
        .with('sandbox-123', 3000, 'token-val')

      sandbox.expire_signed_preview_url(3000, 'token-val')
      expect(sandbox_api).to have_received(:expire_signed_port_preview_url)
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
end
