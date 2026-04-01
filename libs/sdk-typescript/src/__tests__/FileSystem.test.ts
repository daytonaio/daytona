import type { Configuration } from '@daytonaio/toolbox-api-client'
import { createApiResponse } from './helpers'

const mockProcessWithBusboy = jest.fn()
const mockProcessWithBuffered = jest.fn()
const mockNormalizeResponseStream = jest.fn((x: unknown) => x)
const mockDynamicImport = jest.fn()

jest.mock('@daytonaio/toolbox-api-client', () => ({}), { virtual: true })
jest.mock('../utils/FileTransfer', () => ({
  normalizeResponseStream: (data: unknown) => mockNormalizeResponseStream(data),
  processDownloadFilesResponseWithBusboy: (...args: unknown[]) => mockProcessWithBusboy(...args),
  processDownloadFilesResponseWithBuffered: (...args: unknown[]) => mockProcessWithBuffered(...args),
}))
jest.mock('../utils/Import', () => ({
  dynamicImport: (...args: unknown[]) => mockDynamicImport(...args),
}))
jest.mock('../utils/Runtime', () => ({
  Runtime: {
    NODE: 'node',
    BROWSER: 'browser',
    SERVERLESS: 'serverless',
    DENO: 'deno',
  },
  RUNTIME: 'node',
}))

describe('FileSystem', () => {
  const makeFs = async () => {
    const { FileSystem } = await import('../FileSystem')
    const apiClient = {
      createFolder: jest.fn(),
      deleteFile: jest.fn(),
      downloadFiles: jest.fn(),
      findInFiles: jest.fn(),
      getFileInfo: jest.fn(),
      listFiles: jest.fn(),
      moveFile: jest.fn(),
      replaceInFiles: jest.fn(),
      searchFiles: jest.fn(),
      setFilePermissions: jest.fn(),
      uploadFiles: jest.fn(),
    }

    const cfg: Configuration = {
      basePath: 'http://sandbox',
      baseOptions: { headers: {} },
    } as unknown as Configuration

    return {
      fileSystem: new FileSystem(cfg, apiClient as unknown as never),
      apiClient,
    }
  }

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('delegates core file operations', async () => {
    const { fileSystem, apiClient } = await makeFs()

    apiClient.createFolder.mockResolvedValue(createApiResponse(undefined))
    apiClient.deleteFile.mockResolvedValue(createApiResponse(undefined))
    apiClient.findInFiles.mockResolvedValue(createApiResponse([{ file: 'a.ts', line: 1, content: 'x' }]))
    apiClient.getFileInfo.mockResolvedValue(createApiResponse({ path: '/a', size: 1 }))
    apiClient.listFiles.mockResolvedValue(createApiResponse([{ path: '/a' }]))
    apiClient.moveFile.mockResolvedValue(createApiResponse(undefined))
    apiClient.replaceInFiles.mockResolvedValue(createApiResponse([{ file: 'a.ts', count: 1 }]))
    apiClient.searchFiles.mockResolvedValue(createApiResponse({ files: ['a.ts'] }))
    apiClient.setFilePermissions.mockResolvedValue(createApiResponse(undefined))

    await fileSystem.createFolder('/tmp/a', '755')
    await fileSystem.deleteFile('/tmp/a', true)
    await expect(fileSystem.findFiles('/tmp', 'TODO')).resolves.toHaveLength(1)
    await expect(fileSystem.getFileDetails('/tmp/a')).resolves.toEqual({ path: '/a', size: 1 })
    await expect(fileSystem.listFiles('/tmp')).resolves.toEqual([{ path: '/a' }])
    await fileSystem.moveFiles('/tmp/a', '/tmp/b')
    await expect(fileSystem.replaceInFiles(['a.ts'], 'x', 'y')).resolves.toEqual([{ file: 'a.ts', count: 1 }])
    await expect(fileSystem.searchFiles('/tmp', '*.ts')).resolves.toEqual({ files: ['a.ts'] })
    await fileSystem.setFilePermissions('/tmp/a', { owner: 'u', group: 'g', mode: '644' })
  })

  it('downloadFile uses downloadFiles and throws on per-file error', async () => {
    const { fileSystem } = await makeFs()
    const spy = jest.spyOn(fileSystem, 'downloadFiles')

    spy.mockResolvedValueOnce([{ source: 'a.txt', result: Buffer.from('ok') }])
    await expect(fileSystem.downloadFile('a.txt')).resolves.toEqual(Buffer.from('ok'))

    spy.mockResolvedValueOnce([{ source: 'a.txt', error: 'boom' }])
    await expect(fileSystem.downloadFile('a.txt')).rejects.toThrow('boom')
  })

  it('downloadFiles processes multipart output in node mode', async () => {
    const { fileSystem, apiClient } = await makeFs()

    apiClient.downloadFiles.mockResolvedValue(
      createApiResponse('stream-content') as { data: string; headers: Record<string, string> },
    )

    mockProcessWithBusboy.mockImplementation(
      async (_stream: unknown, _headers: Record<string, string>, metadata: Map<string, { result?: Buffer }>) => {
        metadata.set('file-a.txt', { result: Buffer.from('a') })
      },
    )

    const result = await fileSystem.downloadFiles([{ source: 'file-a.txt' }], 10)
    expect(apiClient.downloadFiles).toHaveBeenCalled()
    expect(result[0].result).toEqual(Buffer.from('a'))
    expect(result[0].error).toBeUndefined()
  })

  it('uploadFile delegates to uploadFiles', async () => {
    const { fileSystem } = await makeFs()
    const uploadSpy = jest.spyOn(fileSystem, 'uploadFiles').mockResolvedValue()

    await fileSystem.uploadFile(Buffer.from('abc'), '/tmp/file.txt', 5)
    expect(uploadSpy).toHaveBeenCalledWith([{ source: Buffer.from('abc'), destination: '/tmp/file.txt' }], 5)
  })

  it('uploadFiles builds form payload and calls toolbox api', async () => {
    const { fileSystem, apiClient } = await makeFs()

    class FormDataStub {
      append = jest.fn()
    }

    const readablePayload = { stream: true }
    const streamModule = { Readable: { from: jest.fn(() => readablePayload) } }
    const fsModule = { createReadStream: jest.fn(() => ({ file: true })) }

    mockDynamicImport.mockImplementation(async (moduleName: string) => {
      if (moduleName === 'form-data') return FormDataStub
      if (moduleName === 'stream') return streamModule
      if (moduleName === 'fs') return fsModule
      return {}
    })

    apiClient.uploadFiles.mockResolvedValue(createApiResponse(undefined))

    await fileSystem.uploadFiles([{ source: Buffer.from('abc'), destination: '/tmp/a.txt' }], 7)

    expect(apiClient.uploadFiles).toHaveBeenCalled()
  })
})
