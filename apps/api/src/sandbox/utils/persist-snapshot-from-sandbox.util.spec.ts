/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ConflictException } from '@nestjs/common'
import { Snapshot } from '../entities/snapshot.entity'
import { SnapshotState } from '../enums/snapshot-state.enum'
import { SnapshotRunnerState } from '../enums/snapshot-runner-state.enum'
import { SnapshotEvents } from '../constants/snapshot-events'
import { SnapshotCreatedEvent } from '../events/snapshot-created.event'
import { SandboxClass } from '../enums/sandbox-class.enum'
import { BuildInfo } from '../entities/build-info.entity'
import {
  activateSnapshotFromSandbox,
  createPendingSnapshotFromSandbox,
  failSnapshotFromSandbox,
  isPendingCaptureSnapshot,
  PersistSnapshotFromSandboxDeps,
} from './persist-snapshot-from-sandbox.util'

interface DepsStub {
  snapshotRepository: {
    create: jest.Mock
    insert: jest.Mock
    update: jest.Mock
    findOne: jest.Mock
  }
  snapshotRunnerRepository: {
    create: jest.Mock
    save: jest.Mock
  }
  eventEmitter: {
    emit: jest.Mock
  }
}

function createDeps(): { stub: DepsStub; deps: PersistSnapshotFromSandboxDeps } {
  const stub: DepsStub = {
    snapshotRepository: {
      create: jest.fn((data: Partial<Snapshot>) => data as Snapshot),
      insert: jest.fn(async (snapshot: Snapshot) => snapshot),
      update: jest.fn(async (_id: string, params: { updateData: Partial<Snapshot>; entity?: Snapshot }) => ({
        ...params.entity,
        ...params.updateData,
      })),
      findOne: jest.fn(async () => null),
    },
    snapshotRunnerRepository: {
      create: jest.fn((data: unknown) => data),
      save: jest.fn(async (runner: unknown) => runner),
    },
    eventEmitter: {
      emit: jest.fn(),
    },
  }
  return { stub, deps: stub as unknown as PersistSnapshotFromSandboxDeps }
}

const pendingParams = {
  organizationId: 'org-1',
  name: 'my-snap',
  regionId: 'region-1',
  sandboxClass: SandboxClass.CONTAINER,
  cpu: 2,
  gpu: 0,
  gpuType: null,
  mem: 4,
  disk: 10,
}

const activateParams = {
  ...pendingParams,
  ref: 'registry.example.com/daytona/daytona-abc:daytona',
  runnerId: 'runner-1',
  sizeGB: 4.2,
}

function captureRecord(overrides: Partial<Snapshot> = {}): Snapshot {
  return {
    id: 'snap-1',
    organizationId: 'org-1',
    name: 'my-snap',
    imageName: '',
    state: SnapshotState.PENDING,
    buildInfo: undefined,
    ...overrides,
  } as Snapshot
}

describe('isPendingCaptureSnapshot', () => {
  it('is true only for pending records with empty imageName and no buildInfo', () => {
    expect(isPendingCaptureSnapshot(captureRecord())).toBe(true)
  })

  it.each([
    ['pending with imageName', captureRecord({ imageName: 'ubuntu:22.04' })],
    ['pending with buildInfo', captureRecord({ buildInfo: { snapshotRef: 'ref' } as BuildInfo })],
    ['active capture-shaped record', captureRecord({ state: SnapshotState.ACTIVE })],
    ['errored capture-shaped record', captureRecord({ state: SnapshotState.ERROR })],
  ])('is false for %s', (_label, snapshot) => {
    expect(isPendingCaptureSnapshot(snapshot)).toBe(false)
  })
})

