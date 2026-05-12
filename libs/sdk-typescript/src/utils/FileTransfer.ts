/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Readable } from 'stream'
import { DaytonaError } from '../errors/DaytonaError'
import { dynamicImport, dynamicRequire } from './Import'
import { collectStreamBytes, toBuffer, toUint8Array } from './Binary'
import { extractBoundary, getHeader, parseMultipartWithFormData } from './Multipart'
import { parseMultipart } from './Multipart'
import type { DownloadMetadata, FileDownloadErrorDetails, UploadProgress, UploadSource } from '../FileSystem'

type DownloadErrorPartResult = {
  message: string
  errorDetails?: FileDownloadErrorDetails
}

/**
 * Parses a bulk-download error part into the legacy message and structured metadata.
 */
function parseDownloadErrorPart(data: Uint8Array, contentType?: string): DownloadErrorPartResult {
  let message = new TextDecoder('utf-8').decode(data).trim()
  if (!contentType || !/application\/json/i.test(contentType)) {
    return { message }
  }

  try {
    const payload = JSON.parse(message) as unknown
    if (!payload || typeof payload !== 'object' || Array.isArray(payload)) {
      return { message }
    }

    const payloadObject = payload as Record<string, unknown>
    const structuredMessage = payloadObject.message
    const statusCode = payloadObject.statusCode ?? payloadObject.status_code
    const errorCode = payloadObject.code ?? payloadObject.error_code

    if (typeof structuredMessage === 'string') {
      message = structuredMessage
    }

    return {
      message,
      errorDetails: {
        message,
        statusCode: typeof statusCode === 'number' ? statusCode : undefined,
        errorCode: typeof errorCode === 'string' ? errorCode : undefined,
      },
    }
  } catch {
    return { message }
  }
}

/**
 * Records a per-file error part on the corresponding download metadata entry.
 */
function assignDownloadErrorPart(metadata: DownloadMetadata, data: Uint8Array, contentType?: string): void {
  const { message, errorDetails } = parseDownloadErrorPart(data, contentType)
  metadata.error = message
  metadata.errorDetails = errorDetails
}

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
 * Side-channel parser that runs alongside busboy. Busboy 1.x only exposes
 * `{ filename, encoding, mimeType }` for each file part and discards the
 * remaining headers, so it cannot surface `Content-Length` for progress
 * reporting. This observer scans the raw byte stream — fed in lockstep via a
 * `Transform` tap — and queues each part's `Content-Length`. The `'file'`
 * event handler pops one value per emitted part, keeping the queue aligned
 * with busboy's event order.
 */
class MultipartHeaderObserver {
  private readonly boundary: Buffer
  private readonly closingBoundary: Buffer
  private buffer: Buffer = Buffer.alloc(0)
  private state: 'preamble' | 'headers' | 'body' | 'done' = 'preamble'
  private readonly queue: (number | undefined)[] = []

  constructor(boundary: string) {
    this.boundary = Buffer.from(`--${boundary}`)
    this.closingBoundary = Buffer.concat([Buffer.from('\r\n'), this.boundary])
  }

  observe(chunk: Buffer): void {
    if (this.state === 'done') return
    this.buffer = Buffer.concat([this.buffer, chunk])
    while (this.advance()) {
      /* drive the state machine until it stalls awaiting more bytes */
    }
  }

  /** Returns and removes the next part's Content-Length value (undefined when header was absent). */
  next(): number | undefined {
    return this.queue.shift()
  }

  private advance(): boolean {
    switch (this.state) {
      case 'preamble':
        return this.consumeBoundary(0)
      case 'headers':
        return this.consumeHeaders()
      case 'body':
        return this.consumeBodyToBoundary()
      default:
        return false
    }
  }

