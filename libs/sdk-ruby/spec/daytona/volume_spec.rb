# frozen_string_literal: true

RSpec.describe Daytona::Volume do
  describe '#initialize' do
    it 'populates attributes from DTO' do
      dto = build_volume_dto(
        id: 'vol-abc',
        name: 'my-volume',
        organization_id: 'org-99',
        state: 'ready',
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-02T00:00:00Z',
        last_used_at: '2025-01-03T00:00:00Z',
        error_reason: nil
      )

      volume = described_class.new(dto)

      expect(volume.id).to eq('vol-abc')
      expect(volume.name).to eq('my-volume')
      expect(volume.organization_id).to eq('org-99')
      expect(volume.state).to eq('ready')
      expect(volume.created_at).to eq('2025-01-01T00:00:00Z')
      expect(volume.updated_at).to eq('2025-01-02T00:00:00Z')
      expect(volume.last_used_at).to eq('2025-01-03T00:00:00Z')
      expect(volume.error_reason).to be_nil
    end

    it 'stores error_reason when present' do
      dto = build_volume_dto(error_reason: 'disk full')
      volume = described_class.new(dto)
      expect(volume.error_reason).to eq('disk full')
    end
  end
end
