// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import type { Configuration } from '@daytona/api-client'
import { createApiResponse } from './helpers'
import { SnapshotService } from '../Snapshot'
import { Image } from '../Image'

const mockProcessStreamingResponse = jest.fn()
const mockDynamicImport = jest.fn()

jest.mock(
  '@daytona/api-client',
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

jest.mock('../utils/Stream', () => ({
  processStreamingResponse: (...args: unknown[]) => mockProcessStreamingResponse(...args),
}))

jest.mock('../utils/Import', () => ({
  dynamicImport: (...args: unknown[]) => mockDynamicImport(...args),
}))

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
    jest.restoreAllMocks()
    jest.clearAllMocks()
    mockDynamicImport.mockReset()
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

  it('passes timeout values in milliseconds to snapshot creation', async () => {
    snapshotsApi.createSnapshot.mockResolvedValue(createApiResponse({ id: 's2', name: 'snap2', state: 'active' }))

    await service.create({ name: 'snap2', image: 'python:3.12' }, { timeout: 12 })

    expect(snapshotsApi.createSnapshot).toHaveBeenCalledWith(expect.any(Object), undefined, { timeout: 12000 })
  })

  it('creates snapshot from declarative image using processImageContext', async () => {
    const contextSpy = jest.spyOn(SnapshotService, 'processImageContext').mockResolvedValue(['hash1'])
    snapshotsApi.createSnapshot.mockResolvedValue(createApiResponse({ id: 's3', name: 'snap3', state: 'active' }))

    const image = Image.base('python:3.12').runCommands('echo hi')
    const snapshot = await service.create({ name: 'snap3', image })

    expect(snapshot.id).toBe('s3')
    expect(contextSpy).toHaveBeenCalled()
  })

  it('throws when the api returns no created snapshot', async () => {
    snapshotsApi.createSnapshot.mockResolvedValue(createApiResponse(undefined))

    await expect(service.create({ name: 'snap-missing', image: 'python:3.12' })).rejects.toThrow(
      "Failed to create snapshot. Didn't receive a snapshot from the server API.",
    )
  })

  it('throws when terminal snapshot states indicate failure', async () => {
    snapshotsApi.createSnapshot.mockResolvedValue(
      createApiResponse({ id: 's4', name: 'snap4', state: 'error', errorReason: 'build failed' }),
    )

    await expect(service.create({ name: 'snap4', image: 'python:3.12' })).rejects.toThrow(
      'Failed to create snapshot. Name: snap4 Reason: build failed',
    )
  })

  it('returns empty context hashes when an image has no context files', async () => {
    await expect(
      SnapshotService.processImageContext(objectStorageApi as never, Image.base('python:3.12')),
    ).resolves.toEqual([])
  })

  it('uploads image contexts through object storage push credentials', async () => {
    const upload = jest.fn().mockResolvedValue('ctx-hash')
    const ObjectStorage = jest.fn().mockImplementation(() => ({ upload }))
    objectStorageApi.getPushAccess.mockResolvedValue(
      createApiResponse({
        storageUrl: 'https://s3.us-east-1.amazonaws.com',
        accessKey: 'key',
        secret: 'secret',
        sessionToken: 'session',
        bucket: 'bucket',
        organizationId: 'org-1',
      }),
    )
    mockDynamicImport.mockResolvedValue({ ObjectStorage })

    const image = Image.base('python:3.12')
    ;(image as unknown as { _contextList: Array<{ sourcePath: string; archivePath: string }> })._contextList = [
      { sourcePath: '/tmp/context', archivePath: '.' },
    ]

    await expect(SnapshotService.processImageContext(objectStorageApi as never, image)).resolves.toEqual(['ctx-hash'])
    expect(objectStorageApi.getPushAccess).toHaveBeenCalledTimes(1)
    expect(upload).toHaveBeenCalledWith('/tmp/context', 'org-1', '.')
  })

  it('streams build logs when onLogs is provided for build snapshots', async () => {
    const fetchSpy = jest.spyOn(global, 'fetch' as never).mockResolvedValue({ ok: true } as never)
    snapshotsApi.createSnapshot.mockResolvedValue(createApiResponse({ id: 's5', name: 'snap5', state: 'building' }))
    snapshotsApi.getSnapshotBuildLogsUrl.mockResolvedValue(createApiResponse({ url: 'https://logs.daytona/snap5' }))
    snapshotsApi.getSnapshot.mockResolvedValue(createApiResponse({ id: 's5', name: 'snap5', state: 'active' }))
    mockProcessStreamingResponse.mockImplementation(async (_fetchLogs, onChunk: (chunk: string) => void) => {
      onChunk('log line')
    })

    const onLogs = jest.fn()
    await service.create({ name: 'snap5', image: Image.base('python:3.12').runCommands('echo hi') }, { onLogs })

    expect(snapshotsApi.getSnapshotBuildLogsUrl).toHaveBeenCalledWith('s5')
    expect(mockProcessStreamingResponse).toHaveBeenCalled()
    expect(onLogs).toHaveBeenCalledWith(expect.stringContaining('Creating snapshot snap5'))
    expect(onLogs).toHaveBeenCalledWith('log line')

    fetchSpy.mockRestore()
  })

  it('activates snapshots', async () => {
    snapshotsApi.activateSnapshot.mockResolvedValue(createApiResponse({ id: 's1', name: 'snap1', state: 'active' }))

    await expect(service.activate({ id: 's1' } as never)).resolves.toEqual({ id: 's1', name: 'snap1', state: 'active' })
  })
})
