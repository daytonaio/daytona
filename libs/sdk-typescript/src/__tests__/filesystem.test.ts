/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Readable, Writable, PassThrough } from 'stream'
import { FileSystem } from '../FileSystem'
import {
  DaytonaNotFoundError,
  DaytonaValidationError,
  UploadAbortedError,
  DownloadAbortedError,
} from '../errors/DaytonaError'

describe('FileSystem.downloadFile', () => {
  function newFileSystem() {
    const fileSystem = Object.create(FileSystem.prototype) as FileSystem & {
      downloadFiles: jest.Mock
    }

    fileSystem.downloadFiles = jest.fn().mockResolvedValue([
      {
        error: 'download failed',
        errorDetails: {
          errorCode: 'FILE_NOT_FOUND',
          message: 'missing file',
          statusCode: 404,
        },
        source: '/workspace/missing.txt',
      },
    ])

    return fileSystem
  }

  it('rethrows structured errors for buffered downloads', async () => {
    const fileSystem = newFileSystem()

    await expect(FileSystem.prototype.downloadFile.call(fileSystem, '/workspace/missing.txt')).rejects.toBeInstanceOf(
      DaytonaNotFoundError,
    )

    await expect(FileSystem.prototype.downloadFile.call(fileSystem, '/workspace/missing.txt')).rejects.toMatchObject({
      errorCode: 'FILE_NOT_FOUND',
      statusCode: 404,
    })
  })

  it('rethrows structured errors for streamed downloads', async () => {
    const fileSystem = newFileSystem()

    await expect(
      FileSystem.prototype.downloadFile.call(fileSystem, '/workspace/missing.txt', '/tmp/out.txt'),
    ).rejects.toBeInstanceOf(DaytonaNotFoundError)

    await expect(
      FileSystem.prototype.downloadFile.call(fileSystem, '/workspace/missing.txt', '/tmp/out.txt'),
    ).rejects.toMatchObject({
      errorCode: 'FILE_NOT_FOUND',
      statusCode: 404,
    })
  })
})

