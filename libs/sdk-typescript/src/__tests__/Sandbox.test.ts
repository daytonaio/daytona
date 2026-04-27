// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import type { Configuration, Sandbox as SandboxDto } from '@daytona/api-client'
import { createApiResponse } from './helpers'
import { DaytonaNotFoundError } from '../errors/DaytonaError'

jest.mock(
  '@daytona/api-client',
  () => ({
    SandboxState: {
      RESIZING: 'resizing',
      ERROR: 'error',
      BUILD_FAILED: 'build_failed',
      DESTROYED: 'destroyed',
    },
  }),
  { virtual: true },
)

jest.mock(
  '@daytona/toolbox-api-client',
  () => ({
    FileSystemApi: jest.fn(() => ({})),
    GitApi: jest.fn(() => ({})),
    ProcessApi: jest.fn(() => ({})),
    LspApi: jest.fn(() => ({})),
    InfoApi: jest.fn(() => ({ getUserHomeDir: jest.fn(), getWorkDir: jest.fn() })),
    ComputerUseApi: jest.fn(() => ({})),
    InterpreterApi: jest.fn(() => ({})),
  }),
  { virtual: true },
)

const baseDto: SandboxDto = {
  id: 'sb-1',
  name: 'sandbox-one',
  organizationId: 'org-1',
  user: 'daytona',
  env: {},
  labels: {},
  public: false,
  target: 'eu',
  cpu: 2,
  gpu: 0,
  memory: 4,
  disk: 10,
  state: 'stopped',
  networkBlockAll: false,
  toolboxProxyUrl: 'http://proxy',
}

const makeSandbox = (
  overrides: Partial<SandboxDto> = {},
): { sandbox: import('../Sandbox').Sandbox; sandboxApi: Record<string, jest.Mock> } => {
  const { Sandbox } = require('../Sandbox') as typeof import('../Sandbox')
  const sandboxApi = {
    startSandbox: jest.fn(),
    recoverSandbox: jest.fn(),
    stopSandbox: jest.fn(),
    deleteSandbox: jest.fn(),
    getSandbox: jest.fn(),
    replaceLabels: jest.fn(),
    setAutostopInterval: jest.fn(),
    setAutoArchiveInterval: jest.fn(),
    setAutoDeleteInterval: jest.fn(),
    getPortPreviewUrl: jest.fn(),
    getSignedPortPreviewUrl: jest.fn(),
    expireSignedPortPreviewUrl: jest.fn(),
    archiveSandbox: jest.fn(),
    resizeSandbox: jest.fn(),
    updateLastActivity: jest.fn(),
    createSandboxSnapshot: jest.fn(),
    createSshAccess: jest.fn(),
    revokeSshAccess: jest.fn(),
    validateSshAccess: jest.fn(),
  }

  const cfg: Configuration = {
    basePath: 'http://proxy',
    baseOptions: { headers: {} },
  } as unknown as Configuration

  const axiosInstance = {
    defaults: { baseURL: '' },
  }

  const sandbox = new Sandbox(
    { ...baseDto, ...overrides },
    cfg,
    axiosInstance as unknown as never,
    sandboxApi as unknown as never,
  )

  return { sandbox, sandboxApi }
}

