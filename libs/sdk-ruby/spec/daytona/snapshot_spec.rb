# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::CreateSnapshotParams do
  it 'stores snapshot creation arguments' do
    resources = Daytona::Resources.new(cpu: 2)
    params = described_class.new(name: 'snap', image: 'ruby:3.4', resources: resources, entrypoint: ['/bin/bash'],
                                 region_id: 'eu')

    expect(params.name).to eq('snap')
    expect(params.image).to eq('ruby:3.4')
    expect(params.resources).to eq(resources)
    expect(params.entrypoint).to eq(['/bin/bash'])
    expect(params.region_id).to eq('eu')
  end
end

RSpec.describe Daytona::Snapshot do
  it 'populates all attributes from the DTO' do
    dto = build_snapshot_dto(
      name: 'my-snapshot',
      general: true,
      entrypoint: ['/bin/bash'],
      error_reason: 'none',
      last_used_at: '2025-01-02T00:00:00Z'
    )

    snapshot = described_class.new(dto)

    expect(snapshot.id).to eq('snap-123')
    expect(snapshot.organization_id).to eq('org-1')
    expect(snapshot.general).to be(true)
    expect(snapshot.name).to eq('my-snapshot')
    expect(snapshot.image_name).to eq('ubuntu:22.04')
    expect(snapshot.state).to eq('active')
    expect(snapshot.size).to eq(1024)
    expect(snapshot.entrypoint).to eq(['/bin/bash'])
    expect(snapshot.cpu).to eq(4)
    expect(snapshot.gpu).to eq(0)
    expect(snapshot.mem).to eq(8)
    expect(snapshot.disk).to eq(30)
    expect(snapshot.error_reason).to eq('none')
    expect(snapshot.created_at).to eq('2025-01-01T00:00:00Z')
    expect(snapshot.updated_at).to eq('2025-01-01T00:00:00Z')
    expect(snapshot.last_used_at).to eq('2025-01-02T00:00:00Z')
  end

  it 'builds a snapshot via .from_dto' do
    snapshot = described_class.from_dto(build_snapshot_dto(name: 'from-dto'))

    expect(snapshot).to be_a(described_class)
    expect(snapshot.name).to eq('from-dto')
  end
end