describe('FileSystem streaming options-bag', () => {
  function newFileSystemWithMocks(uploadFilesMock?: jest.Mock, downloadFilesMock?: jest.Mock): any {
    const fileSystem: any = Object.create(FileSystem.prototype)

    fileSystem.uploadFiles = uploadFilesMock ?? jest.fn().mockResolvedValue(undefined)
    fileSystem.downloadFiles =
      downloadFilesMock ??
      jest.fn().mockResolvedValue([{ source: '/f', result: Buffer.from('hello'), error: undefined }])
    fileSystem.apiClient = {
      uploadFiles: jest.fn().mockResolvedValue({ data: undefined }),
      downloadFiles: jest.fn(),
    }
    fileSystem.clientConfig = {
      basePath: 'http://localhost',
      baseOptions: { headers: {} },
    }

    return fileSystem
  }

  // a. uploadFile with options-bag dispatch — Buffer source delegates to uploadFiles
  it('a. uploadFile with Buffer source dispatches to uploadFiles', async () => {
    const fileSystem = newFileSystemWithMocks()
    await FileSystem.prototype.uploadFile.call(fileSystem, { source: Buffer.from('x'), destination: '/f' })
    expect(fileSystem.uploadFiles).toHaveBeenCalledWith([{ source: Buffer.from('x'), destination: '/f' }], 30 * 60)
  })

  // b. uploadFile with Readable source and onProgress — verifies progress callback called
  it('b. uploadFile with Readable source calls onProgress with increasing totals', async () => {
    const chunkSize = 1024
    const chunks = [Buffer.alloc(chunkSize, 0x61), Buffer.alloc(chunkSize, 0x62)]

    const mockReadable = new Readable({
      read() {
        for (const chunk of chunks) {
          this.push(chunk)
        }
        this.push(null)
      },
    })

    const progressValues: number[] = []
    const onProgress = jest.fn((bytes: number) => progressValues.push(bytes))

    // Mock apiClient.uploadFiles to consume the readable stream (drain it)
    const uploadFilesMock = jest.fn().mockImplementation(async (config: any) => {
      const formData = config.data
      // drain the form-data stream
      await new Promise<void>((resolve, reject) => {
        const pipe = formData.pipe(new PassThrough())
        pipe.resume()
        pipe.on('end', resolve)
        pipe.on('error', reject)
      })
      return { data: undefined }
    })

    const fileSystem = newFileSystemWithMocks()
    fileSystem.apiClient.uploadFiles = uploadFilesMock

    await FileSystem.prototype.uploadFile.call(fileSystem, {
      source: mockReadable,
      destination: '/f',
      onProgress,
    })

    expect(onProgress).toHaveBeenCalled()
    // Progress values should be increasing
    for (let i = 1; i < progressValues.length; i++) {
      expect(progressValues[i]).toBeGreaterThanOrEqual(progressValues[i - 1])
    }
    // Final value should be total bytes
    const totalBytes = chunks.reduce((sum, c) => sum + c.length, 0)
    expect(progressValues[progressValues.length - 1]).toBe(totalBytes)
  })

  // c. uploadFile aborted before call — already-aborted signal throws UploadAbortedError
  it('c. uploadFile with already-aborted signal throws UploadAbortedError', async () => {
    const fileSystem = newFileSystemWithMocks()
    const controller = new AbortController()
    controller.abort()

    const mockReadable = new Readable({
      read() {
        this.push(null)
      },
    })

    await expect(
      FileSystem.prototype.uploadFile.call(fileSystem, {
        source: mockReadable,
        destination: '/f',
        signal: controller.signal,
      }),
    ).rejects.toBeInstanceOf(UploadAbortedError)
  })

  // d. uploadFile onProgress + Buffer source — throws DaytonaValidationError
  it('d. uploadFile with onProgress and Buffer source throws DaytonaValidationError', async () => {
    const fileSystem = newFileSystemWithMocks()
    const onProgress = jest.fn()

    await expect(
      FileSystem.prototype.uploadFile.call(fileSystem, {
        source: Buffer.from('hello'),
        destination: '/f',
        onProgress,
      }),
    ).rejects.toBeInstanceOf(DaytonaValidationError)
  })

  // e. downloadFile with options-bag buffer return — returns buffer
  it('e. downloadFile with options-bag returns buffer', async () => {
    const expectedBuffer = Buffer.from('file-content')
    const downloadFilesMock = jest.fn().mockResolvedValue([{ source: '/f', result: expectedBuffer, error: undefined }])
    const fileSystem = newFileSystemWithMocks(undefined, downloadFilesMock)

    const result = await FileSystem.prototype.downloadFile.call(fileSystem, { remotePath: '/f' })
    expect(result).toEqual(expectedBuffer)
  })

  // f. downloadFile aborted before call — already-aborted signal throws DownloadAbortedError
  it('f. downloadFile with already-aborted signal throws DownloadAbortedError', async () => {
    const fileSystem = newFileSystemWithMocks()
    const controller = new AbortController()
    controller.abort()

    await expect(
      FileSystem.prototype.downloadFile.call(fileSystem, {
        remotePath: '/f',
        signal: controller.signal,
      }),
    ).rejects.toBeInstanceOf(DownloadAbortedError)
  })

  // g. downloadFile onProgress + no stream destination — throws DaytonaValidationError
  it('g. downloadFile with onProgress and no stream destination throws DaytonaValidationError', async () => {
    const fileSystem = newFileSystemWithMocks()
    const onProgress = jest.fn()

    await expect(
      FileSystem.prototype.downloadFile.call(fileSystem, {
        remotePath: '/f',
        onProgress,
        // no destination (buffer return path)
      }),
    ).rejects.toBeInstanceOf(DaytonaValidationError)
  })

  // h. 2GB heap-flat test — streaming upload doesn't buffer all data in memory
  it('h. 2GB heap-flat: streaming upload keeps heap delta under 100MB', async () => {
    const CHUNK_COUNT = 100
    const CHUNK_SIZE = 20 * 1024 * 1024 // 20MB per chunk

    let chunksEmitted = 0
    const mockReadable = new Readable({
      read() {
        if (chunksEmitted < CHUNK_COUNT) {
          chunksEmitted++
          // Use Buffer.alloc lazily to avoid pre-allocating 2GB
          this.push(Buffer.alloc(CHUNK_SIZE, 0))
        } else {
          this.push(null)
        }
      },
    })

    // Mock apiClient.uploadFiles to drain the form-data stream without buffering
    const uploadFilesMock = jest.fn().mockImplementation(async (config: any) => {
      const formData = config.data
      await new Promise<void>((resolve, reject) => {
        const sink = new Writable({
          write(_chunk, _enc, cb) {
            cb()
          },
        })
        formData.pipe(sink)
        sink.on('finish', resolve)
        sink.on('error', reject)
        formData.on('error', reject)
      })
      return { data: undefined }
    })

    const fileSystem = newFileSystemWithMocks()
    fileSystem.apiClient.uploadFiles = uploadFilesMock

    if (global.gc) global.gc()
    const heapBefore = process.memoryUsage().heapUsed

    await FileSystem.prototype.uploadFile.call(fileSystem, {
      source: mockReadable,
      destination: '/large-file.bin',
    })

    if (global.gc) global.gc()
    const heapAfter = process.memoryUsage().heapUsed
    const heapDeltaMB = (heapAfter - heapBefore) / (1024 * 1024)

    // The heap delta should be well under 100MB — data flows through, not buffered
    expect(heapDeltaMB).toBeLessThan(100)
  }, 60000) // 60s timeout for large data
})
