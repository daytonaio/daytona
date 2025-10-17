/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Buffer } from 'buffer'
import busboy from 'busboy'
import { DaytonaError } from '../errors/DaytonaError'
import { dynamicImport } from './Import'
import { collectStreamBytes, toBuffer, toUint8Array } from './Binary'
import { extractBoundary, getHeader, parseMultipartWithFormData } from './Multipart'
import { parseMultipart } from './Multipart'
import { DownloadMetadata } from '../FileSystem'

/**
 * Safely aborts a stream
 */
export function abortStream(stream: any): void {
  if (stream && typeof stream.destroy === 'function') {
    stream.destroy()
  } else if (stream && typeof stream.cancel === 'function') {
    stream.cancel()
  }
}

/**
 * Normalizes response data to extract the actual stream
 */
export function normalizeResponseStream(responseData: any): any {
  if (!responseData || typeof responseData !== 'object') {
    return responseData
  }

  // WHATWG ReadableStream
  if (responseData.body && typeof responseData.body.getReader === 'function') {
    return responseData.body
  }

  // Some adapters use .stream
  if (responseData.stream) {
    return responseData.stream
  }

  return responseData
}

/**
 * Processes multipart response using busboy (Node.js path)
 */
export async function processDownloadFilesResponseWithBusboy(
  stream: any,
  headers: Record<string, string>,
  metadataMap: Map<string, DownloadMetadata>,
): Promise<void> {
  const fileTasks: Promise<void>[] = []

  await new Promise<void>((resolve, reject) => {
    const bb = busboy({
      headers,
      preservePath: true,
    })

    bb.on('file', (fieldName: string, fileStream: any, fileInfo: { filename?: string }) => {
      const source = fileInfo?.filename
      if (!source) {
        abortStream(stream)
        reject(new DaytonaError(`Received unexpected file "${fileInfo?.filename}".`))
        return
      }

      const metadata = metadataMap.get(source)
      if (!metadata) {
        abortStream(stream)
        reject(new DaytonaError(`Target metadata missing for valid source: ${source}`))
        return
      }

      if (fieldName === 'error') {
        // Collect error message
        const chunks: Buffer[] = []
        fileStream.on('data', (chunk: Buffer) => chunks.push(chunk))
        fileStream.on('end', () => {
          metadata.error = Buffer.concat(chunks).toString('utf-8').trim()
        })
        fileStream.on('error', (err: any) => {
          metadata.error = `Stream error: ${err.message}`
        })
      } else if (fieldName === 'file') {
        if (metadata.destination) {
          // Stream to file
          fileTasks.push(
            new Promise((resolveTask) => {
              dynamicImport('fs', 'Downloading files to local files is not supported: ').then((fs) => {
                const writeStream = fs.createWriteStream(metadata.destination!, { autoClose: true })
                fileStream.pipe(writeStream)
                writeStream.on('finish', () => {
                  metadata.result = metadata.destination!
                  resolveTask()
                })
                writeStream.on('error', (err: any) => {
                  metadata.error = `Write stream failed: ${err.message}`
                  resolveTask()
                })
                fileStream.on('error', (err: any) => {
                  metadata.error = `Read stream failed: ${err.message}`
                })
              })
            }),
          )
        } else {
          // Collect to buffer
          const chunks: Buffer[] = []
          fileStream.on('data', (chunk: Buffer) => {
            chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk))
          })
          fileStream.on('end', () => {
            metadata.result = Buffer.concat(chunks)
          })
          fileStream.on('error', (err: any) => {
            metadata.error = `Read failed: ${err.message}`
          })
        }
      } else {
        // Unknown field, drain it
        fileStream.resume()
      }
    })

    bb.on('error', (err: any) => {
      abortStream(stream)
      reject(err)
    })

    bb.on('finish', resolve)

    // Feed stream into busboy
    feedStreamToBusboy(stream, bb).catch((err) => bb.destroy(err as Error))
  })

  await Promise.all(fileTasks)
}

/**
 * Feeds various stream types into busboy
 */
async function feedStreamToBusboy(stream: any, bb: any): Promise<void> {
  // Node.js stream (piping)
  if (typeof stream?.pipe === 'function') {
    stream.pipe(bb)
    return
  }

  // Direct buffer-like data
  if (typeof stream === 'string' || stream instanceof ArrayBuffer || ArrayBuffer.isView(stream)) {
    const data = toUint8Array(stream)
    bb.write(Buffer.from(data))
    bb.end()
    return
  }

  // WHATWG ReadableStream
  if (typeof stream?.getReader === 'function') {
    const reader = stream.getReader()
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      bb.write(Buffer.from(value))
    }
    bb.end()
    return
  }

  // AsyncIterable
  if (stream?.[Symbol.asyncIterator]) {
    for await (const chunk of stream) {
      const buffer = Buffer.isBuffer(chunk) ? chunk : Buffer.from(toUint8Array(chunk))
      bb.write(buffer)
    }
    bb.end()
    return
  }

  // Unsupported stream type
  throw new DaytonaError(`Unsupported stream type: ${stream?.constructor?.name || typeof stream}`)
}

export async function processDownloadFilesResponseWithBuffered(
  stream: any,
  headers: Record<string, string>,
  metadataMap: Map<string, DownloadMetadata>,
): Promise<void> {
  const contentType = getHeader(headers, 'content-type') || ''
  const bodyBytes = await collectStreamBytes(stream)

  // Try native FormData parsing for multipart/form-data
  if (/^multipart\/form-data/i.test(contentType) && typeof Response !== 'undefined') {
    try {
      const formDataMap = await parseMultipartWithFormData(bodyBytes, contentType)

      formDataMap.forEach((value, fieldName) => {
        const metadata = metadataMap.get(value.filename)
        if (!metadata) return

        if (fieldName.includes('error')) {
          metadata.error = new TextDecoder('utf-8').decode(value.data).trim()
        } else {
          metadata.result = toBuffer(value.data)
        }
      })

      return
    } catch {
      // Fall through to manual parsing
    }
  }

  // Manual multipart parsing (handles multipart/mixed, etc.)
  const boundary = extractBoundary(contentType)
  if (!boundary) {
    throw new DaytonaError(`Missing multipart boundary in Content-Type: "${contentType}"`)
  }

  const parts = parseMultipart(bodyBytes, boundary)
  for (const part of parts) {
    if (!part.filename) continue
    const metadata = metadataMap.get(part.filename)
    if (!metadata) continue

    if (part.name === 'error') {
      metadata.error = new TextDecoder('utf-8').decode(part.data).trim()
    } else if (part.name === 'file') {
      metadata.result = toBuffer(part.data)
    }
  }

  return
}
