# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::Util do
  describe '.demux' do
    let(:stdout_prefix) { "\x01\x01\x01" }
    let(:stderr_prefix) { "\x02\x02\x02" }

    it 'separates stdout and stderr from multiplexed output' do
      stdout, stderr = described_class.demux("#{stdout_prefix}hello#{stderr_prefix}error")

      expect(stdout).to eq('hello')
      expect(stderr).to eq('error')
    end

    it 'handles stdout-only output' do
      stdout, stderr = described_class.demux("#{stdout_prefix}hello world")

      expect(stdout).to eq('hello world')
      expect(stderr).to eq('')
    end

    it 'handles stderr-only output' do
      stdout, stderr = described_class.demux("#{stderr_prefix}error only")

      expect(stdout).to eq('')
      expect(stderr).to eq('error only')
    end

    it 'handles empty input' do
      stdout, stderr = described_class.demux('')

      expect(stdout).to eq('')
      expect(stderr).to eq('')
    end

    it 'handles interleaved stdout and stderr chunks' do
      stdout, stderr = described_class.demux("#{stdout_prefix}line1#{stderr_prefix}err1#{stdout_prefix}line2")

      expect(stdout).to eq('line1line2')
      expect(stderr).to eq('err1')
    end

    it 'ignores text before the first stream prefix' do
      stdout, stderr = described_class.demux("orphan#{stdout_prefix}hello")

      expect(stdout).to eq('hello')
      expect(stderr).to eq('')
    end
  end

  describe '.stream_async' do
    it 'streams chunks asynchronously to the callback' do
      http = double('Http')
      response = double('Response')
      captured_request = nil
      received = []

      allow(Net::HTTP).to receive(:start).with('example.com', 443, use_ssl: true).and_yield(http)
      allow(http).to receive(:request) do |request, &block|
        captured_request = request
        block.call(response)
      end
      allow(response).to receive(:read_body).and_yield('first').and_yield('second')

      thread = described_class.stream_async(
        uri: URI('https://example.com/logs'),
        headers: { 'Authorization' => 'Bearer token' },
        on_chunk: ->(chunk) { received << chunk }
      )
      thread.join

      expect(captured_request['Authorization']).to eq('Bearer token')
      expect(received).to eq(%w[first second])
    end

    it 'swallows read timeouts' do
      allow(Net::HTTP).to receive(:start).and_raise(Net::ReadTimeout)

      expect { described_class.stream_async(uri: URI('https://example.com/logs'), on_chunk: ->(_chunk) {}).join }
        .not_to raise_error
    end
  end
end
