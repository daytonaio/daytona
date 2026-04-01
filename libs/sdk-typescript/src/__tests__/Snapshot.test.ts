import type { Configuration } from '@daytonaio/api-client'
import { createApiResponse } from './helpers'
import { SnapshotService } from '../Snapshot'
import { Image } from '../Image'

jest.mock(
  '@daytonaio/api-client',
  () => ({
    SnapshotState: {
      ACTIVE: 'active',
      ERROR: 'error',
      BUILD_FAILED: 'build_failed',
      PENDING: 'pending',
    },
  }),
  { virtual: true },
)

describe('SnapshotService', () => {
  const cfg: Configuration = {
    basePath: 'http://api',
    baseOptions: { headers: { Authorization: 'Bearer token' } },
  } as unknown as Configuration

  const snapshotsApi = {
    getAllSnapshots: jest.fn(),
    getSnapshot: jest.fn(),
    removeSnapshot: jest.fn(),
    createSnapshot: jest.fn(),
    getSnapshotBuildLogsUrl: jest.fn(),
    activateSnapshot: jest.fn(),
  }
  const objectStorageApi = {
    getPushAccess: jest.fn(),
  }

  const service = new SnapshotService(cfg, snapshotsApi as unknown as never, objectStorageApi as unknown as never, 'eu')

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('lists/gets/deletes snapshots', async () => {
    snapshotsApi.getAllSnapshots.mockResolvedValue(
      createApiResponse({ items: [{ id: 's1', name: 'snap1' }], total: 1, page: 1, totalPages: 1 }),
    )
    snapshotsApi.getSnapshot.mockResolvedValue(createApiResponse({ id: 's1', name: 'snap1' }))
    snapshotsApi.removeSnapshot.mockResolvedValue(createApiResponse(undefined))

    await expect(service.list(1, 10)).resolves.toEqual({
      items: [{ id: 's1', name: 'snap1' }],
      total: 1,
      page: 1,
      totalPages: 1,
    })
    await expect(service.get('snap1')).resolves.toEqual({ id: 's1', name: 'snap1' })
    await service.delete({ id: 's1' } as never)
  })

  it('creates snapshot from image name with resources and region', async () => {
    snapshotsApi.createSnapshot.mockResolvedValue(createApiResponse({ id: 's2', name: 'snap2', state: 'active' }))

    const snapshot = await service.create({
      name: 'snap2',
      image: 'python:3.12',
      resources: { cpu: 4, memory: 8 },
      entrypoint: ['python', 'main.py'],
    })

    expect(snapshot).toEqual({ id: 's2', name: 'snap2', state: 'active' })
    expect(snapshotsApi.createSnapshot).toHaveBeenCalled()
  })

  it('creates snapshot from declarative image using processImageContext', async () => {
    const contextSpy = jest.spyOn(SnapshotService, 'processImageContext').mockResolvedValue(['hash1'])
    snapshotsApi.createSnapshot.mockResolvedValue(createApiResponse({ id: 's3', name: 'snap3', state: 'active' }))

    const image = Image.base('python:3.12').runCommands('echo hi')
    const snapshot = await service.create({ name: 'snap3', image })

    expect(snapshot.id).toBe('s3')
    expect(contextSpy).toHaveBeenCalled()
  })

  it('activates snapshots', async () => {
    snapshotsApi.activateSnapshot.mockResolvedValue(createApiResponse({ id: 's1', name: 'snap1', state: 'active' }))

    await expect(service.activate({ id: 's1' } as never)).resolves.toEqual({ id: 's1', name: 'snap1', state: 'active' })
  })
})
