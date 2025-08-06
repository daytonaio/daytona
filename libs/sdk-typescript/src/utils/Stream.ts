/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */
import WebSocket from 'isomorphic-ws'

export const STDOUT_PREFIX = 1
export const STDERR_PREFIX = 2

/**
 * Process a streaming response from fetch(), where getStream() returns a Fetch Response.
 *
 * @param getStream – zero-arg function that does `await fetch(...)` and returns the Response
 * @param onChunk – called with each decoded UTF-8 chunk
 * @param shouldTerminate – pollable; if true for two consecutive timeouts (or once if requireConsecutiveTermination=false), the loop breaks
 * @param chunkTimeout – milliseconds to wait for a new chunk before calling shouldTerminate()
 * @param requireConsecutiveTermination – whether you need two time-outs in a row to break
 */
export async function processStreamingResponse(
  getStream: () => Promise<Response>,
  onChunk: (chunk: string) => void,
  shouldTerminate: () => Promise<boolean>,
  chunkTimeout = 2000,
  requireConsecutiveTermination = true,
): Promise<void> {
  const res = await getStream()
  if (!res.body) throw new Error('No streaming support')
  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  const TIMEOUT = Symbol()
  let exitCheckStreak = 0

  // Only one pending read promise at a time:
  let readPromise: Promise<Uint8Array | null> | null = null

  try {
    while (true) {
      // Start a read if none in flight
      if (!readPromise) {
        readPromise = reader.read().then((r) => (r.done ? null : (r.value ?? new Uint8Array(0))))
      }

      // Race that single read against your timeout
      const timeoutPromise = new Promise<typeof TIMEOUT>((r) => setTimeout(() => r(TIMEOUT), chunkTimeout))
      const result = await Promise.race([readPromise, timeoutPromise])

      if (result === TIMEOUT) {
        // no data yet, but the readPromise is still pending
        const stop = await shouldTerminate()
        if (stop) {
          exitCheckStreak++
          if (!requireConsecutiveTermination || exitCheckStreak > 1) break
        } else {
          exitCheckStreak = 0
        }
        // loop again—but do NOT overwrite readPromise!
      } else {
        // readPromise has resolved
        readPromise = null
        if (result === null) {
          // stream closed
          break
        }
        // valid chunk
        onChunk(decoder.decode(result))
        exitCheckStreak = 0
      }
    }
  } finally {
    await reader.cancel()
  }
}

/**
 * Demultiplexes a WebSocket stream into separate stdout and stderr streams.
 *
 * @param socket - The WebSocket instance to demultiplex.
 * @param onStdout - Callback function for stdout messages.
 * @param onStderr - Callback function for stderr messages.
 */
