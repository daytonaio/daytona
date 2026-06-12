/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { FileSystem } from '../FileSystem'
import { DaytonaNotFoundError } from '../errors/DaytonaError'

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

describe('FileSystem.listFiles', () => {
  function newFileSystem(listFilesMock: jest.Mock) {
    const fileSystem = Object.create(FileSystem.prototype) as FileSystem
    Object.defineProperty(fileSystem, 'apiClient', { value: { listFiles: listFilesMock } })
    return fileSystem
  }

  it.each([0, -1, 1.5, Number.NaN])('rejects invalid depth %p before calling the api', async (depth) => {
    const listFilesMock = jest.fn()
    const fileSystem = newFileSystem(listFilesMock)

    await expect(FileSystem.prototype.listFiles.call(fileSystem, '/workspace', { depth })).rejects.toThrow(
      'depth must be an integer of at least 1',
    )
    expect(listFilesMock).not.toHaveBeenCalled()
  })

  it('passes depth to the api client', async () => {
    const listFilesMock = jest.fn().mockResolvedValue({ data: [] })
    const fileSystem = newFileSystem(listFilesMock)

    await expect(FileSystem.prototype.listFiles.call(fileSystem, '/workspace', { depth: 2 })).resolves.toEqual([])
    expect(listFilesMock).toHaveBeenCalledWith('/workspace', 2)
  })

  it('omits depth when not provided', async () => {
    const listFilesMock = jest.fn().mockResolvedValue({ data: [] })
    const fileSystem = newFileSystem(listFilesMock)

    await expect(FileSystem.prototype.listFiles.call(fileSystem, '/workspace')).resolves.toEqual([])
    expect(listFilesMock).toHaveBeenCalledWith('/workspace', undefined)
  })
})
