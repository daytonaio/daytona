import { createApiResponse } from './helpers'
import { DaytonaNotFoundError } from '../errors/DaytonaError'
import { VolumeService } from '../Volume'

jest.mock('@daytonaio/api-client', () => ({}), { virtual: true })

describe('VolumeService', () => {
  const volumesApi = {
    listVolumes: jest.fn(),
    getVolumeByName: jest.fn(),
    createVolume: jest.fn(),
    deleteVolume: jest.fn(),
  }
  const service = new VolumeService(volumesApi as unknown as never)

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('lists and gets volumes', async () => {
    volumesApi.listVolumes.mockResolvedValue(createApiResponse([{ id: 'v1', name: 'vol1' }]))
    volumesApi.getVolumeByName.mockResolvedValue(createApiResponse({ id: 'v1', name: 'vol1' }))

    await expect(service.list()).resolves.toEqual([{ id: 'v1', name: 'vol1' }])
    await expect(service.get('vol1')).resolves.toEqual({ id: 'v1', name: 'vol1' })
  })

  it('creates volume on not found when create=true', async () => {
    volumesApi.getVolumeByName.mockRejectedValue(new DaytonaNotFoundError('missing', 404))
    volumesApi.createVolume.mockResolvedValue(createApiResponse({ id: 'v2', name: 'vol2' }))

    await expect(service.get('vol2', true)).resolves.toEqual({ id: 'v2', name: 'vol2' })
  })

  it('creates and deletes volume', async () => {
    volumesApi.createVolume.mockResolvedValue(createApiResponse({ id: 'v3', name: 'vol3' }))
    await expect(service.create('vol3')).resolves.toEqual({ id: 'v3', name: 'vol3' })

    await service.delete({ id: 'v3' } as never)
    expect(volumesApi.deleteVolume).toHaveBeenCalledWith('v3')
  })
})
