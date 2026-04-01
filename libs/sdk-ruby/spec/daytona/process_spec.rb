# frozen_string_literal: true

RSpec.describe Daytona::Process do
  let(:toolbox_api) { instance_double(DaytonaToolboxApiClient::ProcessApi) }
  let(:code_toolbox) { Daytona::SandboxPythonCodeToolbox.new }
  let(:get_preview_link) { proc { |_port| double('PreviewUrl', url: 'https://preview.example.com', token: 'tok') } }

  let(:process) do
    described_class.new(
      sandbox_id: 'sandbox-123',
      code_toolbox: code_toolbox,
      toolbox_api: toolbox_api,
      get_preview_link: get_preview_link
    )
  end

  describe '#exec' do
    let(:exec_response) { double('ExecResponse', exit_code: 0, result: "Hello\n") }

    it 'executes a command and returns ExecuteResponse' do
      allow(toolbox_api).to receive(:execute_command).and_return(exec_response)

      response = process.exec(command: 'echo Hello')
      expect(response).to be_a(Daytona::ExecuteResponse)
      expect(response.exit_code).to eq(0)
      expect(response.result).to eq("Hello\n")
    end

    it 'handles env variables with base64 encoding' do
      allow(toolbox_api).to receive(:execute_command).and_return(exec_response)

      process.exec(command: 'echo $MY_VAR', env: { 'MY_VAR' => 'value' })
      expect(toolbox_api).to have_received(:execute_command) do |req|
        expect(req.command).to include('export MY_VAR=')
      end
    end

    it 'raises ArgumentError on invalid env var names' do
      expect { process.exec(command: 'echo', env: { '123bad' => 'v' }) }
        .to raise_error(ArgumentError, /Invalid environment variable name/)
    end

    it 'parses artifact lines from output' do
      artifact_line = 'dtn_artifact_k39fd2:{"type":"chart","value":{"chart_type":"bar"}}'
      response_with_artifact = double('ExecResponse', exit_code: 0, result: "line1\n#{artifact_line}\nline2")
      allow(toolbox_api).to receive(:execute_command).and_return(response_with_artifact)

      result = process.exec(command: 'run_plot')
      expect(result.artifacts).to be_a(Daytona::ExecutionArtifacts)
      expect(result.artifacts.stdout).to eq("line1\nline2")
    end
  end

  describe '#code_run' do
    it 'delegates to exec with code_toolbox command' do
      exec_response = double('ExecResponse', exit_code: 0, result: 'output')
      allow(toolbox_api).to receive(:execute_command).and_return(exec_response)

      result = process.code_run(code: 'print("hello")')
      expect(result).to be_a(Daytona::ExecuteResponse)
    end
  end

  describe '#create_session' do
    it 'delegates to toolbox_api' do
      allow(toolbox_api).to receive(:create_session)

      process.create_session('my-session')
      expect(toolbox_api).to have_received(:create_session) do |req|
        expect(req.session_id).to eq('my-session')
      end
    end
  end

  describe '#get_session' do
    it 'returns session from toolbox_api' do
      session = double('Session', session_id: 'my-session')
      allow(toolbox_api).to receive(:get_session).with('my-session').and_return(session)

      expect(process.get_session('my-session')).to eq(session)
    end
  end

  describe '#get_session_command' do
    it 'delegates with session_id and command_id' do
      cmd = double('Command', id: 'cmd-1')
      allow(toolbox_api).to receive(:get_session_command).with('sess-1', 'cmd-1').and_return(cmd)

      expect(process.get_session_command(session_id: 'sess-1', command_id: 'cmd-1')).to eq(cmd)
    end
  end

  describe '#execute_session_command' do
    it 'executes command and returns SessionExecuteResponse' do
      api_response = double('ApiResponse', cmd_id: 'cmd-1', output: "\x01\x01\x01hello", exit_code: 0)
      allow(toolbox_api).to receive(:session_execute_command).and_return(api_response)

      req = Daytona::SessionExecuteRequest.new(command: 'echo hello')
      result = process.execute_session_command(session_id: 'sess-1', req: req)

      expect(result).to be_a(Daytona::SessionExecuteResponse)
      expect(result.cmd_id).to eq('cmd-1')
      expect(result.stdout).to eq('hello')
    end
  end

  describe '#get_session_command_logs' do
    it 'returns parsed SessionCommandLogsResponse' do
      raw = "\x01\x01\x01stdout content\x02\x02\x02stderr content"
      allow(toolbox_api).to receive(:get_session_command_logs).with('sess-1', 'cmd-1').and_return(raw)

      result = process.get_session_command_logs(session_id: 'sess-1', command_id: 'cmd-1')
      expect(result).to be_a(Daytona::SessionCommandLogsResponse)
      expect(result.stdout).to eq('stdout content')
      expect(result.stderr).to eq('stderr content')
    end
  end

  describe '#list_sessions' do
    it 'delegates to toolbox_api' do
      sessions = [double('Session')]
      allow(toolbox_api).to receive(:list_sessions).and_return(sessions)

      expect(process.list_sessions).to eq(sessions)
    end
  end

  describe '#delete_session' do
    it 'delegates to toolbox_api' do
      allow(toolbox_api).to receive(:delete_session).with('sess-1')

      process.delete_session('sess-1')
      expect(toolbox_api).to have_received(:delete_session).with('sess-1')
    end
  end

  describe '#get_entrypoint_session' do
    it 'delegates to toolbox_api' do
      session = double('Session')
      allow(toolbox_api).to receive(:get_entrypoint_session).and_return(session)

      expect(process.get_entrypoint_session).to eq(session)
    end
  end

  describe '#get_entrypoint_logs' do
    it 'returns parsed log response' do
      raw = "\x01\x01\x01log data"
      allow(toolbox_api).to receive(:get_entrypoint_logs).and_return(raw)

      result = process.get_entrypoint_logs
      expect(result).to be_a(Daytona::SessionCommandLogsResponse)
      expect(result.stdout).to eq('log data')
    end
  end

  describe '#send_session_command_input' do
    it 'sends input to a command' do
      allow(toolbox_api).to receive(:send_input)

      process.send_session_command_input(session_id: 'sess-1', command_id: 'cmd-1', data: 'yes\n')
      expect(toolbox_api).to have_received(:send_input).with('sess-1', 'cmd-1', anything)
    end
  end
end