describe('createPendingSnapshotFromSandbox', () => {
  it('inserts a PENDING record with capture fields and no result fields', async () => {
    const { stub, deps } = createDeps()

    const inserted = await createPendingSnapshotFromSandbox(deps, pendingParams)

    expect(stub.snapshotRepository.insert).toHaveBeenCalledTimes(1)
    const entity = stub.snapshotRepository.insert.mock.calls[0][0]
    expect(entity).toMatchObject({
      organizationId: 'org-1',
      name: 'my-snap',
      state: SnapshotState.PENDING,
      sandboxClass: SandboxClass.CONTAINER,
      cpu: 2,
      gpu: 0,
      gpuType: null,
      mem: 4,
      disk: 10,
    })
    expect(entity.snapshotRegions).toEqual([{ snapshotId: entity.id, regionId: 'region-1' }])
    expect(entity).not.toHaveProperty('ref')
    expect(entity).not.toHaveProperty('size')
    expect(entity).not.toHaveProperty('lastUsedAt')
    expect(entity).not.toHaveProperty('initialRunnerId')
    expect(inserted).toBe(entity)
  })

  it('emits the CREATED event with the inserted snapshot', async () => {
    const { stub, deps } = createDeps()

    const inserted = await createPendingSnapshotFromSandbox(deps, pendingParams)

    expect(stub.eventEmitter.emit).toHaveBeenCalledTimes(1)
    const [event, payload] = stub.eventEmitter.emit.mock.calls[0]
    expect(event).toBe(SnapshotEvents.CREATED)
    expect(payload).toBeInstanceOf(SnapshotCreatedEvent)
    expect(payload.snapshot).toBe(inserted)
  })

  it('creates no SnapshotRunner', async () => {
    const { stub, deps } = createDeps()

    await createPendingSnapshotFromSandbox(deps, pendingParams)

    expect(stub.snapshotRunnerRepository.create).not.toHaveBeenCalled()
    expect(stub.snapshotRunnerRepository.save).not.toHaveBeenCalled()
  })

  it('maps a unique-violation insert error to ConflictException naming the snapshot', async () => {
    const { stub, deps } = createDeps()
    stub.snapshotRepository.insert.mockRejectedValue(Object.assign(new Error('duplicate'), { code: '23505' }))

    await expect(createPendingSnapshotFromSandbox(deps, pendingParams)).rejects.toThrow(ConflictException)
    await expect(createPendingSnapshotFromSandbox(deps, pendingParams)).rejects.toThrow('my-snap')
    expect(stub.eventEmitter.emit).not.toHaveBeenCalled()
  })

  it('rethrows other insert errors unchanged', async () => {
    const { stub, deps } = createDeps()
    const failure = new Error('connection lost')
    stub.snapshotRepository.insert.mockRejectedValue(failure)

    await expect(createPendingSnapshotFromSandbox(deps, pendingParams)).rejects.toBe(failure)
    expect(stub.eventEmitter.emit).not.toHaveBeenCalled()
  })
})