export function stdDemuxStream(
  ws: WebSocket,
  onStdout: (data: string) => void,
  onStderr: (data: string) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    // If running in a browser or any WebSocket supporting binaryType, use ArrayBuffer for binary data
    if ('binaryType' in ws) {
      ws.binaryType = 'arraybuffer' // ensure binary frames yield ArrayBuffer, not Blob
    }

    const textDecoder = new TextDecoder() // for decoding UTF-8 bytes to string

    // Event handler for incoming messages (Node: Buffer/ArrayBuffer/String; Browser: event.data etc.)
    const handleMessage = (event: MessageEvent | Buffer | ArrayBuffer | string | any) => {
      // Normalize event/data between Node (ws) and browser WebSocket
      const data = event && event instanceof Object && 'data' in event ? event.data : event
      try {
        // Prepare a Uint8Array for the message data (so we can inspect the first byte and decode the rest)
        let bytes: Uint8Array
        if (typeof data === 'string') {
          // If a text message is received (e.g., older ws or if server sent text), first char is the prefix byte
          const prefixCode = data.charCodeAt(0)
          const contentText = data.substring(1)
          // Deliver to appropriate stream based on prefix
          if (prefixCode === STDOUT_PREFIX) {
            onStdout(contentText)
          } else if (prefixCode === STDERR_PREFIX) {
            onStderr(contentText)
          }
          return // done handling this message
        } else if (data instanceof ArrayBuffer) {
          bytes = new Uint8Array(data)
        } else if (ArrayBuffer.isView(data)) {
          // Covers Node.js Buffer (Uint8Array subclass) and other TypedArrays
          bytes = new Uint8Array(data.buffer, data.byteOffset, data.byteLength)
        } else if (data instanceof Blob) {
          // Browser binary frames might be Blob if binaryType wasn't set in time. Convert to ArrayBuffer asynchronously.
          data.arrayBuffer().then(
            (buf: ArrayBuffer) => {
              try {
                processBytes(new Uint8Array(buf))
              } catch (err) {
                handleError(err)
              }
            },
            (err: any) => {
              handleError(err)
            },
          )
          return // will continue asynchronously once blob is read
        } else {
          throw new Error(`Unsupported message data type: ${Object.prototype.toString.call(data)}`)
        }

        // We have a Uint8Array 'bytes' representing the full message.
        processBytes(bytes)
      } catch (err) {
        // On any synchronous error in processing, clean up and reject.
        cleanup()
        try {
          ws.close()
        } catch (_) {
          /* ignore if already closed */
        }
        reject(err)
      }
    }

    // Process a Uint8Array message: demux by prefix and decode the content
    const processBytes = (bytes: Uint8Array) => {
      if (bytes.length < 1) return
      const channel = bytes[0] // prefix byte: 1 = STDOUT, 2 = STDERR
      const contentBytes = bytes.subarray(1)
      // Decode remaining bytes to UTF-8 string (TextDecoder handles multi-byte chars properly)
      const text = textDecoder.decode(contentBytes)
      if (channel === 1) {
        onStdout(text)
      } else if (channel === 2) {
        onStderr(text)
      } else {
        // If other channels exist (e.g., 0 for stdin echo, 3 for server error messages), ignore or handle as needed.
        // Here we ignore unexpected channels, but you could add handling for channel 3 (error messages) if required.
      }
    }

    // Event handler for errors
    const handleError = (error: any) => {
      // Convert Event or plain error to Error instance for consistency
      const err = error && error instanceof Event ? new Error('WebSocket error') : error
      cleanup()
      try {
        ws.close()
      } catch (_) {
        /* ignore if already closed */
      }
      reject(err)
    }

    // Event handler for socket closure
    const handleClose = () => {
      cleanup()
      resolve()
    }

    // Cleanup function to remove all listeners to avoid memory leaks
    const cleanup = () => {
      if (ws.removeEventListener) {
        // Browser (EventTarget) style cleanup
        ws.removeEventListener('message', handleMessage as any)
        ws.removeEventListener('error', handleError as any)
        ws.removeEventListener('close', handleClose as any)
      }
      if (ws.off) {
        // Node.js ws (EventEmitter) style cleanup (supported in Node 14+)
        ws.off('message', handleMessage)
        ws.off('error', handleError)
        ws.off('close', handleClose)
      } else if ((ws as any).removeListener) {
        // Node.js ws fallback for older Node versions
        ;(ws as any).removeListener('message', handleMessage)
        ;(ws as any).removeListener('error', handleError)
        ;(ws as any).removeListener('close', handleClose)
      }
    }

    // Attach event listeners in a way compatible with both Node (EventEmitter) and browser (EventTarget):
    if (ws.addEventListener) {
      // Browser or WebSocket implementation with EventTarget interface
      ws.addEventListener('message', handleMessage as any)
      ws.addEventListener('error', handleError as any)
      ws.addEventListener('close', handleClose as any)
    } else if ((ws as any).on) {
      // Node.js ws library (EventEmitter) interface
      ws.on('message', handleMessage) // ws@8+ yields Buffer for text frames, which we handle via TextDecoder
      ws.on('error', handleError)
      ws.on('close', handleClose)
    } else {
      // Unknown WebSocket interface - should not happen with isomorphic-ws
      throw new Error('Unsupported WebSocket implementation')
    }
  })
}
