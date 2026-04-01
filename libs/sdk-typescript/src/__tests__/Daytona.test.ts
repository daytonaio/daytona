import { createApiResponse } from './helpers'

const mockAxiosCreate = jest.fn()
const mockSandboxApi = {
  createSandbox: jest.fn(),
  getSandbox: jest.fn(),
  listSandboxesPaginated: jest.fn(),
  getBuildLogsUrl: jest.fn(),
}
const mockSnapshotsApi = {}
const mockObjectStorageApi = {}
const mockConfigApi = {}
const mockVolumesApi = {}

const mockSnapshotServiceCtor = jest.fn()
const mockVolumeServiceCtor = jest.fn()
const mockSandboxCtor = jest.fn()

jest.mock('axios', () => {
  class MockAxiosError extends Error {}
  return {
    __esModule: true,
    default: {
      create: mockAxiosCreate,
      AxiosError: MockAxiosError,
    },
    create: mockAxiosCreate,
    AxiosError: MockAxiosError,
  }
})

jest.mock(
  '@daytonaio/api-client',
  () => {
    const configurationCtor = jest.fn().mockImplementation((args: Record<string, unknown>) => ({
      ...args,
      baseOptions: (args.baseOptions as Record<string, unknown>) ?? { headers: {} },
    }))

    return {
      __esModule: true,
      Configuration: configurationCtor,
      SandboxApi: jest.fn(() => mockSandboxApi),
      SnapshotsApi: jest.fn(() => mockSnapshotsApi),
      ObjectStorageApi: jest.fn(() => mockObjectStorageApi),
      ConfigApi: jest.fn(() => mockConfigApi),
      VolumesApi: jest.fn(() => mockVolumesApi),
      SandboxState: {
        PENDING_BUILD: 'pending_build',
        STARTED: 'started',
        STARTING: 'starting',
        ERROR: 'error',
        BUILD_FAILED: 'build_failed',
      },
    }
  },
  { virtual: true },
)

jest.mock('../Snapshot', () => ({
  SnapshotService: jest.fn().mockImplementation((...args: unknown[]) => {
    mockSnapshotServiceCtor(...args)
    return { list: jest.fn() }
  }),
}))

jest.mock('../Volume', () => ({
  VolumeService: jest.fn().mockImplementation((...args: unknown[]) => {
    mockVolumeServiceCtor(...args)
    return { list: jest.fn() }
  }),
}))

jest.mock('../Sandbox', () => ({
  Sandbox: jest.fn().mockImplementation((dto: { id: string; state?: string }) => {
    const sandbox = {
      ...dto,
      start: jest.fn(),
      stop: jest.fn(),
      delete: jest.fn(),
      waitUntilStarted: jest.fn(),
    }
    mockSandboxCtor(dto)
    return sandbox
  }),
}))

