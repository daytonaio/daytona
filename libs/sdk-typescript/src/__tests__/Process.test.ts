import type { Configuration } from '@daytonaio/toolbox-api-client'
import { createApiResponse, toMuxedOutput } from './helpers'

const mockCreateSandboxWebSocket = jest.fn()
const mockStdDemuxStream = jest.fn()

jest.mock('@daytonaio/toolbox-api-client', () => ({}), { virtual: true })
jest.mock('../utils/WebSocket', () => ({
  createSandboxWebSocket: (...args: unknown[]) => mockCreateSandboxWebSocket(...args),
}))
jest.mock('../utils/Stream', () => ({
  stdDemuxStream: (...args: unknown[]) => mockStdDemuxStream(...args),
}))

describe('Process', () => {
  const makeProcess = async () => {
    const { Process } = await import('../Process')
    const apiClient = {
      executeCommand: jest.fn(),
      createSession: jest.fn(),
      getSession: jest.fn(),
      sessionExecuteCommand: jest.fn(),
      getSessionCommandLogs: jest.fn(),
      listSessions: jest.fn(),
      deleteSession: jest.fn(),
      sendInput: jest.fn(),
    }

    const cfg: Configuration = {
      basePath: 'http://sandbox',
      baseOptions: { headers: { Authorization: 'Bearer t' } },
    } as unknown as Configuration

    const process = new Process(
      cfg,
      { getRunCommand: (code: string) => `python -c ${JSON.stringify(code)}` },
      apiClient as unknown as never,
      async () => 'preview-token',
    )

    return { process, apiClient }
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('executeCommand validates env keys and parses artifacts', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.executeCommand.mockResolvedValue(createApiResponse({ exitCode: 0, result: 'hello' }))
    await expect(process.executeCommand('echo hi', '/tmp', { GOOD_KEY: '1' }, 4)).resolves.toMatchObject({
      exitCode: 0,
      result: 'hello',
    })

    await expect(process.executeCommand('echo hi', '/tmp', { 'BAD-KEY': '1' })).rejects.toThrow(
      "Invalid environment variable name: 'BAD-KEY'",
    )
  })

  it('codeRun delegates through code toolbox', async () => {
    const { process, apiClient } = await makeProcess()
    apiClient.executeCommand.mockResolvedValue(createApiResponse({ exitCode: 0, result: 'ok' }))
    const result = await process.codeRun('print(1)')
    expect(result.exitCode).toBe(0)
    expect(apiClient.executeCommand).toHaveBeenCalled()
  })

  it('session methods delegate to api client', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.createSession.mockResolvedValue(createApiResponse(undefined))
    apiClient.getSession.mockResolvedValue(createApiResponse({ sessionId: 's1', commands: [] }))
    apiClient.listSessions.mockResolvedValue(createApiResponse([{ sessionId: 's1', commands: [] }]))
    apiClient.deleteSession.mockResolvedValue(createApiResponse(undefined))
    apiClient.sendInput.mockResolvedValue(createApiResponse(undefined))

    await process.createSession('s1')
    await expect(process.getSession('s1')).resolves.toEqual({ sessionId: 's1', commands: [] })
    await expect(process.listSessions()).resolves.toHaveLength(1)
    await process.sendSessionCommandInput('s1', 'c1', 'input')
    await process.deleteSession('s1')
  })

  it('executeSessionCommand and getSessionCommandLogs demultiplex output', async () => {
    const { process, apiClient } = await makeProcess()
    const output = toMuxedOutput('std-out', 'std-err')

    apiClient.sessionExecuteCommand.mockResolvedValue(
      createApiResponse({ cmdId: 'c1', output, exitCode: 0, runAsync: false }),
    )
    apiClient.getSessionCommandLogs.mockResolvedValue(createApiResponse(output))

    const execResult = await process.executeSessionCommand('s1', { command: 'ls' })
    expect(execResult.stdout).toBe('std-out')
    expect(execResult.stderr).toBe('std-err')

    const logs = await process.getSessionCommandLogs('s1', 'c1')
    expect(logs).toEqual({ output, stdout: 'std-out', stderr: 'std-err' })
  })

  it('streaming getSessionCommandLogs uses websocket + stdDemuxStream', async () => {
    const { process } = await makeProcess()
    const ws = { close: jest.fn() }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)
    mockStdDemuxStream.mockResolvedValue(undefined)

    const onStdout = jest.fn()
    const onStderr = jest.fn()

    await process.getSessionCommandLogs('s1', 'c1', onStdout, onStderr)
    expect(mockCreateSandboxWebSocket).toHaveBeenCalled()
    expect(mockStdDemuxStream).toHaveBeenCalledWith(ws, onStdout, onStderr)
  })
})
