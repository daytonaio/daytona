// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

import * as fs from 'fs'
import * as os from 'os'
import * as pathe from 'pathe'
import { DaytonaNotFoundError } from '../errors/DaytonaError'

const mockSend = jest.fn()
const mockUploadDone = jest.fn()
const mockTarCreate = jest.fn()
const mockDynamicImport = jest.fn()

class MockS3Client {
  public readonly config: Record<string, unknown>

  constructor(config: Record<string, unknown>) {
    this.config = config
  }

  send(command: { input: Record<string, unknown> }) {
    return mockSend(command)
  }
}

class MockUpload {
  public readonly params: Record<string, unknown>

  constructor(params: Record<string, unknown>) {
    this.params = params
  }

  done() {
    return mockUploadDone()
  }
}

jest.mock('@aws-sdk/client-s3', () => ({
  S3Client: MockS3Client,
  ListObjectsV2Command: class ListObjectsV2Command {
    input: Record<string, unknown>

    constructor(input: Record<string, unknown>) {
      this.input = input
    }
  },
}))

jest.mock('@aws-sdk/lib-storage', () => ({
  Upload: MockUpload,
}))

jest.mock('../utils/Import', () => ({
  dynamicImport: (...args: unknown[]) => mockDynamicImport(...args),
}))

