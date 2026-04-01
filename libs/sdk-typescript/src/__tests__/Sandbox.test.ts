import type { Configuration, Sandbox as SandboxDto } from '@daytonaio/api-client'
import { createApiResponse } from './helpers'

jest.mock(
  '@daytonaio/api-client',
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
  '@daytonaio/toolbox-api-client',
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
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  const { Sandbox } = require('../Sandbox') as typeof import('../Sandbox')
  const sandboxApi = {
    startSandbox: jest.fn(),
    stopSandbox: jest.fn(),
    deleteSandbox: jest.fn(),
    getSandbox: jest.fn(),
    replaceLabels: jest.fn(),
    setAutostopInterval: jest.fn(),
    setAutoArchiveInterval: jest.fn(),
    setAutoDeleteInterval: jest.fn(),
    getPortPreviewUrl: jest.fn(),
    archiveSandbox: jest.fn(),
    resizeSandbox: jest.fn(),
    updateLastActivity: jest.fn(),
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
    { getRunCommand: () => 'python run.py' },
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