describe('Daytona', () => {
  beforeEach(() => {
    jest.resetModules()
    jest.clearAllMocks()

    process.env.DAYTONA_API_KEY = undefined
    process.env.DAYTONA_JWT_TOKEN = undefined
    process.env.DAYTONA_ORGANIZATION_ID = undefined
    process.env.DAYTONA_API_URL = undefined
    process.env.DAYTONA_TARGET = undefined

    mockAxiosCreate.mockReturnValue({
      defaults: { baseURL: 'http://sandbox-proxy/' },
      interceptors: {
        request: { use: jest.fn() },
        response: { use: jest.fn() },
      },
    })
  })

  it('uses explicit api key config and builds dependent services', async () => {
    const { Daytona } = await import('../Daytona')
    const { Configuration } = await import('@daytonaio/api-client')

    const instance = new Daytona({
      apiKey: 'api-key',
      apiUrl: 'https://example.daytona.test/api',
      target: 'eu',
    })

    expect(instance).toBeTruthy()
    expect(Configuration).toHaveBeenCalled()
    expect(mockVolumeServiceCtor).toHaveBeenCalledTimes(1)
    expect(mockSnapshotServiceCtor).toHaveBeenCalledTimes(1)
  })

  it('reads constructor values from env when config omitted', async () => {
    process.env.DAYTONA_API_KEY = 'env-key'
    process.env.DAYTONA_API_URL = 'https://env.daytona/api'
    process.env.DAYTONA_TARGET = 'us'

    const { Daytona } = await import('../Daytona')
    const { Configuration } = await import('@daytonaio/api-client')

    new Daytona()

    const firstConfigArg = (Configuration as jest.Mock).mock.calls[0][0] as {
      basePath: string
      baseOptions: { headers: Record<string, string> }
    }

    expect(firstConfigArg.basePath).toBe('https://env.daytona/api')
    expect(firstConfigArg.baseOptions.headers.Authorization).toBe('Bearer env-key')
  })

  it('supports jwt auth with organization header', async () => {
    const { Daytona } = await import('../Daytona')
    const { Configuration } = await import('@daytonaio/api-client')

    new Daytona({
      jwtToken: 'jwt-token',
      organizationId: 'org-1',
      apiUrl: 'https://jwt.daytona/api',
      target: 'us',
    })

    const firstConfigArg = (Configuration as jest.Mock).mock.calls[0][0] as {
      baseOptions: { headers: Record<string, string> }
    }
    expect(firstConfigArg.baseOptions.headers.Authorization).toBe('Bearer jwt-token')
    expect(firstConfigArg.baseOptions.headers['X-Daytona-Organization-ID']).toBe('org-1')
  })

  it('throws when jwt auth has no organization id', async () => {
    const { Daytona } = await import('../Daytona')
    process.env.DAYTONA_ORGANIZATION_ID = ''
    expect(
      () =>
        new Daytona({
          jwtToken: 'jwt-token',
          apiUrl: 'https://jwt.daytona/api',
          target: 'us',
        }),
    ).toThrow('Organization ID is required when using JWT token')
  })

  it('throws unsupported language in create', async () => {
    const { Daytona } = await import('../Daytona')
    const instance = new Daytona({ apiKey: 'k', apiUrl: 'http://api', target: 'us' })

    await expect(instance.create({ language: 'rust' })).rejects.toThrow('Unsupported language: rust')
  })

  test.each<
    [{ timeout?: number }, string, { language?: string; autoStopInterval?: number; autoArchiveInterval?: number }]
  >([
    [{ timeout: -1 }, 'Timeout must be a non-negative number', { language: 'python' }],
    [{}, 'autoStopInterval must be a non-negative integer', { autoStopInterval: -1 }],
    [{}, 'autoArchiveInterval must be a non-negative integer', { autoArchiveInterval: -1 }],
  ])('validates create input %#', async (optionsPart, message, params) => {
    const { Daytona } = await import('../Daytona')
    const instance = new Daytona({ apiKey: 'k', apiUrl: 'http://api', target: 'us' })

    await expect(instance.create(params, optionsPart)).rejects.toThrow(message)
  })

  it('forces autoDeleteInterval to 0 for ephemeral sandboxes', async () => {
    const { Daytona } = await import('../Daytona')
    const instance = new Daytona({ apiKey: 'k', apiUrl: 'http://api', target: 'us' })

    mockSandboxApi.createSandbox.mockResolvedValue(
      createApiResponse({ id: 'sb-1', state: 'started', labels: { 'code-toolbox-language': 'python' } }),
    )

    await instance.create({ language: 'python', ephemeral: true, autoDeleteInterval: 12 })

    const payload = mockSandboxApi.createSandbox.mock.calls[0][0] as { autoDeleteInterval?: number }
    expect(payload.autoDeleteInterval).toBe(0)
  })

  it('delegates get/list/start/stop/delete methods', async () => {
    const { Daytona } = await import('../Daytona')
    const instance = new Daytona({ apiKey: 'k', apiUrl: 'http://api', target: 'us' })

    mockSandboxApi.getSandbox.mockResolvedValue(
      createApiResponse({ id: 'sb-1', labels: { 'code-toolbox-language': 'python' } }),
    )
    mockSandboxApi.listSandboxesPaginated.mockResolvedValue(
      createApiResponse({
        items: [{ id: 'sb-1', labels: { 'code-toolbox-language': 'python' } }],
        total: 1,
        page: 1,
        totalPages: 1,
      }),
    )

    const sandboxFromGet = (await instance.get('sb-1')) as unknown as {
      id: string
      start: jest.Mock
      stop: jest.Mock
      delete: jest.Mock
    }
    expect(mockSandboxApi.getSandbox).toHaveBeenCalledWith('sb-1')

    const list = await instance.list({ project: 'sdk' }, 1, 10)
    expect(mockSandboxApi.listSandboxesPaginated).toHaveBeenCalled()
    expect(list.total).toBe(1)

    await instance.start(sandboxFromGet as unknown as never)
    await instance.stop(sandboxFromGet as unknown as never)
    await instance.delete(sandboxFromGet as unknown as never, 33)

    expect(sandboxFromGet.start).toHaveBeenCalled()
    expect(sandboxFromGet.stop).toHaveBeenCalled()
    expect(sandboxFromGet.delete).toHaveBeenCalledWith(33)
  })
})
