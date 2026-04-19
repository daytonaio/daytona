/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { SnapshotState } from '@daytona/api-client'
import { SnapshotService } from '../Snapshot'

describe('SnapshotService', () => {
  function snapshotPayload(overrides: Partial<ReturnType<typeof snapshotPayloadBase>> = {}) {
    return {
      ...snapshotPayloadBase(),
      ...overrides,
    }
  }

  function snapshotPayloadBase() {
    return {
      id: 'snap_123',
      organizationId: 'org_123',
      general: false,
      name: 'demo',
      imageName: 'img',
      state: SnapshotState.ACTIVE,
      size: 1,
      entrypoint: [],
      cpu: 1,
      gpu: 0,
      mem: 1,
      disk: 1,
      errorReason: '',
      createdAt: '2025-09-16T17:56:53.000Z',
      updatedAt: '2025-09-16T17:57:53.000Z',
      lastUsedAt: '2025-09-16T17:58:53.000Z',
    }
  }

  function newSnapshotService(overrides: Partial<ReturnType<typeof snapshotPayloadBase>> = {}) {
    const snapshotsApi = {
      getAllSnapshots: jest.fn().mockResolvedValue({
        data: {
          items: [snapshotPayload(overrides)],
          total: 1,
          page: 1,
          totalPages: 1,
        },
      }),
      getSnapshot: jest.fn().mockResolvedValue({ data: snapshotPayload(overrides) }),
      createSnapshot: jest.fn().mockResolvedValue({ data: snapshotPayload(overrides) }),
      activateSnapshot: jest.fn().mockResolvedValue({ data: snapshotPayload(overrides) }),
    }

    return {
      service: new SnapshotService({} as never, snapshotsApi as never, {} as never),
      snapshotsApi,
    }
  }

  it('deserializes snapshot dates returned by list', async () => {
    const { service } = newSnapshotService()

    const result = await service.list()

    expect(result.items[0].createdAt).toBeInstanceOf(Date)
    expect(result.items[0].updatedAt).toBeInstanceOf(Date)
    expect(result.items[0].lastUsedAt).toBeInstanceOf(Date)
  })

  it('deserializes snapshot dates returned by get', async () => {
    const { service } = newSnapshotService()

    const snapshot = await service.get('demo')

    expect(snapshot.createdAt).toBeInstanceOf(Date)
    expect(snapshot.updatedAt).toBeInstanceOf(Date)
    expect(snapshot.lastUsedAt).toBeInstanceOf(Date)
  })

  it('deserializes snapshot dates returned by create', async () => {
    const { service } = newSnapshotService()

    const snapshot = await service.create({ name: 'demo', image: 'img' })

    expect(snapshot.createdAt).toBeInstanceOf(Date)
    expect(snapshot.updatedAt).toBeInstanceOf(Date)
    expect(snapshot.lastUsedAt).toBeInstanceOf(Date)
  })

  it('deserializes snapshot dates returned by activate', async () => {
    const { service } = newSnapshotService()

    const snapshot = await service.activate({ id: 'snap_123' } as never)

    expect(snapshot.createdAt).toBeInstanceOf(Date)
    expect(snapshot.updatedAt).toBeInstanceOf(Date)
    expect(snapshot.lastUsedAt).toBeInstanceOf(Date)
  })

  it('preserves null lastUsedAt values', async () => {
    const { service } = newSnapshotService({ lastUsedAt: null })

    const snapshot = await service.get('demo')

    expect(snapshot.lastUsedAt).toBeNull()
  })
})
