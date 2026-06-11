/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RunnerAdapterV0 } from './runnerAdapter.v0'
import { TypedConfigService } from '../../config/typed-config.service'
import { DockerRegistry } from '../../docker-registry/entities/docker-registry.entity'

const POLL_INTERVAL_MS = 5_000

interface SandboxApiStub {
  snapshotFromSandbox: jest.Mock
  snapshotFromSandboxStatus: jest.Mock
}

function createAdapter(): { adapter: RunnerAdapterV0; sandboxApi: SandboxApiStub } {
  const configService = {
    // 1 minute poll budget -> 12 polls of 5s
    getOrThrow: jest.fn(() => 1),
  } as unknown as TypedConfigService // stub only needs getOrThrow('sandboxSnapshottingTimeoutMin')

  const adapter = new RunnerAdapterV0(configService)

  const sandboxApi: SandboxApiStub = {
    snapshotFromSandbox: jest.fn(),
    snapshotFromSandboxStatus: jest.fn(),
  }
  // Bypass init(): inject the API client stub directly.
  ;(adapter as unknown as { sandboxApiClient: SandboxApiStub }).sandboxApiClient = sandboxApi

  return { adapter, sandboxApi }
}

const registry = {
  project: 'proj',
  url: 'https://registry.example.com',
  username: 'user',
  password: 'pass',
} as DockerRegistry

const completedSnapshot = {
  name: 'registry.example.com/proj/daytona-abc123:daytona',
  hash: 'abc123',
  sizeGB: 2.5,
  entrypoint: ['/entry'],
  cmd: ['serve'],
}