describe('ObjectStorage', () => {
  const originalFs = fs
  const originalStream = require('stream') as typeof import('stream')
  let tempDir = ''

  const makeStorage = async (endpointUrl = 'https://s3.eu-central-1.amazonaws.com') => {
    const { ObjectStorage } = await import('../ObjectStorage')
    return new ObjectStorage({
      endpointUrl,
      accessKeyId: 'key',
      secretAccessKey: 'secret',
      sessionToken: 'session',
      bucketName: 'custom-bucket',
    })
  }

  beforeEach(async () => {
    jest.clearAllMocks()
    tempDir = await fs.promises.mkdtemp(pathe.join(os.tmpdir(), 'daytona-object-storage-'))

    mockDynamicImport.mockImplementation(async (moduleName: string) => {
      if (moduleName === 'fs') return originalFs
      if (moduleName === 'stream') return originalStream
      if (moduleName === 'tar') {
        return {
          create: mockTarCreate.mockReturnValue({
            pipe: jest.fn(),
          }),
        }
      }
      throw new Error(`Unexpected module: ${moduleName}`)
    })
  })

  afterEach(async () => {
    if (tempDir) {
      await fs.promises.rm(tempDir, { recursive: true, force: true })
    }
  })

  it('configures the s3 client with extracted region and credentials', async () => {
    const storage = await makeStorage('https://s3.us-west-2.amazonaws.com')

    const client = storage as unknown as { s3Client: MockS3Client }
    expect(client.s3Client.config).toMatchObject({
      region: 'us-west-2',
      endpoint: 'https://s3.us-west-2.amazonaws.com',
      forcePathStyle: true,
      credentials: {
        accessKeyId: 'key',
        secretAccessKey: 'secret',
        sessionToken: 'session',
      },
    })
  })

  it('falls back to the default region when the endpoint does not encode one', async () => {
    const storage = await makeStorage('https://storage.example.com')

    const client = storage as unknown as { s3Client: MockS3Client }
    expect(client.s3Client.config.region).toBe('us-east-1')
  })

  it('throws when uploading a missing path', async () => {
    const storage = await makeStorage()

    await expect(storage.upload(pathe.join(tempDir, 'missing'), 'org-1', '.')).rejects.toBeInstanceOf(
      DaytonaNotFoundError,
    )
    await expect(storage.upload(pathe.join(tempDir, 'missing'), 'org-1', '.')).rejects.toThrow(
      `Path does not exist: ${pathe.join(tempDir, 'missing')}`,
    )
  })

  it('returns the existing hash without re-uploading when the prefix already exists', async () => {
    const filePath = pathe.join(tempDir, 'a.txt')
    await fs.promises.writeFile(filePath, 'hello')

    const storage = await makeStorage()
    const storageRuntime = storage as unknown as {
      folderExistsInS3: jest.MockedFunction<(prefix: string) => Promise<boolean>>
      uploadAsTar: jest.MockedFunction<(s3Key: string, sourcePath: string, archiveBasePath: string) => Promise<void>>
    }

    const expectedHash = await (
      storage as unknown as { computeHashForPathMd5: (pathStr: string, archiveBasePath: string) => Promise<string> }
    ).computeHashForPathMd5(filePath, '.')

    storageRuntime.folderExistsInS3 = jest.fn().mockResolvedValue(true)
    storageRuntime.uploadAsTar = jest.fn()

    await expect(storage.upload(filePath, 'org-1', '.')).resolves.toBe(expectedHash)
    expect(storageRuntime.folderExistsInS3).toHaveBeenCalledWith(`org-1/${expectedHash}/`)
    expect(storageRuntime.uploadAsTar).not.toHaveBeenCalled()
  })

  it('uploads a tarball when the prefix does not exist', async () => {
    const filePath = pathe.join(tempDir, 'b.txt')
    await fs.promises.writeFile(filePath, 'content')

    const storage = await makeStorage()
    const storageRuntime = storage as unknown as {
      folderExistsInS3: jest.MockedFunction<(prefix: string) => Promise<boolean>>
      uploadAsTar: jest.MockedFunction<(s3Key: string, sourcePath: string, archiveBasePath: string) => Promise<void>>
    }

    const expectedHash = await (
      storage as unknown as { computeHashForPathMd5: (pathStr: string, archiveBasePath: string) => Promise<string> }
    ).computeHashForPathMd5(filePath, '.')

    storageRuntime.folderExistsInS3 = jest.fn().mockResolvedValue(false)
    storageRuntime.uploadAsTar = jest.fn().mockResolvedValue(undefined)

    await expect(storage.upload(filePath, 'org-2', '.')).resolves.toBe(expectedHash)
    expect(storageRuntime.uploadAsTar).toHaveBeenCalledWith(`org-2/${expectedHash}/context.tar`, filePath, '.')
  })

  it('computes stable hashes for files', async () => {
    const filePath = pathe.join(tempDir, 'stable.txt')
    await fs.promises.writeFile(filePath, 'abc123')

    const storage = await makeStorage()
    const runtime = storage as unknown as {
      computeHashForPathMd5: (pathStr: string, archiveBasePath: string) => Promise<string>
    }

    const hashA = await runtime.computeHashForPathMd5(filePath, '.')
    const hashB = await runtime.computeHashForPathMd5(filePath, '.')
    const hashC = await runtime.computeHashForPathMd5(filePath, 'other-base')

    expect(hashA).toHaveLength(32)
    expect(hashA).toBe(hashB)
    expect(hashC).not.toBe(hashA)
  })

  it('hashes directories recursively and includes nested content', async () => {
    const dirPath = pathe.join(tempDir, 'dir')
    await fs.promises.mkdir(pathe.join(dirPath, 'nested'), { recursive: true })
    await fs.promises.writeFile(pathe.join(dirPath, 'nested', 'a.txt'), 'a')
    await fs.promises.writeFile(pathe.join(dirPath, 'nested', 'b.txt'), 'b')

    const storage = await makeStorage()
    const runtime = storage as unknown as {
      computeHashForPathMd5: (pathStr: string, archiveBasePath: string) => Promise<string>
    }

    const hashBefore = await runtime.computeHashForPathMd5(dirPath, 'dir')
    await fs.promises.writeFile(pathe.join(dirPath, 'nested', 'b.txt'), 'updated')
    const hashAfter = await runtime.computeHashForPathMd5(dirPath, 'dir')

    expect(hashBefore).not.toBe(hashAfter)
  })

  it('hashes empty directories without throwing', async () => {
    const dirPath = pathe.join(tempDir, 'empty-dir')
    await fs.promises.mkdir(dirPath)

    const storage = await makeStorage()
    const runtime = storage as unknown as {
      computeHashForPathMd5: (pathStr: string, archiveBasePath: string) => Promise<string>
    }

    await expect(runtime.computeHashForPathMd5(dirPath, 'empty-dir')).resolves.toHaveLength(32)
  })

  it('checks folder existence using ListObjectsV2Command', async () => {
    const storage = await makeStorage()
    const runtime = storage as unknown as {
      folderExistsInS3: (prefix: string) => Promise<boolean>
    }

    mockSend.mockResolvedValueOnce({ Contents: [{ Key: 'a' }] })
    await expect(runtime.folderExistsInS3('org/hash/')).resolves.toBe(true)

    expect(mockSend).toHaveBeenCalledWith(
      expect.objectContaining({
        input: {
          Bucket: 'custom-bucket',
          Prefix: 'org/hash/',
          MaxKeys: 1,
        },
      }),
    )

    mockSend.mockResolvedValueOnce({ Contents: [] })
    await expect(runtime.folderExistsInS3('org/hash/')).resolves.toBe(false)
  })

  it('creates and uploads a tar stream with archive base path', async () => {
    const filePath = pathe.join(tempDir, 'archive', 'a.txt')
    await fs.promises.mkdir(pathe.dirname(filePath), { recursive: true })
    await fs.promises.writeFile(filePath, 'archive-me')

    mockTarCreate.mockReturnValue({ pipe: jest.fn() })
    mockUploadDone.mockResolvedValue(undefined)

    const storage = await makeStorage()
    const runtime = storage as unknown as {
      uploadAsTar: (s3Key: string, sourcePath: string, archiveBasePath: string) => Promise<void>
    }

    await runtime.uploadAsTar('org/hash/context.tar', filePath, '.')

    expect(mockTarCreate).toHaveBeenCalledWith(
      expect.objectContaining({
        cwd: pathe.resolve(filePath),
        portable: true,
        gzip: false,
      }),
      ['.'],
    )
    expect(mockUploadDone).toHaveBeenCalledTimes(1)
  })

  it('extracts aws regions from supported s3 endpoints', async () => {
    const storage = await makeStorage()
    const runtime = storage as unknown as {
      extractAwsRegion: (endpoint: string) => string | undefined
    }

    expect(runtime.extractAwsRegion('https://s3.us-east-1.amazonaws.com')).toBe('us-east-1')
    expect(runtime.extractAwsRegion('https://s3-eu-west-1.amazonaws.com')).toBe('eu-west-1')
    expect(runtime.extractAwsRegion('https://files.example.com')).toBeUndefined()
  })
})
