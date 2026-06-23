/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { BadRequestException, NotFoundException } from '@nestjs/common'
import { SnapshotEvents } from '../constants/snapshot-events'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { SnapshotService } from './snapshot.service'

describe('SnapshotService.retryFailedSnapshot', () => {
  const snapshotRepository = {
    findOne: jest.fn(),
    update: jest.fn(),
  }
  const eventEmitter = {
    emit: jest.fn(),
  }
  const service = new SnapshotService(
    {} as any,
    snapshotRepository as any,
    {} as any,
    {} as any,
    {} as any,
    {} as any,
    {} as any,
    {} as any,
    {} as any,
    {} as any,
    {} as any,
    {} as any,
    eventEmitter as any,
    {} as any,
  )

  beforeEach(() => {
    jest.resetAllMocks()
  })

  it('resets a failed pull snapshot and emits a created event', async () => {
    const snapshot = {
      id: 'snapshot-1',
      state: SnapshotState.ERROR,
      errorReason: 'registry timeout',
      initialRunnerId: 'runner-1',
      ref: 'registry/snapshot@sha256:digest',
      size: 1,
    }
    const retriedSnapshot = {
      ...snapshot,
      state: SnapshotState.PENDING,
      errorReason: null,
      initialRunnerId: null,
      ref: null,
      size: null,
    }
    snapshotRepository.findOne.mockResolvedValue(snapshot)
    snapshotRepository.update.mockResolvedValue(retriedSnapshot)

    await expect(service.retryFailedSnapshot(snapshot.id)).resolves.toBe(retriedSnapshot)
    expect(snapshotRepository.update).toHaveBeenCalledWith(snapshot.id, {
      updateData: {
        state: SnapshotState.PENDING,
        errorReason: null,
        initialRunnerId: null,
        ref: null,
        size: null,
      },
      entity: snapshot,
    })
    expect(eventEmitter.emit).toHaveBeenCalledWith(
      SnapshotEvents.CREATED,
      expect.objectContaining({ snapshot: retriedSnapshot }),
    )
  })

  it('preserves build metadata when retrying a failed build snapshot', async () => {
    const snapshot = {
      id: 'snapshot-1',
      state: SnapshotState.BUILD_FAILED,
      errorReason: 'build timeout',
      initialRunnerId: 'runner-1',
      ref: 'registry/build@sha256:digest',
      size: 1,
      buildInfo: { id: 'build-1' },
    }
    snapshotRepository.findOne.mockResolvedValue(snapshot)
    snapshotRepository.update.mockResolvedValue({ ...snapshot, state: SnapshotState.PENDING })

    await service.retryFailedSnapshot(snapshot.id)

    expect(snapshotRepository.update).toHaveBeenCalledWith(snapshot.id, {
      updateData: {
        state: SnapshotState.PENDING,
        errorReason: null,
        initialRunnerId: null,
      },
      entity: snapshot,
    })
  })

  it('rejects snapshots that do not exist or are not failed', async () => {
    snapshotRepository.findOne.mockResolvedValueOnce(null)
    await expect(service.retryFailedSnapshot('missing')).rejects.toThrow(NotFoundException)

    snapshotRepository.findOne.mockResolvedValueOnce({
      id: 'snapshot-1',
      state: SnapshotState.ACTIVE,
    })
    await expect(service.retryFailedSnapshot('snapshot-1')).rejects.toThrow(BadRequestException)
  })
})