describe('RunnerAdapterV0.createSnapshotFromSandbox', () => {
  beforeEach(() => {
    jest.useFakeTimers()
  })

  afterEach(() => {
    jest.useRealTimers()
  })

  it('returns the sync result without polling when the runner answers 200 with name+hash', async () => {
    const { adapter, sandboxApi } = createAdapter()
    sandboxApi.snapshotFromSandbox.mockResolvedValue({ status: 200, data: completedSnapshot })

    const result = await adapter.createSnapshotFromSandbox('sbx-1', 'my-snap', 'org-1', registry)

    expect(result).toEqual({
      ref: completedSnapshot.name,
      hash: completedSnapshot.hash,
      sizeGB: completedSnapshot.sizeGB,
      entrypoint: completedSnapshot.entrypoint,
      cmd: completedSnapshot.cmd,
    })
    expect(sandboxApi.snapshotFromSandbox).toHaveBeenCalledWith(
      'sbx-1',
      expect.objectContaining({ name: 'my-snap', async: true }),
    )
    expect(sandboxApi.snapshotFromSandboxStatus).not.toHaveBeenCalled()
  })

  it('polls until COMPLETED and maps ref/hash/sizeGB/entrypoint/cmd', async () => {
    const { adapter, sandboxApi } = createAdapter()
    sandboxApi.snapshotFromSandbox.mockResolvedValue({ status: 202, data: 'Snapshot capture started' })
    sandboxApi.snapshotFromSandboxStatus
      .mockResolvedValueOnce({ status: 200, data: { state: 'IN_PROGRESS', name: 'my-snap' } })
      .mockResolvedValueOnce({
        status: 200,
        data: { state: 'COMPLETED', name: 'my-snap', snapshot: completedSnapshot },
      })

    const resultPromise = adapter.createSnapshotFromSandbox('sbx-1', 'my-snap', 'org-1', registry)
    await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)
    await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)

    await expect(resultPromise).resolves.toEqual({
      ref: completedSnapshot.name,
      hash: completedSnapshot.hash,
      sizeGB: completedSnapshot.sizeGB,
      entrypoint: completedSnapshot.entrypoint,
      cmd: completedSnapshot.cmd,
    })
    expect(sandboxApi.snapshotFromSandboxStatus).toHaveBeenCalledTimes(2)
  })

  it('throws the runner-reported error when status is FAILED', async () => {
    const { adapter, sandboxApi } = createAdapter()
    sandboxApi.snapshotFromSandbox.mockResolvedValue({ status: 202, data: 'Snapshot capture started' })
    sandboxApi.snapshotFromSandboxStatus.mockResolvedValue({
      status: 200,
      data: { state: 'FAILED', name: 'my-snap', error: 'commit exploded on runner' },
    })

    const resultPromise = adapter.createSnapshotFromSandbox('sbx-1', 'my-snap', 'org-1', registry)
    const expectation = expect(resultPromise).rejects.toThrow('commit exploded on runner')
    await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)
    await expectation
  })

  it('fails fast when status is NONE after initiation', async () => {
    const { adapter, sandboxApi } = createAdapter()
    sandboxApi.snapshotFromSandbox.mockResolvedValue({ status: 202, data: 'Snapshot capture started' })
    sandboxApi.snapshotFromSandboxStatus.mockResolvedValue({ status: 200, data: { state: 'NONE' } })

    const resultPromise = adapter.createSnapshotFromSandbox('sbx-1', 'my-snap', 'org-1', registry)
    const expectation = expect(resultPromise).rejects.toThrow(
      'runner is no longer tracking the snapshot capture (runner restarted?)',
    )
    await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)
    await expectation
  })

  it('throws when the completed capture name does not match the requested snapshot', async () => {
    const { adapter, sandboxApi } = createAdapter()
    sandboxApi.snapshotFromSandbox.mockResolvedValue({ status: 202, data: 'Snapshot capture started' })
    sandboxApi.snapshotFromSandboxStatus.mockResolvedValue({
      status: 200,
      data: { state: 'COMPLETED', name: 'other-snap', snapshot: completedSnapshot },
    })

    const resultPromise = adapter.createSnapshotFromSandbox('sbx-1', 'my-snap', 'org-1', registry)
    const expectation = expect(resultPromise).rejects.toThrow(
      'snapshot capture state on runner does not match this capture (superseded or stale)',
    )
    await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)
    await expectation
  })

  it('tolerates transient poll errors and keeps polling', async () => {
    const { adapter, sandboxApi } = createAdapter()
    sandboxApi.snapshotFromSandbox.mockResolvedValue({ status: 202, data: 'Snapshot capture started' })
    sandboxApi.snapshotFromSandboxStatus
      .mockRejectedValueOnce(new Error('ECONNRESET'))
      .mockResolvedValueOnce({
        status: 200,
        data: { state: 'COMPLETED', name: 'my-snap', snapshot: completedSnapshot },
      })

    const resultPromise = adapter.createSnapshotFromSandbox('sbx-1', 'my-snap', 'org-1', registry)
    await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)
    await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)

    await expect(resultPromise).resolves.toEqual(
      expect.objectContaining({ ref: completedSnapshot.name, hash: completedSnapshot.hash }),
    )
    expect(sandboxApi.snapshotFromSandboxStatus).toHaveBeenCalledTimes(2)
  })

  it('times out after sandboxSnapshottingTimeoutMin and throws', async () => {
    const { adapter, sandboxApi } = createAdapter()
    sandboxApi.snapshotFromSandbox.mockResolvedValue({ status: 202, data: 'Snapshot capture started' })
    sandboxApi.snapshotFromSandboxStatus.mockResolvedValue({
      status: 200,
      data: { state: 'IN_PROGRESS', name: 'my-snap' },
    })

    const resultPromise = adapter.createSnapshotFromSandbox('sbx-1', 'my-snap', 'org-1', registry)
    const expectation = expect(resultPromise).rejects.toThrow('Timed out waiting for snapshot capture after 1 minutes')
    // 1 minute budget / 5s interval = 12 polls; advance past the budget.
    for (let i = 0; i < 13; i++) {
      await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)
    }
    await expectation
  })
})
