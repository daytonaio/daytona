import type { Configuration } from '@daytonaio/api-client'

const mockCreateSandboxWebSocket = jest.fn()

jest.mock('@daytonaio/toolbox-api-client', () => ({}), { virtual: true })
jest.mock('../utils/WebSocket', () => ({
  createSandboxWebSocket: (...args: unknown[]) => mockCreateSandboxWebSocket(...args),
}))

describe('CodeInterpreter', () => {
  const cfg: Configuration = {
    basePath: 'http://sandbox',
    baseOptions: { headers: { Authorization: 'Bearer token' } },
  } as unknown as Configuration

  const apiClient = {
    createInterpreterContext: jest.fn(),
    listInterpreterContexts: jest.fn(),
    deleteInterpreterContext: jest.fn(),
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  const makeInterpreter = async () => {
    const { CodeInterpreter } = await import('../CodeInterpreter')
    return new CodeInterpreter(cfg, apiClient as unknown as never, async () => 'preview')
  }

  it('validates code input', async () => {
    const interpreter = await makeInterpreter()
    await expect(interpreter.runCode('  ')).rejects.toThrow('Code is required for execution')
  })

  it('streams stdout/stderr/error chunks and resolves on completed control message', async () => {
    const handlers: Record<string, (event?: unknown) => unknown> = {}
    const ws = {
      readyState: 1,
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn((event: string, cb: (event?: unknown) => unknown) => {
        handlers[event] = cb
      }),
      removeEventListener: jest.fn(),
    }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)

    const interpreter = await makeInterpreter()
    const onStdout = jest.fn()
    const onStderr = jest.fn()
    const onError = jest.fn()

    const runPromise = interpreter.runCode('print(1)', { onStdout, onStderr, onError, timeout: 3 })
    await Promise.resolve()

    handlers.open()
    await handlers.message?.({ data: JSON.stringify({ type: 'stdout', text: 'ok-out' }) })
    await handlers.message?.({ data: JSON.stringify({ type: 'stderr', text: 'ok-err' }) })
    await handlers.message?.({
      data: JSON.stringify({ type: 'error', name: 'ValueError', value: 'bad', traceback: 'tb' }),
    })
    await handlers.message?.({ data: JSON.stringify({ type: 'control', text: 'completed' }) })

    const result = await runPromise
    expect(result.stdout).toBe('ok-out')
    expect(result.stderr).toBe('ok-err')
    expect(result.error).toEqual({ name: 'ValueError', value: 'bad', traceback: 'tb' })
    expect(onStdout).toHaveBeenCalled()
    expect(onStderr).toHaveBeenCalled()
    expect(onError).toHaveBeenCalled()
  })

  it('throws DaytonaTimeoutError when websocket closes with timeout code', async () => {
    const handlers: Record<string, (event?: unknown) => unknown> = {}
    const ws = {
      readyState: 1,
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn((event: string, cb: (event?: unknown) => unknown) => {
        handlers[event] = cb
      }),
      removeEventListener: jest.fn(),
    }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)

    const interpreter = await makeInterpreter()
    const runPromise = interpreter.runCode('print(1)', { timeout: 1 })
    await Promise.resolve()

    handlers.open()
    handlers.close?.({ code: 4008, reason: '' })

    await expect(runPromise).rejects.toThrow('Execution timed out')
  })

  it('delegates context CRUD methods', async () => {
    const interpreter = await makeInterpreter()
    apiClient.createInterpreterContext.mockResolvedValue({ data: { id: 'ctx-1', cwd: '/tmp' } })
    apiClient.listInterpreterContexts.mockResolvedValue({ data: { contexts: [{ id: 'ctx-1' }] } })
    apiClient.deleteInterpreterContext.mockResolvedValue({ data: undefined })

    const ctx = await interpreter.createContext('/tmp')
    expect(ctx).toEqual({ id: 'ctx-1', cwd: '/tmp' })
    await expect(interpreter.listContexts()).resolves.toEqual([{ id: 'ctx-1' }])
    await interpreter.deleteContext({
      id: 'ctx-1',
      active: true,
      createdAt: '2026-01-01T00:00:00Z',
      cwd: '/tmp',
      language: 'python',
    })
  })
})
