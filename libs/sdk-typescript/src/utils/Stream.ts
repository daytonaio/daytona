/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { DaytonaError } from "../errors/DaytonaError"
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

export function stdDemuxStream(
  socket: WebSocket,
  onStdout: (log: string) => void,
  onStderr: (log: string) => void
): Promise<void> {
  if ('binaryType' in socket) {
    (socket as any).binaryType = 'arraybuffer';
  }

  const decoder = new TextDecoder('utf-8', { fatal: false });
  const encoder = new TextEncoder();

  return new Promise((resolve, reject) => {
    let cleanedUp = false;

    function cleanup() {
      if (cleanedUp) return;
      cleanedUp = true;
      if (typeof (socket as any).off === 'function') {
        (socket as any).off('message', unifiedHandler);
        (socket as any).off('close',   onClose);
        (socket as any).off('error',   onError);
      } else {
        socket.removeEventListener('message', unifiedHandler as any);
        socket.removeEventListener('close',   onClose   as any);
        socket.removeEventListener('error',   onError   as any);
      }
    }

    function process(buf: Uint8Array) {
      if (!buf.length) return;
      const prefix = buf[0];
      const msg    = decoder.decode(buf.subarray(1));
      if (prefix === STDOUT_PREFIX)       return onStdout(msg);
      if (prefix === STDERR_PREFIX)       return onStderr(msg);
      cleanup();
      reject(new Error(`Unknown data prefix ${prefix}`));
    }

    function unifiedHandler(raw: any) {
      const data = raw instanceof MessageEvent ? raw.data : raw;
      if (typeof data === 'string') {
        process(encoder.encode(data));
      }
      else if (data instanceof ArrayBuffer) {
        process(new Uint8Array(data));
      }
      else if (Array.isArray(data)) {
        process(new Uint8Array(Buffer.concat(data)));
      }
      else if (ArrayBuffer.isView(data)) {
        const v = data as ArrayBufferView;
        process(new Uint8Array(v.buffer, v.byteOffset, v.byteLength));
      }
      else if (data instanceof Blob) {
        data.arrayBuffer()
          .then(ab => process(new Uint8Array(ab)))
          .catch(err => { cleanup(); reject(err); });
      }
      else {
        cleanup();
        reject(new Error(`Unknown data type: ${typeof data}`));
      }
    }

    const onClose = () => { cleanup(); resolve(); };
    const onError = (err: any) => {
      cleanup();
      if (err instanceof Event) {
        reject(new Error(`WebSocket error event: ${err.type}`));
      } else {
        reject(err);
      }
    };

    // Register for both Node-style and DOM-style
    if (typeof (socket as any).on === 'function') {
      (socket as any).on('message', unifiedHandler);
      (socket as any).on('close',   onClose);
      (socket as any).on('error',   onError);
    } else {
      socket.addEventListener('message', unifiedHandler as any);
      socket.addEventListener('close',   onClose   as any);
      socket.addEventListener('error',   onError   as any);
    }
  });
}
