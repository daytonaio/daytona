/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

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
