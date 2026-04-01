# frozen_string_literal: true

RSpec.describe Daytona::Util do
  describe '.demux' do
    let(:stdout_prefix) { "\x01\x01\x01" }
    let(:stderr_prefix) { "\x02\x02\x02" }

    it 'separates stdout and stderr from multiplexed output' do
      input = "#{stdout_prefix}hello#{stderr_prefix}error"
      stdout, stderr = described_class.demux(input)

      expect(stdout).to eq('hello')
      expect(stderr).to eq('error')
    end

    it 'handles stdout-only output' do
      input = "#{stdout_prefix}hello world"
      stdout, stderr = described_class.demux(input)

      expect(stdout).to eq('hello world')
      expect(stderr).to eq('')
    end

    it 'handles stderr-only output' do
      input = "#{stderr_prefix}error only"
      stdout, stderr = described_class.demux(input)

      expect(stdout).to eq('')
      expect(stderr).to eq('error only')
    end

    it 'handles empty input' do
      stdout, stderr = described_class.demux('')
      expect(stdout).to eq('')
      expect(stderr).to eq('')
    end

    it 'handles interleaved stdout and stderr' do
      input = "#{stdout_prefix}line1#{stderr_prefix}err1#{stdout_prefix}line2"
      stdout, stderr = described_class.demux(input)

      expect(stdout).to eq('line1line2')
      expect(stderr).to eq('err1')
    end
  end
end
