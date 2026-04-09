/*
 * Copyright 2025 Daytona Platforms Inc.
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
