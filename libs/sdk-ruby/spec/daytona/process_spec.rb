# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::Process do
  class ScriptedWebSocket
    attr_reader :handlers
    attr_reader :sent_messages
    attr_reader :close_calls

    def initialize(script = [])
      @script = script
      @handlers = {}
      @sent_messages = []
      @close_calls = 0
      @open = true
      @started = false
    end

    def on(event, &block)
      handlers[event] = block
      run_script_if_ready
    end

    def send(data)
      sent_messages << data
    end

    def close
      @close_calls += 1
      @open = false
    end

    def open?
      @open
    end

    private

    attr_reader :script

    def run_script_if_ready
      return if @started
      return unless %i[open message error close].all? { |event| handlers.key?(event) }

      @started = true
      handlers[:open].call
      script.each { |event, payload| handlers[event]&.call(payload) }
    end
  end

  class PassiveWebSocket
    attr_reader :handlers
    attr_reader :close_calls

    def initialize
      @handlers = {}
      @close_calls = 0
      @open = true
    end

    def on(event, &block)
      handlers[event] = block
    end

    def emit(event, payload = nil)
      handlers[event]&.call(payload)
    end

    def close
      @close_calls += 1
      @open = false
    end

    def open?
      @open
    end
  end

  let(:api_client) { double('ApiClient', default_headers: { 'Authorization' => 'Bearer token' }) }
  let(:toolbox_api) { instance_double(DaytonaToolboxApiClient::ProcessApi, api_client: api_client) }
  let(:preview_link) { double('PreviewUrl', url: 'https://preview.example.com/base', token: 'tok') }
  let(:get_preview_link) { proc { |_port| preview_link } }

  let(:process) do
    described_class.new(
      sandbox_id: 'sandbox-123',
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
      expect(response.artifacts.stdout).to eq("Hello\n")
    end

    it 'passes cwd, timeout, and env variables through as envs' do
      allow(toolbox_api).to receive(:execute_command).and_return(exec_response)

      process.exec(command: 'echo $MY_VAR', cwd: '/workspace', env: { 'MY_VAR' => 'value' }, timeout: 5)

      expect(toolbox_api).to have_received(:execute_command) do |req|
        expect(req.command).to eq('echo $MY_VAR')
        expect(req.cwd).to eq('/workspace')
        expect(req.envs).to eq({ 'MY_VAR' => 'value' })
        expect(req.timeout).to eq(5)
      end
    end

    it 'normalizes empty env hashes to nil' do
      allow(toolbox_api).to receive(:execute_command).and_return(exec_response)

      process.exec(command: 'true', env: {})

      expect(toolbox_api).to have_received(:execute_command) do |req|
        expect(req.envs).to be_nil
      end
    end

    it 'normalizes empty output to an empty string' do
      allow(toolbox_api).to receive(:execute_command).and_return(double('ExecResponse', exit_code: 0, result: nil))

      response = process.exec(command: 'true')

      expect(response.result).to eq('')
      expect(response.artifacts.stdout).to eq('')
    end
  end

  describe '#code_run' do
    let(:chart_element) { double('ChartElement', label: 'series', points: [[1, 2]]) }
    let(:chart_dto) do
      double('Chart', type: Daytona::ChartType::LINE, title: 'chart', png: 'png', elements: [chart_element],
                      x_label: 'x', y_label: 'y', x_ticks: nil, y_ticks: nil, x_tick_labels: nil,
                      y_tick_labels: nil, x_scale: nil, y_scale: nil)
    end

    it 'calls ProcessApi#code_run with language, argv, env, and timeout' do
      api_response = double('CodeRunResponse', exit_code: 0, result: 'output', artifacts: double(charts: [chart_dto]))
      allow(toolbox_api).to receive(:code_run).and_return(api_response)

      params = Daytona::CodeRunParams.new(argv: ['-m', 'app'], env: { 'DEBUG' => '1' })
      result = process.code_run(code: 'print("hello")', params: params, timeout: 9)

      expect(result).to be_a(Daytona::ExecuteResponse)
      expect(result.artifacts.charts.first).to be_a(Daytona::LineChart)
      expect(toolbox_api).to have_received(:code_run) do |req|
        expect(req.code).to eq('print("hello")')
        expect(req.language).to eq('python')
        expect(req.argv).to eq(['-m', 'app'])
        expect(req.envs).to eq({ 'DEBUG' => '1' })
        expect(req.timeout).to eq(9)
      end
    end

    it 'handles missing artifacts' do
      allow(toolbox_api).to receive(:code_run).and_return(double('CodeRunResponse', exit_code: 0, result: 'ok',
                                                                                    artifacts: nil))

      result = process.code_run(code: 'print(1)')

      expect(result.artifacts.charts).to eq([])
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

  describe '#get_entrypoint_session' do
    it 'delegates to toolbox_api' do
      session = double('Session')
      allow(toolbox_api).to receive(:get_entrypoint_session).and_return(session)

      expect(process.get_entrypoint_session).to eq(session)
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
    it 'executes a command and returns SessionExecuteResponse' do
      api_response = double('ApiResponse', cmd_id: 'cmd-1', output: 'hello', stdout: nil, stderr: nil, exit_code: 0)
      allow(toolbox_api).to receive(:session_execute_command).and_return(api_response)

      req = Daytona::SessionExecuteRequest.new(command: 'echo hello', run_async: true, suppress_input_echo: true)
      result = process.execute_session_command(session_id: 'sess-1', req: req)

      expect(result).to be_a(Daytona::SessionExecuteResponse)
      expect(result.stdout).to eq('')
      expect(result.stderr).to eq('')
      expect(toolbox_api).to have_received(:session_execute_command) do |session_id, request|
        expect(session_id).to eq('sess-1')
        expect(request.command).to eq('echo hello')
        expect(request.run_async).to be(true)
        expect(request.suppress_input_echo).to be(true)
      end
    end
  end

  describe '#get_session_command_logs' do
    it 'returns parsed SessionCommandLogsResponse' do
      raw = double('CommandLogs', output: 'all', stdout: 'out', stderr: 'err')
      allow(toolbox_api).to receive(:get_session_command_logs).with('sess-1', 'cmd-1').and_return(raw)

      result = process.get_session_command_logs(session_id: 'sess-1', command_id: 'cmd-1')

      expect(result.stdout).to eq('out')
      expect(result.stderr).to eq('err')
    end
  end

  describe '#get_entrypoint_logs' do
    it 'returns parsed entrypoint logs' do
      raw = double('EntrypointLogs', output: 'log data', stdout: 'log data', stderr: '')
      allow(toolbox_api).to receive(:get_entrypoint_logs).and_return(raw)

      result = process.get_entrypoint_logs

      expect(result).to be_a(Daytona::SessionCommandLogsResponse)
      expect(result.stdout).to eq('log data')
    end
  end

  describe '#send_session_command_input' do
    it 'sends input to a command' do
      allow(toolbox_api).to receive(:send_input)

      process.send_session_command_input(session_id: 'sess-1', command_id: 'cmd-1', data: "yes\n")

      expect(toolbox_api).to have_received(:send_input) do |session_id, command_id, request|
        expect(session_id).to eq('sess-1')
        expect(command_id).to eq('cmd-1')
        expect(request.data).to eq("yes\n")
      end
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

  describe '#get_session_command_logs_async' do
    it 'streams stdout and stderr chunks until close' do
      socket = PassiveWebSocket.new
      allow(WebSocket::Client::Simple).to receive(:connect).and_return(socket)
      stdout_chunks = []
      stderr_chunks = []

      thread = Thread.new do
        process.get_session_command_logs_async(
          session_id: 'sess-1',
          command_id: 'cmd-1',
          on_stdout: ->(chunk) { stdout_chunks << chunk },
          on_stderr: ->(chunk) { stderr_chunks << chunk }
        )
      end

      sleep 0.02
      socket.emit(:message, double(type: :text, data: "\x01\x01\x01hello\x02\x02\x02oops"))
      socket.emit(:close)
      thread.join

      expect(stdout_chunks).to eq(['hello'])
      expect(stderr_chunks).to eq(['oops'])
      expect(WebSocket::Client::Simple).to have_received(:connect).with(
        'wss://preview.example.com/process/session/sess-1/command/cmd-1/logs?follow=true',
        headers: hash_including('X-Daytona-Preview-Token' => 'tok')
      )
    end
  end

  describe '#get_entrypoint_logs_async' do
    it 'streams entrypoint logs until close' do
      socket = PassiveWebSocket.new
      allow(WebSocket::Client::Simple).to receive(:connect).and_return(socket)
      stdout_chunks = []

      thread = Thread.new do
        process.get_entrypoint_logs_async(
          on_stdout: ->(chunk) { stdout_chunks << chunk },
          on_stderr: ->(_chunk) {}
        )
      end

      sleep 0.02
      socket.emit(:message, double(type: :text, data: "\x01\x01\x01entrypoint"))
      socket.emit(:message, double(type: :close, data: ''))
      thread.join

      expect(stdout_chunks).to eq(['entrypoint'])
      expect(WebSocket::Client::Simple).to have_received(:connect).with(
        'wss://preview.example.com/process/session/entrypoint/logs?follow=true',
        headers: hash_including('X-Daytona-Preview-Token' => 'tok')
      )
    end
  end

  describe '#create_pty_session' do
    it 'creates a PTY session and connects to it' do
      response = double('PtyCreateResponse', session_id: 'pty-1')
      handle = instance_double(Daytona::PtyHandle)
      allow(toolbox_api).to receive(:create_pty_session).and_return(response)
      allow(process).to receive(:connect_pty_session).with('pty-1').and_return(handle)

      result = process.create_pty_session(
        id: 'pty-1',
        cwd: '/workspace',
        envs: { 'TERM' => 'xterm' },
        pty_size: Daytona::PtySize.new(rows: 24, cols: 80)
      )

      expect(result).to eq(handle)
      expect(toolbox_api).to have_received(:create_pty_session) do |req|
        expect(req.id).to eq('pty-1')
        expect(req.cwd).to eq('/workspace')
        expect(req.envs).to eq({ 'TERM' => 'xterm' })
        expect(req.rows).to eq(24)
        expect(req.cols).to eq(80)
        expect(req.lazy_start).to be(true)
      end
    end
  end

  describe '#connect_pty_session' do
    it 'connects via websocket and waits for the PTY connection' do
      socket = double('PtySocket')
      handle = instance_double(Daytona::PtyHandle)
      allow(WebSocket::Client::Simple).to receive(:connect).and_return(socket)
      allow(Daytona::PtyHandle).to receive(:new).and_return(handle)
      allow(handle).to receive(:wait_for_connection)

      result = process.connect_pty_session('pty-1')

      expect(result).to eq(handle)
      expect(WebSocket::Client::Simple).to have_received(:connect).with(
        'wss://preview.example.com/process/pty/pty-1/connect',
        headers: hash_including('X-Daytona-Preview-Token' => 'tok')
      )
      expect(Daytona::PtyHandle).to have_received(:new).with(
        socket,
        session_id: 'pty-1',
        handle_resize: instance_of(Proc),
        handle_kill: instance_of(Proc)
      )
      expect(handle).to have_received(:wait_for_connection)
    end
  end

  describe '#resize_pty_session' do
    it 'resizes a PTY session' do
      result = double('PtySessionInfo')
      allow(toolbox_api).to receive(:resize_pty_session).and_return(result)

      response = process.resize_pty_session('pty-1', Daytona::PtySize.new(rows: 30, cols: 120))

      expect(response).to eq(result)
      expect(toolbox_api).to have_received(:resize_pty_session) do |session_id, request|
        expect(session_id).to eq('pty-1')
        expect(request.rows).to eq(30)
        expect(request.cols).to eq(120)
      end
    end
  end

  describe '#delete_pty_session' do
    it 'deletes a PTY session' do
      allow(toolbox_api).to receive(:delete_pty_session).with('pty-1')

      process.delete_pty_session('pty-1')

      expect(toolbox_api).to have_received(:delete_pty_session).with('pty-1')
    end
  end

  describe '#list_pty_sessions' do
    it 'unwraps sessions from the PtyListResponse' do
      sessions = [double('PtySessionInfo')]
      response = double('PtyListResponse', sessions: sessions)
      allow(toolbox_api).to receive(:list_pty_sessions).and_return(response)

      expect(process.list_pty_sessions).to eq(sessions)
    end

    it 'returns an empty array when sessions is nil' do
      response = double('PtyListResponse', sessions: nil)
      allow(toolbox_api).to receive(:list_pty_sessions).and_return(response)

      expect(process.list_pty_sessions).to eq([])
    end
  end

  describe '#get_pty_session_info' do
    it 'returns PTY session details' do
      session = double('PtySessionInfo')
      allow(toolbox_api).to receive(:get_pty_session).with('pty-1').and_return(session)

      expect(process.get_pty_session_info('pty-1')).to eq(session)
    end
  end
end
