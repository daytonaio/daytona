// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import type { Configuration } from '@daytona/toolbox-api-client'
import { createApiResponse } from './helpers'

const mockCreateSandboxWebSocket = jest.fn()
const mockStdDemuxStream = jest.fn()
const mockParseChart = jest.fn((chart: { title?: string }) => ({ ...chart, parsed: true }))
const mockPtyHandleCtor = jest.fn()

jest.mock('@daytona/toolbox-api-client', () => ({}), { virtual: true })
jest.mock('../utils/WebSocket', () => ({
  createSandboxWebSocket: (...args: unknown[]) => mockCreateSandboxWebSocket(...args),
}))
jest.mock('../utils/Stream', () => ({
  stdDemuxStream: (...args: unknown[]) => mockStdDemuxStream(...args),
}))
jest.mock('../types/Charts', () => ({
  parseChart: (chart: unknown) => mockParseChart(chart),
}))
jest.mock('../PtyHandle', () => ({
  PtyHandle: jest.fn().mockImplementation((...args: unknown[]) => {
    mockPtyHandleCtor(...args)
    return {
      waitForConnection: jest.fn().mockResolvedValue(undefined),
      resize: jest.fn(),
      kill: jest.fn(),
      disconnect: jest.fn(),
      sessionId: args[4],
    }
  }),
}))

