/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */
import WebSocket from 'isomorphic-ws'
import { STDOUT_PREFIX_BYTES, STDERR_PREFIX_BYTES, MAX_PREFIX_LEN } from '../Process'

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
    const buf: number[] = [] // Buffer to accumulate incoming chunks
    let currentDataType: 'stdout' | 'stderr' | null = null // Track current stream type

    // Helper function to emit payload data
    const emit = (payload: Uint8Array) => {
      if (payload.length === 0) return
      const text = textDecoder.decode(payload)
      if (currentDataType === 'stdout') {
        onStdout(text)
      } else if (currentDataType === 'stderr') {
        onStderr(text)
      }
      // If currentDataType is null, drop unlabeled bytes (shouldn't happen with proper labeling)
    }

    // Helper function to find a subarray within a larger array
    const findSubarray = (haystack: Uint8Array, needle: Uint8Array): number => {
      if (needle.length === 0) return 0
      if (haystack.length < needle.length) return -1

      for (let i = 0; i <= haystack.length - needle.length; i++) {
        let found = true
        for (let j = 0; j < needle.length; j++) {
          if (haystack[i + j] !== needle[j]) {
            found = false
            break
          }
        }
        if (found) return i
      }
      return -1
    }

    // Event handler for incoming messages (Node: Buffer/ArrayBuffer/String; Browser: event.data etc.)
    const handleMessage = (event: MessageEvent | Buffer | ArrayBuffer | string | any) => {
      // Normalize event/data between Node (ws) and browser WebSocket
      const data = event && event instanceof Object && 'data' in event ? event.data : event
      try {
        // Prepare a Uint8Array for the message data
        let bytes: Uint8Array
        if (typeof data === 'string') {
          // Convert string to bytes for consistent byte-based demuxing
          bytes = new TextEncoder().encode(data)
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
                processChunk(new Uint8Array(buf))
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

        // Process the chunk
        processChunk(bytes)
      } catch (err) {
        // On any synchronous error in processing, clean up and reject.
        cleanup()
        try {
          ws.close()
        } catch {
          /* ignore if already closed */
        }
        reject(err)
      }
    }

    // Process a chunk of data with buffering and safe region handling
    const processChunk = (chunk: Uint8Array) => {
      if (chunk.length === 0) return

      // Add chunk to buffer
      buf.push(...chunk)

      // Process as much as we can, preserving only bytes that could be part of a prefix
      while (true) {
        const bufArray = new Uint8Array(buf)

        // Calculate how many bytes we can safely process
        // We need to keep bytes that could potentially be the start of a prefix marker
        let safeLen = buf.length

        // Check if the last few bytes could be part of a prefix marker
        if (buf.length >= MAX_PREFIX_LEN) {
          // Check if the last byte could be part of a prefix (must be 0x01 or 0x02)
          const lastByte = buf[buf.length - 1]
          if (lastByte !== 0x01 && lastByte !== 0x02) {
            // Last byte can't be part of any prefix, safe to process everything
            safeLen = buf.length
          } else if (buf.length >= MAX_PREFIX_LEN + 1) {
            // Check second-to-last byte if buffer is long enough
            const secondLastByte = buf[buf.length - 2]
            if (secondLastByte !== 0x01 && secondLastByte !== 0x02) {
              // Second-to-last byte can't be part of any prefix, safe to process all but last byte
              safeLen = buf.length - 1
            } else {
              // Both last bytes could be part of prefix, keep MAX_PREFIX_LEN - 1 bytes
              safeLen = buf.length - (MAX_PREFIX_LEN - 1)
            }
          } else {
            // Buffer is exactly MAX_PREFIX_LEN, keep MAX_PREFIX_LEN - 1 bytes
            safeLen = buf.length - (MAX_PREFIX_LEN - 1)
          }
        } else {
          // Buffer shorter than MAX_PREFIX_LEN, keep MAX_PREFIX_LEN - 1 bytes
          safeLen = buf.length - (MAX_PREFIX_LEN - 1)
        }

        if (safeLen <= 0) {
          break
        }

        // Find earliest next marker within the safe region
        const safeRegion = bufArray.subarray(0, safeLen)
        const stdoutIndex = findSubarray(safeRegion, STDOUT_PREFIX_BYTES)
        const stderrIndex = findSubarray(safeRegion, STDERR_PREFIX_BYTES)

        let nextIdx = -1
        let nextKind: 'stdout' | 'stderr' | null = null
        let nextLen = 0

        if (stdoutIndex !== -1 && (stderrIndex === -1 || stdoutIndex < stderrIndex)) {
          nextIdx = stdoutIndex
          nextKind = 'stdout'
          nextLen = STDOUT_PREFIX_BYTES.length
        } else if (stderrIndex !== -1) {
          nextIdx = stderrIndex
          nextKind = 'stderr'
          nextLen = STDERR_PREFIX_BYTES.length
        }

        if (nextIdx === -1) {
          // No full marker in safe region: emit everything we safely can as payload
          const toEmit = bufArray.subarray(0, safeLen)
          emit(toEmit)
          buf.splice(0, safeLen)
          break // wait for more data to resolve any partial marker at the end
        }

        // We found a marker. Emit preceding bytes (if any) under the current stream.
        if (nextIdx > 0) {
          const toEmit = bufArray.subarray(0, nextIdx)
          emit(toEmit)
        }

        // Advance past the marker and switch current stream
        buf.splice(0, nextIdx + nextLen)
        currentDataType = nextKind
      }
    }

    // Event handler for errors
    const handleError = (error: any) => {
      // Convert Event or plain error to Error instance for consistency
      const err = error && error instanceof Event ? new Error('WebSocket error') : error
      cleanup()
      try {
        ws.close()
      } catch {
        /* ignore if already closed */
      }
      reject(err)
    }

    // Event handler for socket closure
    const handleClose = () => {
      // Flush any remaining buffered payload on clean close
      if (buf.length > 0 && currentDataType) {
        const remainingBytes = new Uint8Array(buf)
        const text = textDecoder.decode(remainingBytes)
        if (currentDataType === 'stdout') {
          onStdout(text)
        } else if (currentDataType === 'stderr') {
          onStderr(text)
        }
      }
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
