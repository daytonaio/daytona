# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::PtyHandle do
  class FakePtyWebSocket
    attr_reader :handlers
    attr_reader :sent_messages
    attr_reader :close_calls

    def initialize(open: true)
      @open = open
      @handlers = Hash.new { |hash, key| hash[key] = [] }
      @sent_messages = []
      @close_calls = 0
    end

    def on(event, &block)
      handlers[event] << block
    end

    def emit(event, payload = nil)
      handlers[event].each { |handler| payload.nil? ? handler.call : handler.call(payload) }
    end

    def open?
      @open
    end

    def send(data)
      @sent_messages << data
    end

    def close
      @close_calls += 1
      @open = false
    end
  end

  let(:websocket) { FakePtyWebSocket.new }
  let(:resize_handler) { proc { |pty_size| "#{pty_size.cols}x#{pty_size.rows}" } }
  let(:kill_handler) { proc { :killed } }
  let(:handle) do
    described_class.new(websocket, session_id: 'pty-1', handle_resize: resize_handler, handle_kill: kill_handler)
  end

  describe Daytona::PtySize do
    it 'stores rows and columns' do
      size = described_class.new(rows: 24, cols: 80)

      expect(size.rows).to eq(24)
      expect(size.cols).to eq(80)
    end
  end

  describe Daytona::PtyResult do
    it 'stores exit code and error' do
      result = described_class.new(exit_code: 2, error: 'boom')

      expect(result.exit_code).to eq(2)
      expect(result.error).to eq('boom')
    end
  end

  describe '#connected?' do
    it 'reflects the websocket open state' do
      expect(handle.connected?).to be(true)

      websocket.close

      expect(handle.connected?).to be(false)
    end
  end

  describe '#wait_for_connection' do
    it 'returns when a connected control message is received' do
      Thread.new do
        sleep 0.01
        websocket.emit(:message, double(type: :text, data: '{"type":"control","status":"connected"}'))
      end

      expect { handle.wait_for_connection(timeout: 0.2) }.not_to raise_error
    end

    it 'raises when the timeout is exceeded' do
      allow(handle).to receive(:sleep)

      expect { handle.wait_for_connection(timeout: 0) }
        .to raise_error(Daytona::Sdk::Error, /PTY connection timeout/)
    end
  end

  describe '#send_input' do
    it 'sends input when the websocket is open' do
      handle.send_input("ls\n")

      expect(websocket.sent_messages).to eq(["ls\n"])
    end

    it 'raises when the websocket is closed' do
      websocket.close

      expect { handle.send_input('ls') }.to raise_error(Daytona::Sdk::Error, /not connected/)
    end
  end

  describe '#resize' do
    it 'delegates to the resize handler' do
      result = handle.resize(Daytona::PtySize.new(rows: 30, cols: 120))

      expect(result).to eq('120x30')
    end

    it 'raises when no resize handler is available' do
      bare_handle = described_class.new(websocket, session_id: 'pty-1')

      expect { bare_handle.resize(Daytona::PtySize.new(rows: 10, cols: 20)) }
        .to raise_error(Daytona::Sdk::Error, /No resize handler available/)
    end
  end

  describe '#kill' do
    it 'delegates to the kill handler' do
      expect(handle.kill).to eq(:killed)
    end

    it 'raises when no kill handler is available' do
      bare_handle = described_class.new(websocket, session_id: 'pty-1')

      expect { bare_handle.kill }.to raise_error(Daytona::Sdk::Error, /No kill handler available/)
    end
  end

  describe '#wait' do
    before do
      handle.instance_variable_set(:@status, Daytona::PtyHandle::Status::CONNECTED)
    end

    it 'returns nil when called before the session is connected' do
      disconnected = described_class.new(FakePtyWebSocket.new, session_id: 'pty-2')

      expect(disconnected.wait(timeout: 0)).to be_nil
    end

    it 'returns the exit code and error from the close message' do
      Thread.new do
        sleep 0.01
        handle.send(:process_websocket_close_message, double(data: '{"exitCode":7,"exitReason":"terminated"}'))
      end

      result = handle.wait(timeout: 0.2)

      expect(result.exit_code).to eq(7)
      expect(result.error).to eq('terminated')
    end

    it 'yields streamed text data to the callback' do
      chunks = []

      Thread.new do
        sleep 0.01
        handle.send(:process_websocket_data_message, 'hello')
        handle.send(:process_websocket_close_message, double(data: '{"exitCode":0}'))
      end

      handle.wait(timeout: 0.2) { |data| chunks << data }

      expect(chunks).to eq(['hello'])
    end
  end

  describe '#each' do
    before do
      websocket.emit(:message, double(type: :text, data: '{"type":"control","status":"connected"}'))
    end

    it 'yields buffered data until the websocket closes' do
      chunks = []

      Thread.new do
        sleep 0.01
        websocket.emit(:message, double(type: :text, data: 'chunk-1'))
        websocket.emit(:message, double(type: :text, data: 'chunk-2'))
        websocket.emit(:message, double(type: :close, data: '{"exitCode":0}'))
      end

      handle.each { |data| chunks << data }

      expect(chunks).to eq(%w[chunk-1 chunk-2])
    end
  end

  describe '#disconnect' do
    it 'closes the websocket' do
      handle.disconnect

      expect(websocket.close_calls).to eq(1)
    end
  end

  describe 'websocket event handling' do
    it 'marks the handle as errored when websocket error fires' do
      handle.send(:on_websocket_error, StandardError.new('boom'))

      expect(handle.send(:status)).to eq('error')
    end

    it 'sets error status and message on control error messages' do
      handle.send(:process_websocket_control_message, { status: 'error', error: 'bad auth' })

      expect(handle.send(:status)).to eq('error')
      expect(handle.error).to eq('bad auth')
    end

    it 'sets default error message when control error has no error field' do
      handle.send(:process_websocket_control_message, { status: 'error' })

      expect(handle.send(:status)).to eq('error')
      expect(handle.error).to eq('Unknown connection error')
    end

    it 'closes websocket and raises on invalid control status' do
      expect do
        handle.send(:process_websocket_control_message, { status: 'weird' })
      end.to raise_error(Daytona::Sdk::Error, /Received invalid control message status: weird/)

      expect(websocket.close_calls).to eq(1)
    end

    it 'forwards non-json text messages to observers' do
      chunks = []
      handle.add_observer(proc { |data| chunks << data }, :call)

      websocket.emit(:message, double(type: :text, data: 'raw text'))

      expect(chunks).to eq(['raw text'])
    end

    it 'ignores malformed close payloads' do
      websocket.emit(:message, double(type: :close, data: 'not-json'))

      expect(handle.exit_code).to be_nil
      expect(handle.error).to be_nil
    end
  end
end