describe('Process', () => {
  const makeProcess = async (language = 'python') => {
    const { Process } = await import('../Process')
    const apiClient = {
      executeCommand: jest.fn(),
      codeRun: jest.fn(),
      createSession: jest.fn(),
      getSession: jest.fn(),
      getEntrypointSession: jest.fn(),
      sessionExecuteCommand: jest.fn(),
      getSessionCommand: jest.fn(),
      getSessionCommandLogs: jest.fn(),
      getEntrypointLogs: jest.fn(),
      listSessions: jest.fn(),
      deleteSession: jest.fn(),
      sendInput: jest.fn(),
      createPtySession: jest.fn(),
      listPtySessions: jest.fn(),
      getPtySession: jest.fn(),
      deletePtySession: jest.fn(),
      resizePtySession: jest.fn(),
    }

    const cfg: Configuration = {
      basePath: 'http://sandbox',
      baseOptions: { headers: { Authorization: 'Bearer t' } },
    } as unknown as Configuration

    const process = new Process(cfg, apiClient as unknown as never, async () => 'preview-token', language)

    return { process, apiClient }
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('executeCommand sends command and returns result', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.executeCommand.mockResolvedValue(createApiResponse({ exitCode: 0, result: 'hello' }))
    const result = await process.executeCommand('echo hi', '/tmp', { GOOD_KEY: '1' }, 4)
    expect(result).toMatchObject({ exitCode: 0, result: 'hello' })
    expect(apiClient.executeCommand).toHaveBeenCalledWith({
      command: 'echo hi',
      timeout: 4,
      cwd: '/tmp',
      envs: { GOOD_KEY: '1' },
    })
  })

  it('executeCommand omits envs when empty', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.executeCommand.mockResolvedValue(createApiResponse({ exitCode: 0, result: '' }))
    await process.executeCommand('ls')
    expect(apiClient.executeCommand).toHaveBeenCalledWith({
      command: 'ls',
      timeout: undefined,
      cwd: undefined,
      envs: undefined,
    })
  })

  it('executeCommand returns artifacts with stdout', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.executeCommand.mockResolvedValue(createApiResponse({ exitCode: 0, result: 'output' }))
    const result = await process.executeCommand('test')
    expect(result.artifacts).toEqual({ stdout: 'output' })
  })

  it('executeCommand falls back to code and empty result defaults', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.executeCommand.mockResolvedValue(createApiResponse({ code: 17 }))

    await expect(process.executeCommand('false')).resolves.toEqual({
      exitCode: 17,
      result: '',
      artifacts: { stdout: '' },
    })
  })

  it('codeRun delegates to apiClient.codeRun', async () => {
    const { process, apiClient } = await makeProcess()
    apiClient.codeRun.mockResolvedValue(createApiResponse({ exitCode: 0, result: 'ok', artifacts: { charts: [] } }))
    const result = await process.codeRun('print(1)')
    expect(result.exitCode).toBe(0)
    expect(result.result).toBe('ok')
    expect(apiClient.codeRun).toHaveBeenCalledWith({
      code: 'print(1)',
      language: 'python',
      argv: undefined,
      envs: undefined,
      timeout: undefined,
    })
  })

  it('codeRun parses chart artifacts and defaults missing fields', async () => {
    const { process, apiClient } = await makeProcess()
    apiClient.codeRun.mockResolvedValue(
      createApiResponse({
        artifacts: { charts: [{ title: 'Chart 1' }] },
      }),
    )

    await expect(process.codeRun('print(1)')).resolves.toEqual({
      exitCode: 0,
      result: '',
      artifacts: {
        stdout: '',
        charts: [{ title: 'Chart 1', parsed: true }],
      },
    })
    expect(mockParseChart).toHaveBeenCalledWith({ title: 'Chart 1' })
  })

  it('codeRun passes argv and env params', async () => {
    const { process, apiClient } = await makeProcess('typescript')
    apiClient.codeRun.mockResolvedValue(createApiResponse({ exitCode: 0, result: '', artifacts: {} }))
    await process.codeRun('console.log(1)', { argv: ['--flag'], env: { NODE_ENV: 'test' } }, 10)
    expect(apiClient.codeRun).toHaveBeenCalledWith({
      code: 'console.log(1)',
      language: 'typescript',
      argv: ['--flag'],
      envs: { NODE_ENV: 'test' },
      timeout: 10,
    })
  })

  it('codeRun throws when language not set', async () => {
    const { Process } = await import('../Process')
    const cfg = { basePath: 'http://sandbox', baseOptions: { headers: {} } } as unknown as Configuration
    const proc = new Process(cfg, {} as never, async () => 'token', undefined)
    await expect(proc.codeRun('x')).rejects.toThrow('Code language is required for codeRun')
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

  it('executeSessionCommand returns stdout/stderr', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.sessionExecuteCommand.mockResolvedValue(
      createApiResponse({ cmdId: 'c1', output: 'combined', exitCode: 0, stdout: 'std-out', stderr: 'std-err' }),
    )

    const execResult = await process.executeSessionCommand('s1', { command: 'ls' })
    expect(execResult.stdout).toBe('std-out')
    expect(execResult.stderr).toBe('std-err')
    expect(execResult.cmdId).toBe('c1')
  })

  it('executeSessionCommand converts timeout seconds to milliseconds', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.sessionExecuteCommand.mockResolvedValue(createApiResponse({ cmdId: 'c2' }))

    await process.executeSessionCommand('s1', { command: 'ls' }, 5)

    expect(apiClient.sessionExecuteCommand).toHaveBeenCalledWith('s1', { command: 'ls' }, { timeout: 5000 })
  })

  it('executeSessionCommand defaults stdout and stderr to empty strings', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.sessionExecuteCommand.mockResolvedValue(createApiResponse({ cmdId: 'c3', output: 'only-output' }))

    await expect(process.executeSessionCommand('s1', { command: 'ls' })).resolves.toEqual({
      cmdId: 'c3',
      output: 'only-output',
      stdout: '',
      stderr: '',
    })
  })

  it('getSessionCommandLogs returns output/stdout/stderr', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.getSessionCommandLogs.mockResolvedValue(
      createApiResponse({ output: 'combined', stdout: 'std-out', stderr: 'std-err' }),
    )

    const logs = await process.getSessionCommandLogs('s1', 'c1')
    expect(logs).toEqual({ output: 'combined', stdout: 'std-out', stderr: 'std-err' })
  })

  it('getSessionCommandLogs defaults missing log fields to empty strings', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.getSessionCommandLogs.mockResolvedValue(createApiResponse({}))

    await expect(process.getSessionCommandLogs('s1', 'c1')).resolves.toEqual({ output: '', stdout: '', stderr: '' })
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

  it('getEntrypointLogs returns buffered logs', async () => {
    const { process, apiClient } = await makeProcess()
    apiClient.getEntrypointLogs.mockResolvedValue(createApiResponse({ output: 'all', stdout: 'out', stderr: 'err' }))

    await expect(process.getEntrypointLogs()).resolves.toEqual({ output: 'all', stdout: 'out', stderr: 'err' })
  })

  it('streaming getEntrypointLogs uses websocket + stdDemuxStream', async () => {
    const { process } = await makeProcess()
    const ws = { close: jest.fn() }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)
    mockStdDemuxStream.mockResolvedValue(undefined)

    const onStdout = jest.fn()
    const onStderr = jest.fn()

    await process.getEntrypointLogs(onStdout, onStderr)

    expect(mockCreateSandboxWebSocket).toHaveBeenCalledWith(
      'ws://sandbox/process/session/entrypoint/logs?follow=true',
      { Authorization: 'Bearer t' },
      expect.any(Function),
    )
    expect(mockStdDemuxStream).toHaveBeenCalledWith(ws, onStdout, onStderr)
  })

  it('getEntrypointSession delegates', async () => {
    const { process, apiClient } = await makeProcess()
    apiClient.getEntrypointSession.mockResolvedValue(createApiResponse({ sessionId: 'entrypoint', commands: [] }))
    await expect(process.getEntrypointSession()).resolves.toEqual({ sessionId: 'entrypoint', commands: [] })
  })

  it('getSessionCommand delegates', async () => {
    const { process, apiClient } = await makeProcess()
    apiClient.getSessionCommand.mockResolvedValue(createApiResponse({ id: 'cmd-1', command: 'ls', exitCode: 0 }))
    await expect(process.getSessionCommand('s1', 'cmd-1')).resolves.toEqual({
      id: 'cmd-1',
      command: 'ls',
      exitCode: 0,
    })
  })

  it('creates a PTY session and connects to it', async () => {
    const { process, apiClient } = await makeProcess()
    const ws = { close: jest.fn() }

    apiClient.createPtySession.mockResolvedValue(createApiResponse({ sessionId: 'pty-1' }))
    mockCreateSandboxWebSocket.mockResolvedValue(ws)

    const onData = jest.fn()
    const handle = await process.createPty({
      id: 'pty-1',
      cwd: '/tmp',
      envs: { TERM: 'xterm-256color' },
      cols: 120,
      rows: 40,
      onData,
    })

    expect(apiClient.createPtySession).toHaveBeenCalledWith({
      id: 'pty-1',
      cwd: '/tmp',
      envs: { TERM: 'xterm-256color' },
      cols: 120,
      rows: 40,
      lazyStart: true,
    })
    expect(mockCreateSandboxWebSocket).toHaveBeenCalledWith(
      'ws://sandbox/process/pty/pty-1/connect',
      { Authorization: 'Bearer t' },
      expect.any(Function),
    )
    expect(handle.sessionId).toBe('pty-1')
  })

  it('connectPty wires resize and kill callbacks into PtyHandle', async () => {
    const { process, apiClient } = await makeProcess()
    const ws = { close: jest.fn() }
    mockCreateSandboxWebSocket.mockResolvedValue(ws)
    apiClient.resizePtySession.mockResolvedValue(createApiResponse({ sessionId: 'pty-2', cols: 150, rows: 45 }))
    apiClient.deletePtySession.mockResolvedValue(createApiResponse(undefined))

    await process.connectPty('pty-2', { onData: jest.fn() })

    const ctorArgs = mockPtyHandleCtor.mock.calls[mockPtyHandleCtor.mock.calls.length - 1]
    const resizeHandler = ctorArgs[1] as (cols: number, rows: number) => Promise<unknown>
    const killHandler = ctorArgs[2] as () => Promise<void>

    await expect(resizeHandler(150, 45)).resolves.toEqual({ sessionId: 'pty-2', cols: 150, rows: 45 })
    await killHandler()

    expect(apiClient.resizePtySession).toHaveBeenCalledWith('pty-2', { cols: 150, rows: 45 })
    expect(apiClient.deletePtySession).toHaveBeenCalledWith('pty-2')
  })

  it('lists, gets, resizes and kills PTY sessions', async () => {
    const { process, apiClient } = await makeProcess()

    apiClient.listPtySessions.mockResolvedValue(createApiResponse({ sessions: [{ sessionId: 'pty-1' }] }))
    apiClient.getPtySession.mockResolvedValue(createApiResponse({ sessionId: 'pty-1', cols: 80, rows: 24 }))
    apiClient.resizePtySession.mockResolvedValue(createApiResponse({ sessionId: 'pty-1', cols: 100, rows: 30 }))
    apiClient.deletePtySession.mockResolvedValue(createApiResponse(undefined))

    await expect(process.listPtySessions()).resolves.toEqual([{ sessionId: 'pty-1' }])
    await expect(process.getPtySessionInfo('pty-1')).resolves.toEqual({ sessionId: 'pty-1', cols: 80, rows: 24 })
    await expect(process.resizePtySession('pty-1', 100, 30)).resolves.toEqual({
      sessionId: 'pty-1',
      cols: 100,
      rows: 30,
    })
    await process.killPtySession('pty-1')
  })
})