describe('activateSnapshotFromSandbox', () => {
  it('updates a pending capture record to ACTIVE and wires up the SnapshotRunner', async () => {
    const { stub, deps } = createDeps()
    const record = captureRecord()
    stub.snapshotRepository.findOne.mockResolvedValue(record)

    const updated = await activateSnapshotFromSandbox(deps, activateParams)

    expect(stub.snapshotRepository.update).toHaveBeenCalledTimes(1)
    const [id, params] = stub.snapshotRepository.update.mock.calls[0]
    expect(id).toBe('snap-1')
    expect(params.entity).toBe(record)
    expect(params.updateData).toMatchObject({
      state: SnapshotState.ACTIVE,
      ref: activateParams.ref,
      size: 4.2,
      initialRunnerId: 'runner-1',
      errorReason: null,
    })
    expect(params.updateData.lastUsedAt).toBeInstanceOf(Date)

    expect(stub.snapshotRunnerRepository.create).toHaveBeenCalledWith({
      snapshotRef: activateParams.ref,
      runnerId: 'runner-1',
      state: SnapshotRunnerState.READY,
    })
    expect(stub.snapshotRunnerRepository.save).toHaveBeenCalledTimes(1)
    expect(stub.snapshotRepository.insert).not.toHaveBeenCalled()
    expect(updated).toMatchObject({ state: SnapshotState.ACTIVE, ref: activateParams.ref })
  })

  it('omits size from the update when sizeGB is not finite', async () => {
    const { stub, deps } = createDeps()
    stub.snapshotRepository.findOne.mockResolvedValue(captureRecord())

    await activateSnapshotFromSandbox(deps, { ...activateParams, sizeGB: Number.NaN })

    const [, params] = stub.snapshotRepository.update.mock.calls[0]
    expect(params.updateData).not.toHaveProperty('size')
  })

  it('falls back to inserting an ACTIVE record when no record exists', async () => {
    const { stub, deps } = createDeps()

    const inserted = await activateSnapshotFromSandbox(deps, activateParams)

    expect(stub.snapshotRepository.update).not.toHaveBeenCalled()
    expect(stub.snapshotRepository.insert).toHaveBeenCalledTimes(1)
    const entity = stub.snapshotRepository.insert.mock.calls[0][0]
    expect(entity).toMatchObject({
      organizationId: 'org-1',
      name: 'my-snap',
      ref: activateParams.ref,
      state: SnapshotState.ACTIVE,
      size: 4.2,
      initialRunnerId: 'runner-1',
    })
    expect(entity.lastUsedAt).toBeInstanceOf(Date)
    expect(entity.snapshotRegions).toEqual([{ snapshotId: entity.id, regionId: 'region-1' }])

    const [event, payload] = stub.eventEmitter.emit.mock.calls[0]
    expect(event).toBe(SnapshotEvents.CREATED)
    expect(payload.snapshot).toBe(inserted)
    expect(stub.snapshotRunnerRepository.save).toHaveBeenCalledTimes(1)
  })

  it('returns null without writing when the record is no longer a pending capture', async () => {
    const { stub, deps } = createDeps()
    stub.snapshotRepository.findOne.mockResolvedValue(captureRecord({ state: SnapshotState.ERROR }))

    const result = await activateSnapshotFromSandbox(deps, activateParams)

    expect(result).toBeNull()
    expect(stub.snapshotRepository.update).not.toHaveBeenCalled()
    expect(stub.snapshotRepository.insert).not.toHaveBeenCalled()
    expect(stub.snapshotRunnerRepository.save).not.toHaveBeenCalled()
    expect(stub.eventEmitter.emit).not.toHaveBeenCalled()
  })

  it('returns null for a pending record that belongs to the pull flow (imageName set)', async () => {
    const { stub, deps } = createDeps()
    stub.snapshotRepository.findOne.mockResolvedValue(captureRecord({ imageName: 'ubuntu:22.04' }))

    const result = await activateSnapshotFromSandbox(deps, activateParams)

    expect(result).toBeNull()
    expect(stub.snapshotRepository.update).not.toHaveBeenCalled()
    expect(stub.snapshotRepository.insert).not.toHaveBeenCalled()
  })

  it('skips the SnapshotRunner when no runnerId is given', async () => {
    const { stub, deps } = createDeps()
    stub.snapshotRepository.findOne.mockResolvedValue(captureRecord())

    await activateSnapshotFromSandbox(deps, { ...activateParams, runnerId: null })

    expect(stub.snapshotRepository.update).toHaveBeenCalledTimes(1)
    expect(stub.snapshotRunnerRepository.create).not.toHaveBeenCalled()
    expect(stub.snapshotRunnerRepository.save).not.toHaveBeenCalled()
  })
})

describe('failSnapshotFromSandbox', () => {
  const failParams = { organizationId: 'org-1', name: 'my-snap', errorReason: 'runner exploded' }

  it('marks a pending capture record as ERROR with the reason', async () => {
    const { stub, deps } = createDeps()
    const record = captureRecord()
    stub.snapshotRepository.findOne.mockResolvedValue(record)

    await failSnapshotFromSandbox(deps, failParams)

    expect(stub.snapshotRepository.update).toHaveBeenCalledTimes(1)
    const [id, params] = stub.snapshotRepository.update.mock.calls[0]
    expect(id).toBe('snap-1')
    expect(params.entity).toBe(record)
    expect(params.updateData).toEqual({
      state: SnapshotState.ERROR,
      errorReason: 'runner exploded',
    })
  })

  it('does nothing when the record is missing', async () => {
    const { stub, deps } = createDeps()

    await failSnapshotFromSandbox(deps, failParams)

    expect(stub.snapshotRepository.update).not.toHaveBeenCalled()
  })

  it.each([
    ['a pull snapshot sharing the name', captureRecord({ imageName: 'ubuntu:22.04' })],
    ['a build snapshot sharing the name', captureRecord({ buildInfo: { snapshotRef: 'ref' } as BuildInfo })],
    ['an already-active record', captureRecord({ state: SnapshotState.ACTIVE })],
  ])('leaves %s untouched', async (_label, record) => {
    const { stub, deps } = createDeps()
    stub.snapshotRepository.findOne.mockResolvedValue(record)

    await failSnapshotFromSandbox(deps, failParams)

    expect(stub.snapshotRepository.update).not.toHaveBeenCalled()
  })
})