  private consumeHeaders(): boolean {
    const sepIdx = this.buffer.indexOf(HEADER_SEPARATOR)
    if (sepIdx < 0) {
      this.retainTail(HEADER_SEPARATOR.length - 1)
      return false
    }
    this.queue.push(parseContentLength(this.buffer.subarray(0, sepIdx)))
    this.buffer = this.buffer.subarray(sepIdx + HEADER_SEPARATOR.length)
    this.state = 'body'
    return true
  }

  private consumeBodyToBoundary(): boolean {
    const idx = this.buffer.indexOf(this.closingBoundary)
    if (idx < 0) {
      this.retainTail(this.closingBoundary.length - 1)
      return false
    }
    // Skip the leading CRLF; the rest is a regular boundary line.
    return this.consumeBoundary(idx + 2)
  }

  // Given a buffer position at-or-after "--boundary", advance past the boundary
  // and its trailing CRLF (next part follows) or "--" (multipart terminator).
  // Returns true to drive the outer loop, false when more bytes are needed.
  private consumeBoundary(searchFrom: number): boolean {
    const idx = this.buffer.indexOf(this.boundary, searchFrom)
    if (idx < 0) {
      this.retainTail(this.boundary.length - 1)
      return false
    }
    const after = idx + this.boundary.length
    if (this.buffer.length < after + 2) {
      this.buffer = this.buffer.subarray(idx)
      return false
    }
    if (this.buffer[after] === 0x2d /* '-' */ && this.buffer[after + 1] === 0x2d /* '-' */) {
      this.state = 'done'
      return false
    }
    if (this.buffer[after] === 0x0d /* CR */ && this.buffer[after + 1] === 0x0a /* LF */) {
      this.buffer = this.buffer.subarray(after + 2)
      this.state = 'headers'
      return true
    }
    // The matched substring was not a real boundary delimiter — skip it.
    return this.consumeBoundary(after)
  }

  private retainTail(keep: number): void {
    if (keep > 0 && this.buffer.length > keep) {
      this.buffer = this.buffer.subarray(this.buffer.length - keep)
    }
  }
}

const HEADER_SEPARATOR = Buffer.from('\r\n\r\n')

function parseContentLength(headersBlock: Buffer): number | undefined {
  const text = headersBlock.toString('latin1')
  for (const line of text.split('\r\n')) {
    const colon = line.indexOf(':')
    if (colon <= 0) continue
    if (line.slice(0, colon).trim().toLowerCase() !== 'content-length') continue
    const value = Number.parseInt(line.slice(colon + 1).trim(), 10)
    return Number.isFinite(value) && value >= 0 ? value : undefined
  }
  return undefined
}

/**
 * Processes multipart response using busboy (Node.js path).
 *
 * Once the file stream has been handed off via `onFileStream`, errors from the
 * busboy stream are the consumer's concern — they arrive via the file stream's
 * own 'error' event. The inner Promise resolves cleanly in that case to avoid
 * surfacing late teardown errors (busboy's `_final`, premature pipe close)
 * after the caller has already started consuming the file.
 */
