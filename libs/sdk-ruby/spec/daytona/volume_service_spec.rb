# Copyright Daytona Platforms Inc.
# SPDX-License-Identifier: Apache-2.0

# frozen_string_literal: true

RSpec.describe Daytona::VolumeService do
  let(:volumes_api) { instance_double(DaytonaApiClient::VolumesApi) }
  let(:service) { described_class.new(volumes_api) }

  describe '#create' do
    it 'creates a volume and returns a Volume model' do
      dto = build_volume_dto(name: 'my-volume')
      allow(volumes_api).to receive(:create_volume).and_return(dto)

      volume = service.create('my-volume')

      expect(volume).to be_a(Daytona::Volume)
      expect(volume.name).to eq('my-volume')
      expect(volumes_api).to have_received(:create_volume) do |request|
        expect(request.name).to eq('my-volume')
      end
    end
  end

  describe '#delete' do
    it 'deletes a volume by id' do
      volume = Daytona::Volume.new(build_volume_dto)
      allow(volumes_api).to receive(:delete_volume).with('vol-123')

      service.delete(volume)

      expect(volumes_api).to have_received(:delete_volume).with('vol-123')
    end
  end

  describe '#get' do
    it 'gets a volume by name' do
      dto = build_volume_dto(name: 'my-vol')
      allow(volumes_api).to receive(:get_volume_by_name).with('my-vol').and_return(dto)

      volume = service.get('my-vol')

      expect(volume).to be_a(Daytona::Volume)
      expect(volume.name).to eq('my-vol')
    end

    it 'creates the volume when not found and create: true' do
      error = DaytonaApiClient::ApiError.new(code: 404, message: 'Volume with name missing-vol not found')
      created_dto = build_volume_dto(name: 'missing-vol')
      allow(volumes_api).to receive(:get_volume_by_name).and_raise(error)
      allow(volumes_api).to receive(:create_volume).and_return(created_dto)

      volume = service.get('missing-vol', create: true)

      expect(volume).to be_a(Daytona::Volume)
      expect(volume.name).to eq('missing-vol')
    end

    it 're-raises not found errors when create is false' do
      error = DaytonaApiClient::ApiError.new(code: 404, message: 'Volume with name x not found')
      allow(volumes_api).to receive(:get_volume_by_name).and_raise(error)

      expect { service.get('x') }.to raise_error(DaytonaApiClient::ApiError)
    end

    it 're-raises unrelated 404 errors even when create is true' do
      error = DaytonaApiClient::ApiError.new(code: 404, message: 'some other 404')
      allow(volumes_api).to receive(:get_volume_by_name).and_raise(error)

      expect { service.get('x', create: true) }.to raise_error(DaytonaApiClient::ApiError)
    end
  end

  describe '#list' do
    it 'returns an array of Volume models' do
      dtos = [build_volume_dto(name: 'v1'), build_volume_dto(name: 'v2')]
      allow(volumes_api).to receive(:list_volumes).and_return(dtos)

      volumes = service.list

      expect(volumes).to all(be_a(Daytona::Volume))
      expect(volumes.map(&:name)).to eq(%w[v1 v2])
    end
  end
end