describe('Sandbox', () => {
  it('maps dto fields in constructor', () => {
    const { sandbox } = makeSandbox({ labels: { team: 'sdk' }, autoStopInterval: 10 })
    expect(sandbox.id).toBe('sb-1')
    expect(sandbox.labels).toEqual({ team: 'sdk' })
    expect(sandbox.autoStopInterval).toBe(10)
  })

  it('delegates start/stop/delete to sandbox api', async () => {
    const { sandbox, sandboxApi } = makeSandbox({ state: 'stopped' })

    sandboxApi.startSandbox.mockResolvedValue(createApiResponse({ ...baseDto, state: 'started' }))
    sandboxApi.getSandbox.mockResolvedValue(createApiResponse({ ...baseDto, state: 'started' }))
    await sandbox.start(1)

    sandboxApi.stopSandbox.mockResolvedValue(createApiResponse(undefined))
    sandboxApi.getSandbox.mockResolvedValue(createApiResponse({ ...baseDto, state: 'stopped' }))
    await sandbox.stop(1)

    sandboxApi.deleteSandbox.mockResolvedValue(createApiResponse(undefined))
    await sandbox.delete(1)

    expect(sandboxApi.startSandbox).toHaveBeenCalledWith('sb-1', undefined, { timeout: 1000 })
    expect(sandboxApi.stopSandbox).toHaveBeenCalledWith('sb-1', undefined, false, { timeout: 1000 })
    expect(sandboxApi.deleteSandbox).toHaveBeenCalledWith('sb-1', undefined, { timeout: 1000 })
  })

  it('recovers sandboxes and waits for them to start', async () => {
    const { sandbox, sandboxApi } = makeSandbox({ state: 'error' })

    sandboxApi.recoverSandbox.mockResolvedValue(createApiResponse({ ...baseDto, state: 'starting' }))
    sandboxApi.getSandbox.mockResolvedValue(createApiResponse({ ...baseDto, state: 'started' }))

    await sandbox.recover(1)

    expect(sandboxApi.recoverSandbox).toHaveBeenCalledWith('sb-1', undefined, { timeout: 1000 })
    expect(sandbox.state).toBe('started')
  })

  test.each([['start'], ['recover'], ['stop'], ['waitUntilStarted'], ['waitUntilStopped'], ['resize']])(
    'validates negative timeout for %s',
    async (method) => {
      const { sandbox } = makeSandbox({ state: 'stopped' })
      const runtime = sandbox as unknown as Record<string, (...args: never[]) => Promise<void>>

      if (method === 'resize') {
        await expect(
          (sandbox as never as { resize: (resources: { cpu: number }, timeout: number) => Promise<void> }).resize(
            { cpu: 2 },
            -1,
          ),
        ).rejects.toThrow('Timeout must be a non-negative number')
        return
      }

      await expect(runtime[method](-1 as never)).rejects.toThrow('Timeout must be a non-negative number')
    },
  )

  test.each([
    ['setAutostopInterval', -1, 'autoStopInterval must be a non-negative integer'],
    ['setAutoArchiveInterval', -1, 'autoArchiveInterval must be a non-negative integer'],
  ])('validates %s', async (method, arg, message) => {
    const { sandbox } = makeSandbox()
    const call = (sandbox as unknown as Record<string, (n: number) => Promise<void>>)[method]
    await expect(call(arg)).rejects.toThrow(message)
  })

  it('updates labels and lifecycle intervals', async () => {
    const { sandbox, sandboxApi } = makeSandbox()
    sandboxApi.replaceLabels.mockResolvedValue(createApiResponse({ labels: { env: 'dev' } }))

    await expect(sandbox.setLabels({ env: 'dev' })).resolves.toEqual({ env: 'dev' })
    await sandbox.setAutostopInterval(20)
    await sandbox.setAutoArchiveInterval(30)
    await sandbox.setAutoDeleteInterval(0)

    expect(sandbox.labels).toEqual({ env: 'dev' })
    expect(sandbox.autoStopInterval).toBe(20)
    expect(sandbox.autoArchiveInterval).toBe(30)
    expect(sandbox.autoDeleteInterval).toBe(0)
  })

  it('handles preview link, archive, resize, refresh calls', async () => {
    const { sandbox, sandboxApi } = makeSandbox({ state: 'resizing' })

    sandboxApi.getPortPreviewUrl.mockResolvedValue(createApiResponse({ url: 'http://preview', token: 't' }))
    await expect(sandbox.getPreviewLink(3000)).resolves.toEqual({ url: 'http://preview', token: 't' })

    sandboxApi.archiveSandbox.mockResolvedValue(createApiResponse(undefined))
    sandboxApi.getSandbox.mockResolvedValue(createApiResponse({ ...baseDto, state: 'stopped' }))
    await sandbox.archive()

    sandboxApi.resizeSandbox.mockResolvedValue(createApiResponse({ ...baseDto, state: 'started' }))
    await sandbox.resize({ cpu: 4 }, 1)

    await sandbox.refreshData()
    await sandbox.refreshActivity()

    expect(sandboxApi.archiveSandbox).toHaveBeenCalledWith('sb-1')
    expect(sandboxApi.resizeSandbox).toHaveBeenCalled()
    expect(sandboxApi.updateLastActivity).toHaveBeenCalledWith('sb-1')
  })

  it('gets and expires signed preview urls', async () => {
    const { sandbox, sandboxApi } = makeSandbox()
    sandboxApi.getSignedPortPreviewUrl.mockResolvedValue(
      createApiResponse({ url: 'http://signed', token: 'signed-token' }),
    )
    sandboxApi.expireSignedPortPreviewUrl.mockResolvedValue(createApiResponse(undefined))

    await expect(sandbox.getSignedPreviewUrl(8080, 90)).resolves.toEqual({
      url: 'http://signed',
      token: 'signed-token',
    })
    await sandbox.expireSignedPreviewUrl(8080, 'signed-token')

    expect(sandboxApi.getSignedPortPreviewUrl).toHaveBeenCalledWith('sb-1', 8080, undefined, 90)
    expect(sandboxApi.expireSignedPortPreviewUrl).toHaveBeenCalledWith('sb-1', 8080, 'signed-token')
  })

  it('creates, revokes and validates ssh access tokens', async () => {
    const { sandbox, sandboxApi } = makeSandbox()
    sandboxApi.createSshAccess.mockResolvedValue(createApiResponse({ token: 'ssh-token' }))
    sandboxApi.revokeSshAccess.mockResolvedValue(createApiResponse(undefined))
    sandboxApi.validateSshAccess.mockResolvedValue(createApiResponse({ valid: true }))

    await expect(sandbox.createSshAccess(30)).resolves.toEqual({ token: 'ssh-token' })
    await sandbox.revokeSshAccess('ssh-token')
    await expect(sandbox.validateSshAccess('ssh-token')).resolves.toEqual({ valid: true })

    expect(sandboxApi.createSshAccess).toHaveBeenCalledWith('sb-1', undefined, 30)
    expect(sandboxApi.revokeSshAccess).toHaveBeenCalledWith('sb-1', undefined, 'ssh-token')
    expect(sandboxApi.validateSshAccess).toHaveBeenCalledWith('ssh-token')
  })

  it('aliases getUserRootDir to getUserHomeDir', async () => {
    const { sandbox } = makeSandbox()
    const infoApi = (sandbox as unknown as { infoApi: { getUserHomeDir: jest.Mock } }).infoApi
    const runtime = sandbox as unknown as { getUserRootDir: () => Promise<string | undefined> }
    infoApi.getUserHomeDir.mockResolvedValue(createApiResponse({ dir: '/home/daytona' }))

    await expect(runtime.getUserRootDir()).resolves.toBe('/home/daytona')
    expect(infoApi.getUserHomeDir).toHaveBeenCalledTimes(1)
  })

  it('creates sandbox snapshots and waits for completion', async () => {
    const { sandbox, sandboxApi } = makeSandbox({ state: 'snapshotting' })
    sandboxApi.createSandboxSnapshot.mockResolvedValue(createApiResponse(undefined))
    sandboxApi.getSandbox.mockResolvedValue(createApiResponse({ ...baseDto, state: 'started' }))

    await sandbox._experimental_createSnapshot('snap-1', 1)

    expect(sandboxApi.createSandboxSnapshot).toHaveBeenCalledWith('sb-1', { name: 'snap-1' }, undefined, {
      timeout: 1000,
    })
  })

  it('waitUntilStarted throws when sandbox enters an error state', async () => {
    const { sandbox } = makeSandbox({ state: 'starting', errorReason: 'boot failed' })

    jest.spyOn(sandbox, 'refreshData').mockImplementation(async () => {
      sandbox.state = 'error'
    })

    await expect(sandbox.waitUntilStarted(1)).rejects.toThrow(
      'Sandbox sb-1 failed to start with status: error, error reason: boot failed',
    )
  })

  it('waitUntilStopped treats deleted sandboxes as destroyed', async () => {
    const { sandbox } = makeSandbox({ state: 'stopping' })

    jest.spyOn(sandbox, 'refreshData').mockRejectedValue(new DaytonaNotFoundError('missing'))

    await expect(sandbox.waitUntilStopped(1)).resolves.toBeUndefined()
    expect(sandbox.state).toBe('destroyed')
  })

  it('waitForResizeComplete throws when resize fails', async () => {
    const { sandbox } = makeSandbox({ state: 'resizing', errorReason: 'no capacity' })

    jest.spyOn(sandbox, 'refreshData').mockImplementation(async () => {
      sandbox.state = 'error'
    })

    await expect(sandbox.waitForResizeComplete(1)).rejects.toThrow(
      'Sandbox sb-1 resize failed with state: error, error reason: no capacity',
    )
  })

  it('updates interval properties after successful api calls', async () => {
    const { sandbox, sandboxApi } = makeSandbox({ autoStopInterval: 1, autoArchiveInterval: 2, autoDeleteInterval: 3 })
    sandboxApi.setAutostopInterval.mockResolvedValue(createApiResponse(undefined))
    sandboxApi.setAutoArchiveInterval.mockResolvedValue(createApiResponse(undefined))
    sandboxApi.setAutoDeleteInterval.mockResolvedValue(createApiResponse(undefined))

    await sandbox.setAutostopInterval(15)
    await sandbox.setAutoArchiveInterval(25)
    await sandbox.setAutoDeleteInterval(-1)

    expect(sandbox.autoStopInterval).toBe(15)
    expect(sandbox.autoArchiveInterval).toBe(25)
    expect(sandbox.autoDeleteInterval).toBe(-1)
  })

  it('waitUntilStarted and waitUntilStopped complete when state changes', async () => {
    const { sandbox } = makeSandbox({ state: 'starting' })

    const refreshSpy = jest.spyOn(sandbox, 'refreshData').mockImplementation(async () => {
      sandbox.state = 'started'
    })
    await sandbox.waitUntilStarted(1)
    expect(refreshSpy).toHaveBeenCalled()

    sandbox.state = 'stopping'
    refreshSpy.mockImplementation(async () => {
      sandbox.state = 'stopped'
    })
    await sandbox.waitUntilStopped(1)
  })

  it('exposes user/work directories and creates LSP server', async () => {
    const { sandbox } = makeSandbox()
    const infoApi = (sandbox as unknown as { infoApi: { getUserHomeDir: jest.Mock; getWorkDir: jest.Mock } }).infoApi
    infoApi.getUserHomeDir.mockResolvedValue(createApiResponse({ dir: '/home/daytona' }))
    infoApi.getWorkDir.mockResolvedValue(createApiResponse({ dir: '/workspace' }))

    await expect(sandbox.getUserHomeDir()).resolves.toBe('/home/daytona')
    await expect(sandbox.getWorkDir()).resolves.toBe('/workspace')

    const lsp = await sandbox.createLspServer('typescript', '/workspace/project')
    expect(lsp).toBeTruthy()
  })
})