export async function processDownloadFilesResponseWithBusboy(
  stream: any,
  headers: Record<string, string>,
  metadataMap: Map<string, DownloadMetadata>,
  onFileStream?: (source: string, fileStream: any, totalBytes?: number) => void,
): Promise<void> {
  const errPrefix = '"downloadFiles" is not supported: '
  const busboy = dynamicRequire('busboy', errPrefix)
  const Buffer = (dynamicRequire('buffer', errPrefix) as any).Buffer
  const fileTasks: Promise<void>[] = []

  const boundary = extractBoundary(getHeader(headers, 'content-type') || '')
  const observer = boundary ? new MultipartHeaderObserver(boundary) : null

  await new Promise<void>((resolve, reject) => {
    const bb = busboy({
      headers,
      preservePath: true,
    })

    let consumerHandedOff = false

    bb.on('file', (fieldName: string, fileStream: any, fileInfo: { filename?: string; mimeType?: string }) => {
      // Pop one queued Content-Length for every emitted part to keep the
      // observer's queue aligned with busboy's event order.
      const totalBytes = observer?.next()

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
        // Collect per-file error metadata.
        const chunks: Buffer[] = []
        fileStream.on('data', (chunk: Buffer) => chunks.push(chunk))
        fileStream.on('end', () => {
          assignDownloadErrorPart(metadata, Buffer.concat(chunks), fileInfo?.mimeType)
        })
        fileStream.on('error', (err: any) => {
          metadata.error = `Stream error: ${err.message}`
        })
      } else if (fieldName === 'file') {
        if (onFileStream) {
          consumerHandedOff = true
          fileStream.on('error', () => {
            return
          })
          onFileStream(source, fileStream, totalBytes)
        } else if (metadata.destination) {
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
      if (consumerHandedOff) {
        resolve()
        return
      }
      abortStream(stream)
      reject(err)
    })

    bb.on('finish', resolve)

    // Feed stream into busboy
    feedStreamToBusboy(stream, bb, observer).catch((err) => bb.destroy(err as Error))
  })

  await Promise.all(fileTasks)
}

/**
 * Feeds various stream types into busboy. When an observer is supplied, every
 * chunk is also handed to the observer before being forwarded to busboy so
 * per-part Content-Length can be queued in lockstep with busboy events.
 */
async function feedStreamToBusboy(stream: any, bb: any, observer: MultipartHeaderObserver | null): Promise<void> {
  const tap = (chunk: Buffer): Buffer => {
    if (observer) observer.observe(chunk)
    return chunk
  }

  // Node.js stream (piping). pipe() does NOT propagate errors from source to
  // downstream, so we manually forward them so that mid-stream cancellations
  // (e.g. AbortSignal firing after the response has started streaming) reach
  // busboy and cause the outer Promise to reject correctly.
  //
  // Errors are normalized at the forwarding site so that axios "CanceledError"
  // is already wrapped as DaytonaError by the time busboy's internal cleanup
  // destroys the file stream — ensuring the stream returned to the caller
  // always emits DaytonaError, not a raw library error.
  if (typeof stream?.pipe === 'function') {
    if (observer) {
      const { Transform } = require('stream') as typeof import('stream')
      const tapStream = new Transform({
        transform(chunk: any, _encoding: BufferEncoding, callback: (err?: Error | null, data?: any) => void) {
          const buf = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk)
          observer.observe(buf)
          callback(null, buf)
        },
      })
      stream.on('error', (err: Error) => {
        if (!tapStream.destroyed) tapStream.destroy(normalizeDownloadStreamError(err))
      })
      tapStream.on('error', (err: Error) => {
        if (!bb.destroyed) bb.destroy(err)
      })
      stream.pipe(tapStream).pipe(bb)
    } else {
      stream.on('error', (err: Error) => {
        if (!bb.destroyed) bb.destroy(normalizeDownloadStreamError(err))
      })
      stream.pipe(bb)
    }
    return
  }

  // Direct buffer-like data
  if (typeof stream === 'string' || stream instanceof ArrayBuffer || ArrayBuffer.isView(stream)) {
    const data = toUint8Array(stream)
    const buf = tap(Buffer.from(data))
    bb.write(buf)
    bb.end()
    return
  }

  // WHATWG ReadableStream
  if (typeof stream?.getReader === 'function') {
    const reader = stream.getReader()
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      bb.write(tap(Buffer.from(value)))
    }
    bb.end()
    return
  }

  // AsyncIterable
  if (stream?.[Symbol.asyncIterator]) {
    for await (const chunk of stream) {
      const buffer = Buffer.isBuffer(chunk) ? chunk : Buffer.from(toUint8Array(chunk))
      bb.write(tap(buffer))
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
      const formDataParts = await parseMultipartWithFormData(bodyBytes, contentType)

      for (const part of formDataParts) {
        const metadata = metadataMap.get(part.filename)
        if (!metadata) {
          continue
        }

        if (part.fieldName === 'error') {
          assignDownloadErrorPart(metadata, part.data, part.contentType)
        } else {
          metadata.result = toBuffer(part.data)
        }
      }

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
      assignDownloadErrorPart(metadata, part.data, part.headers['content-type'])
    } else if (part.name === 'file') {
      metadata.result = toBuffer(part.data)
    }
  }

  return
}

