// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import { createApiResponse } from './helpers'
import { DaytonaValidationError } from '../errors/DaytonaError'

jest.mock('@daytona/toolbox-api-client', () => ({}), { virtual: true })

describe('LspServer', () => {
  const apiClient = {
    start: jest.fn(),
    stop: jest.fn(),
    didOpen: jest.fn(),
    didClose: jest.fn(),
    documentSymbols: jest.fn(),
    workspaceSymbols: jest.fn(),
    completions: jest.fn(),
  }

  const makeServer = async (languageId = 'typescript', pathToProject = '/workspace/project') => {
    const { LspServer } = await import('../LspServer')
    return new LspServer(languageId as never, pathToProject, apiClient as never)
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('validates language id in constructor', async () => {
    const { LspServer } = await import('../LspServer')

    expect(() => new LspServer('rust' as never, '/workspace/project', apiClient as never)).toThrow(
      DaytonaValidationError,
    )
    expect(() => new LspServer('rust' as never, '/workspace/project', apiClient as never)).toThrow(
      'Invalid languageId: rust. Supported values are: python, typescript, javascript',
    )
  })

  it('starts the server with constructor values', async () => {
    const server = await makeServer()
    apiClient.start.mockResolvedValue(createApiResponse(undefined))

    await server.start()

    expect(apiClient.start).toHaveBeenCalledWith({
      languageId: 'typescript',
      pathToProject: '/workspace/project',
    })
  })

  it('stops the server with constructor values', async () => {
    const server = await makeServer('python', '/workspace/py')
    apiClient.stop.mockResolvedValue(createApiResponse(undefined))

    await server.stop()

    expect(apiClient.stop).toHaveBeenCalledWith({
      languageId: 'python',
      pathToProject: '/workspace/py',
    })
  })

  it('opens files with a file uri', async () => {
    const server = await makeServer()
    apiClient.didOpen.mockResolvedValue(createApiResponse(undefined))

    await server.didOpen('/workspace/project/src/index.ts')

    expect(apiClient.didOpen).toHaveBeenCalledWith({
      languageId: 'typescript',
      pathToProject: '/workspace/project',
      uri: 'file:///workspace/project/src/index.ts',
    })
  })

  it('closes files with a file uri', async () => {
    const server = await makeServer('javascript')
    apiClient.didClose.mockResolvedValue(createApiResponse(undefined))

    await server.didClose('/workspace/project/src/index.js')

    expect(apiClient.didClose).toHaveBeenCalledWith({
      languageId: 'javascript',
      pathToProject: '/workspace/project',
      uri: 'file:///workspace/project/src/index.js',
    })
  })

  it('returns document symbols', async () => {
    const server = await makeServer()
    apiClient.documentSymbols.mockResolvedValue(createApiResponse([{ name: 'main', kind: 12 }]))

    await expect(server.documentSymbols('/workspace/project/src/index.ts')).resolves.toEqual([
      { name: 'main', kind: 12 },
    ])
    expect(apiClient.documentSymbols).toHaveBeenCalledWith(
      'typescript',
      '/workspace/project',
      'file:///workspace/project/src/index.ts',
    )
  })

  it('returns sandbox symbols for a query', async () => {
    const server = await makeServer()
    apiClient.workspaceSymbols.mockResolvedValue(createApiResponse([{ name: 'UserService', kind: 5 }]))

    await expect(server.sandboxSymbols('User')).resolves.toEqual([{ name: 'UserService', kind: 5 }])
    expect(apiClient.workspaceSymbols).toHaveBeenCalledWith('User', 'typescript', '/workspace/project')
  })

  it('workspaceSymbols delegates to sandboxSymbols', async () => {
    const server = await makeServer('python', '/workspace/app')
    apiClient.workspaceSymbols.mockResolvedValue(createApiResponse([{ name: 'foo', kind: 1 }]))
    const runtime = server as unknown as { workspaceSymbols: (query: string) => Promise<unknown[]> }

    await expect(runtime.workspaceSymbols('foo')).resolves.toEqual([{ name: 'foo', kind: 1 }])
    expect(apiClient.workspaceSymbols).toHaveBeenCalledWith('foo', 'python', '/workspace/app')
  })

  it('requests completions using the provided position', async () => {
    const server = await makeServer()
    apiClient.completions.mockResolvedValue(createApiResponse({ isIncomplete: false, items: [{ label: 'console' }] }))

    await expect(server.completions('/workspace/project/src/index.ts', { line: 5, character: 9 })).resolves.toEqual({
      isIncomplete: false,
      items: [{ label: 'console' }],
    })

    expect(apiClient.completions).toHaveBeenCalledWith({
      languageId: 'typescript',
      pathToProject: '/workspace/project',
      uri: 'file:///workspace/project/src/index.ts',
      position: {
        line: 5,
        character: 9,
      },
    })
  })

  it('supports zero line and character positions', async () => {
    const server = await makeServer()
    apiClient.completions.mockResolvedValue(createApiResponse({ isIncomplete: true, items: [] }))

    await server.completions('/workspace/project/src/index.ts', { line: 0, character: 0 })

    expect(apiClient.completions).toHaveBeenCalledWith(
      expect.objectContaining({
        position: { line: 0, character: 0 },
      }),
    )
  })

  it('propagates api errors from start', async () => {
    const server = await makeServer()
    const error = new Error('start failed')
    apiClient.start.mockRejectedValue(error)

    await expect(server.start()).rejects.toBe(error)
  })

  it('propagates api errors from completions', async () => {
    const server = await makeServer()
    const error = new Error('completion failed')
    apiClient.completions.mockRejectedValue(error)

    await expect(server.completions('/workspace/project/src/index.ts', { line: 1, character: 1 })).rejects.toBe(error)
  })
})
