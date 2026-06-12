/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { RunnerAdapterV0 } from './runnerAdapter.v0'
import { TypedConfigService } from '../../config/typed-config.service'
import { RunnerApiError } from '../errors/runner-api-error'
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
    // Every poll must carry a short per-request timeout so a hung connection
    // cannot consume the whole poll budget via the instance-wide axios timeout.
    expect(sandboxApi.snapshotFromSandboxStatus).toHaveBeenCalledWith('sbx-1', { timeout: 30_000 })
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

  it('tolerates transient poll errors (network, per-request timeout, 5xx) and keeps polling', async () => {
    const { adapter, sandboxApi } = createAdapter()
    sandboxApi.snapshotFromSandbox.mockResolvedValue({ status: 202, data: 'Snapshot capture started' })
    sandboxApi.snapshotFromSandboxStatus
      .mockRejectedValueOnce(new Error('ECONNRESET'))
      .mockRejectedValueOnce(new RunnerApiError('timeout of 30000ms exceeded', undefined, 'ECONNABORTED'))
      .mockRejectedValueOnce(new RunnerApiError('upstream unavailable', 503))
      .mockResolvedValueOnce({
        status: 200,
        data: { state: 'COMPLETED', name: 'my-snap', snapshot: completedSnapshot },
      })

    const resultPromise = adapter.createSnapshotFromSandbox('sbx-1', 'my-snap', 'org-1', registry)
    for (let i = 0; i < 4; i++) {
      await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)
    }

    await expect(resultPromise).resolves.toEqual(
      expect.objectContaining({ ref: completedSnapshot.name, hash: completedSnapshot.hash }),
    )
    expect(sandboxApi.snapshotFromSandboxStatus).toHaveBeenCalledTimes(4)
  })

  it('fails fast when a status poll returns a permanent 4xx error', async () => {
    const { adapter, sandboxApi } = createAdapter()
    sandboxApi.snapshotFromSandbox.mockResolvedValue({ status: 202, data: 'Snapshot capture started' })
    sandboxApi.snapshotFromSandboxStatus.mockRejectedValue(new RunnerApiError('sandbox is not running', 400))

    const resultPromise = adapter.createSnapshotFromSandbox('sbx-1', 'my-snap', 'org-1', registry)
    const expectation = expect(resultPromise).rejects.toThrow(
      'snapshot capture status poll failed with HTTP 400: sandbox is not running',
    )
    await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)
    await expectation
    expect(sandboxApi.snapshotFromSandboxStatus).toHaveBeenCalledTimes(1)
  })

  it('treats a 404 from the status endpoint as permanent and hints at runner restart/downgrade', async () => {
    const { adapter, sandboxApi } = createAdapter()
    sandboxApi.snapshotFromSandbox.mockResolvedValue({ status: 202, data: 'Snapshot capture started' })
    sandboxApi.snapshotFromSandboxStatus.mockRejectedValue(new RunnerApiError('404 page not found', 404))

    const resultPromise = adapter.createSnapshotFromSandbox('sbx-1', 'my-snap', 'org-1', registry)
    const expectation = expect(resultPromise).rejects.toThrow(
      'runner no longer exposes the snapshot capture status (runner restarted or downgraded?): 404 page not found',
    )
    await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)
    await expectation
    expect(sandboxApi.snapshotFromSandboxStatus).toHaveBeenCalledTimes(1)
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

  it('clamps the per-poll timeout to the remaining capture budget near the deadline', async () => {
    const { adapter, sandboxApi } = createAdapter()
    sandboxApi.snapshotFromSandbox.mockResolvedValue({ status: 202, data: 'Snapshot capture started' })
    sandboxApi.snapshotFromSandboxStatus.mockResolvedValue({
      status: 200,
      data: { state: 'IN_PROGRESS', name: 'my-snap' },
    })

    const resultPromise = adapter.createSnapshotFromSandbox('sbx-1', 'my-snap', 'org-1', registry)
    const expectation = expect(resultPromise).rejects.toThrow('Timed out waiting for snapshot capture after 1 minutes')
    // 1 minute budget / 5s interval -> polls fire at t=5s..60s.
    for (let i = 0; i < 13; i++) {
      await jest.advanceTimersByTimeAsync(POLL_INTERVAL_MS)
    }
    await expectation

    const timeouts = sandboxApi.snapshotFromSandboxStatus.mock.calls.map(([, opts]) => opts.timeout)
    expect(timeouts).toHaveLength(12)
    // While the remaining budget exceeds the per-poll cap (t=5s..30s), the
    // full 30s timeout applies...
    expect(timeouts.slice(0, 6)).toEqual(Array(6).fill(30_000))
    // ...then each poll is clamped to the remaining budget (t=35s..55s)...
    expect(timeouts.slice(6, 11)).toEqual([25_000, 20_000, 15_000, 10_000, 5_000])
    // ...and the last poll (t=60s, budget exhausted mid-sleep) floors at 1s.
    expect(timeouts[11]).toBe(1_000)
  })
})