/**
 * Normalizes errors that arrive from the HTTP transport layer to ensure
 * callers always see a consistent error type. Axios emits "CanceledError"
 * when an AbortSignal fires; we normalize it here (at the source stream's
 * error boundary) so that DaytonaError propagates through all downstream
 * pipe stages rather than the raw library error.
 */
function normalizeDownloadStreamError(err: Error): Error {
  const e = err as { code?: string; name?: string }
  if (e.code === 'ERR_CANCELED' || e.name === 'CanceledError' || e.name === 'AbortError') {
    return new DaytonaError('Download cancelled')
  }
  return err
}

/** Construct the cancellation error thrown when an upload is aborted. */
export function createAbortError(remotePath: string): DaytonaError {
  return new DaytonaError(`Upload cancelled: ${remotePath}`)
}

/**
 * Coerces every accepted upload source shape into a Node ``Readable`` so the
 * downstream multipart writer has a uniform input type. Web ``ReadableStream``
 * is bridged via ``Readable.fromWeb``; in-memory bytes via ``Readable.from``;
 * local paths via ``fs.createReadStream``; existing ``Readable`` is passed
 * through unchanged.
 */
export async function coerceUploadSource(source: UploadSource): Promise<Readable> {
  const errPrefix = 'Uploading files is not supported: '
  const stream = await dynamicImport('stream', errPrefix)
  if (Buffer.isBuffer(source) || source instanceof Uint8Array) {
    return stream.Readable.from(Buffer.from(source))
  }
  if (typeof source === 'string') {
    const fs = await dynamicImport('fs', 'Uploading file from local file system is not supported: ')
    return fs.createReadStream(source)
  }
  if (source instanceof stream.Readable) return source as Readable
  if (typeof (source as ReadableStream).getReader === 'function') {
    return stream.Readable.fromWeb(source as any)
  }
  throw new DaytonaError(
    `Unsupported upload source: ${(source as { constructor?: { name?: string } }).constructor?.name ?? typeof source}`,
  )
}

/**
 * Wraps an upload source in a pass-through that counts bytes and invokes
 * ``onProgress`` per chunk. When ``onProgress`` is omitted the source is
 * returned unchanged so there is zero overhead in the no-progress path.
 */
export function wrapWithUploadProgress(
  source: Readable,
  onProgress: ((progress: UploadProgress) => void) | undefined,
  signal?: AbortSignal,
): Readable {
  if (!onProgress && !signal) return source

  const errPrefix = 'Uploading files is not supported: '
  const { Transform, pipeline } = dynamicRequire('stream', errPrefix) as typeof import('stream')

  let bytesSent = 0
  const tracker = new Transform({
    transform(chunk: Buffer | string, _encoding: BufferEncoding, callback: (err?: Error | null, data?: any) => void) {
      const buf = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk)
      bytesSent += buf.length
      onProgress?.({ bytesSent })
      callback(null, buf)
    },
  })

  pipeline(source, tracker, (err) => {
    if (!err) return
    if (!source.destroyed) source.destroy(err)
    if (!tracker.destroyed) tracker.destroy(err)
  })

  if (signal) {
    const onAbort = () => {
      const abortError = new Error('aborted')
      if (!source.destroyed) source.destroy(abortError)
      if (!tracker.destroyed) tracker.destroy(abortError)
    }

    if (signal.aborted) {
      queueMicrotask(onAbort)
    } else {
      signal.addEventListener('abort', onAbort, { once: true })
      tracker.on('close', () => signal.removeEventListener('abort', onAbort))
    }
  }

  return tracker
}
