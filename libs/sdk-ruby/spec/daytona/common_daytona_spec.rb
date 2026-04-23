# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::CreateSandboxBaseParams do
  describe '#to_h' do
    it 'compacts nil values from the hash representation' do
      params = described_class.new(language: :python, labels: { 'env' => 'test' }, auto_stop_interval: 5)

      expect(params.to_h).to eq(language: :python, labels: { 'env' => 'test' }, auto_stop_interval: 5)
    end

    it 'forces auto_delete_interval to zero for ephemeral sandboxes and warns' do
      expect do
        @params = described_class.new(ephemeral: true, auto_delete_interval: 10)
      end.to output(/auto_delete_interval will be ignored and set to 0/).to_stderr

      expect(@params.auto_delete_interval).to eq(0)
    end

    it 'does not warn when ephemeral sandboxes already delete immediately' do
      expect do
        described_class.new(ephemeral: true, auto_delete_interval: 0)
      end.not_to output.to_stderr
    end
  end
end

RSpec.describe Daytona::CreateSandboxFromImageParams do
  it 'includes image and resource hashes in to_h' do
    params = described_class.new(
      image: 'ruby:3.4',
      resources: Daytona::Resources.new(cpu: 2, memory: 4),
      labels: { 'team' => 'sdk' }
    )

    expect(params.to_h).to eq(image: 'ruby:3.4', resources: { cpu: 2, memory: 4 }, labels: { 'team' => 'sdk' })
  end
end

RSpec.describe Daytona::CreateSandboxFromSnapshotParams do
  it 'includes the snapshot name in to_h' do
    params = described_class.new(snapshot: 'base-snapshot', language: :python)

    expect(params.to_h).to eq(snapshot: 'base-snapshot', language: :python)
  end
end

RSpec.describe Daytona::CodeLanguage do
  it 'defines the supported languages' do
    expect(described_class::ALL).to eq(%i[javascript python typescript])
  end
end
