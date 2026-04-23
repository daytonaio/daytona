# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::CodeInterpreter do
  class InterpreterWebSocket
    attr_reader :handlers
    attr_reader :sent_messages
    attr_reader :close_calls

    def initialize(script)
      @script = script
      @handlers = {}
      @sent_messages = []
      @close_calls = 0
      @open = true
      @started = false
    end

    def on(event, &block)
      handlers[event] = block
      run_if_ready
    end

    def send(data)
      sent_messages << data
    end

    def close
      @close_calls += 1
      @open = false
    end

    private

    attr_reader :script

    def run_if_ready
      return if @started
      return unless %i[open message error close].all? { |event| handlers.key?(event) }

      @started = true
      handlers[:open].call
      script.each { |event, payload| handlers[event]&.call(payload) }
    end
  end

  let(:api_client) { double('ApiClient', default_headers: { 'Authorization' => 'Bearer token' }) }
  let(:toolbox_api) { instance_double(DaytonaToolboxApiClient::InterpreterApi, api_client: api_client) }
  let(:get_preview_link) do
    proc { |_port| double('PreviewUrl', url: 'https://preview.example.com/root', token: 'tok') }
  end
  let(:interpreter) do
    described_class.new(
      sandbox_id: 'sandbox-123',
      toolbox_api: toolbox_api,
      get_preview_link: get_preview_link
    )
  end

  describe '#run_code' do
    it 'sends execution requests and streams stdout, stderr, and errors' do
      socket = InterpreterWebSocket.new([
                                          [:message, double(data: '{"type":"stdout","text":"hello"}')],
                                          [:message, double(data: '{"type":"stderr","text":"warn"}')],
                                          [:message,
                                           double(data: '{"type":"error","name":"ValueError","value":"bad","traceback":"trace"}')],
                                          [:message, double(data: '{"type":"control","text":"completed"}')],
                                          [:close, double(code: 1000, reason: 'normal')]
                                        ])
      context = double('InterpreterContext', id: 'ctx-1')
      stdout = []
      stderr = []
      errors = []
      allow(WebSocket::Client::Simple).to receive(:connect).and_return(socket)

      result = interpreter.run_code(
        'print("hello")',
        context: context,
        envs: { 'DEBUG' => '1' },
        timeout: 5,
        on_stdout: ->(msg) { stdout << msg.output },
        on_stderr: ->(msg) { stderr << msg.output },
        on_error: ->(error) { errors << error.name }
      )

      expect(result.stdout).to eq('hello')
      expect(result.stderr).to eq('warn')
      expect(result.error.name).to eq('ValueError')
      expect(stdout).to eq(['hello'])
      expect(stderr).to eq(['warn'])
      expect(errors).to eq(['ValueError'])
      expect(socket.sent_messages.first).to include('"contextId":"ctx-1"')
      expect(socket.sent_messages.first).to include('"timeout":5')
      expect(socket.sent_messages.first).to include('"envs":{"DEBUG":"1"}')
      expect(WebSocket::Client::Simple).to have_received(:connect).with(
        'wss://preview.example.com/process/interpreter/execute',
        headers: hash_including('X-Daytona-Preview-Token' => 'tok')
      )
    end

    it 'raises TimeoutError when the websocket closes with the timeout code' do
      socket = InterpreterWebSocket.new([[:close,
                                          double(code: described_class::WEBSOCKET_TIMEOUT_CODE, reason: 'timeout')]])
      allow(WebSocket::Client::Simple).to receive(:connect).and_return(socket)

      expect { interpreter.run_code('sleep(10)', timeout: 1) }
        .to raise_error(Daytona::Sdk::TimeoutError, /Execution timed out/)
    end

    it 'wraps websocket errors as SDK errors' do
      socket = InterpreterWebSocket.new([[:error, StandardError.new('socket boom')]])
      allow(WebSocket::Client::Simple).to receive(:connect).and_return(socket)
      allow(interpreter).to receive(:sleep)
      allow(Time).to receive(:now).and_return(Time.now, Time.now + 10)

      expect { interpreter.run_code('print(1)') }.to raise_error(Daytona::Sdk::Error, /WebSocket error: socket boom/)
    end
  end

  describe '#create_context' do
    it 'creates an interpreter context' do
      ctx = double('InterpreterContext', id: 'ctx-1')
      allow(toolbox_api).to receive(:create_interpreter_context).and_return(ctx)

      expect(interpreter.create_context(cwd: '/workspace')).to eq(ctx)
      expect(toolbox_api).to have_received(:create_interpreter_context) do |req|
        expect(req.cwd).to eq('/workspace')
      end
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:create_interpreter_context).and_raise(StandardError, 'err')

      expect do
        interpreter.create_context
      end.to raise_error(Daytona::Sdk::Error, /Failed to create interpreter context: err/)
    end
  end

  describe '#list_contexts' do
    it 'returns array of contexts' do
      ctx = double('InterpreterContext', id: 'ctx-1')
      allow(toolbox_api).to receive(:list_interpreter_contexts).and_return(double('ListResponse', contexts: [ctx]))

      expect(interpreter.list_contexts).to eq([ctx])
    end

    it 'returns empty array when contexts are nil' do
      allow(toolbox_api).to receive(:list_interpreter_contexts).and_return(double('ListResponse', contexts: nil))

      expect(interpreter.list_contexts).to eq([])
    end

    it 'wraps errors' do
      allow(toolbox_api).to receive(:list_interpreter_contexts).and_raise(StandardError, 'err')

      expect do
        interpreter.list_contexts
      end.to raise_error(Daytona::Sdk::Error, /Failed to list interpreter contexts: err/)
    end
  end

  describe '#delete_context' do
    it 'deletes a context and returns nil' do
      ctx = double('InterpreterContext', id: 'ctx-1')
      allow(toolbox_api).to receive(:delete_interpreter_context).with('ctx-1')

      expect(interpreter.delete_context(ctx)).to be_nil
    end

    it 'wraps errors' do
      ctx = double('InterpreterContext', id: 'ctx-1')
      allow(toolbox_api).to receive(:delete_interpreter_context).and_raise(StandardError, 'err')

      expect do
        interpreter.delete_context(ctx)
      end.to raise_error(Daytona::Sdk::Error, /Failed to delete interpreter context: err/)
    end
  end

  describe 'private helpers' do
    let(:result) { Daytona::ExecutionResult.new }

    it 'ignores empty messages' do
      expect(interpreter.send(:handle_message, '', result, nil, nil, nil)).to be_nil
      expect(interpreter.send(:handle_message, nil, result, nil, nil, nil)).to be_nil
    end

    it 'ignores malformed JSON payloads' do
      expect { interpreter.send(:handle_message, '{bad json', result, nil, nil, nil) }.not_to raise_error
    end

    it 'pushes completion for completed control messages' do
      queue = Queue.new

      interpreter.send(:handle_message, '{"type":"control","text":"completed"}', result, nil, nil, nil, queue)

      expect(queue.pop[:type]).to eq(:completed)
    end

    it 'returns nil for normal close events' do
      expect(interpreter.send(:handle_close, double(code: 1000, reason: 'ok'))).to be_nil
      expect(interpreter.send(:handle_close, nil)).to be_nil
    end

    it 'formats unexpected close events' do
      expect(interpreter.send(:handle_close, double(code: 1006, reason: 'lost'))).to eq('lost (close code 1006)')
    end
  end
end
