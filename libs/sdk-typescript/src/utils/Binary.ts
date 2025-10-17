/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Buffer } from 'buffer'
import { DaytonaError } from '../errors/DaytonaError'

/**
 * Converts various data types to Uint8Array
 */
export function toUint8Array(data: string | ArrayBuffer | ArrayBufferView): Uint8Array {
  if (typeof data === 'string') {
    return new TextEncoder().encode(data)
  }
  if (data instanceof ArrayBuffer) {
    return new Uint8Array(data)
  }
  if (ArrayBuffer.isView(data)) {
    return new Uint8Array(data.buffer, data.byteOffset, data.byteLength)
  }
  throw new DaytonaError('Unsupported data type for byte conversion.')
}

/**
 * Concatenates multiple Uint8Array chunks into a single Uint8Array
 */
export function concatUint8Arrays(parts: Uint8Array[]): Uint8Array {
  const size = parts.reduce((sum, part) => sum + part.byteLength, 0)
  const result = new Uint8Array(size)
  let offset = 0
  for (const part of parts) {
    result.set(part, offset)
    offset += part.byteLength
  }
  return result
}

/**
 * Converts Uint8Array to Buffer (uses polyfill in non-Node environments)
 */
export function toBuffer(data: Uint8Array): Buffer {
  return Buffer.from(data)
}

/**
 * Decodes Uint8Array to UTF-8 string
 */
export function utf8Decode(data: Uint8Array): string {
  return new TextDecoder('utf-8').decode(data)
}

/**
 * Finds all occurrences of a pattern in a byte buffer
 */
export function findAllBytes(buffer: Uint8Array, pattern: Uint8Array): number[] {
  const results: number[] = []
  let i = 0
  while (i <= buffer.length - pattern.length) {
    let match = true
    for (let j = 0; j < pattern.length; j++) {
      if (buffer[i + j] !== pattern[j]) {
        match = false
        break
      }
    }
    if (match) {
      results.push(i)
      i += pattern.length
    } else {
      i++
    }
  }
  return results
}

/**
 * Finds the first occurrence of a pattern in a byte buffer within a range
 */
export function findBytesInRange(buffer: Uint8Array, start: number, end: number, pattern: Uint8Array): number {
  let i = start
  while (i <= end - pattern.length) {
    let match = true
    for (let j = 0; j < pattern.length; j++) {
      if (buffer[i + j] !== pattern[j]) {
        match = false
        break
      }
    }
    if (match) return i
    i++
  }
  return -1
}

/**
 * Checks if a sequence starts at a given position in a byte buffer
 * Returns the position after the sequence if found, -1 otherwise
 */
export function indexAfterSequence(buffer: Uint8Array, start: number, sequence: Uint8Array): number {
  for (let j = 0; j < sequence.length; j++) {
    if (buffer[start + j] !== sequence[j]) return -1
  }
  return start + sequence.length
}

/**
 * Collects all bytes from various stream types into a single Uint8Array
 */
export async function collectStreamBytes(stream: any): Promise<Uint8Array> {
  if (!stream) return new Uint8Array(0)

  // ReadableStream (WHATWG)
  if (typeof stream.getReader === 'function') {
    const reader = stream.getReader()
    const chunks: Uint8Array[] = []
    try {
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        if (value?.byteLength) {
          chunks.push(value)
        }
      }
    } finally {
      await reader.cancel()
    }
    return concatUint8Arrays(chunks)
  }

  // AsyncIterable
  if (stream?.[Symbol.asyncIterator]) {
    const chunks: Uint8Array[] = []
    for await (const chunk of stream) {
      chunks.push(toUint8Array(chunk))
    }
    return concatUint8Arrays(chunks)
  }

  // Direct data types
  if (typeof stream === 'string' || stream instanceof ArrayBuffer || ArrayBuffer.isView(stream)) {
    return toUint8Array(stream)
  }

  // Blob
  if (typeof Blob !== 'undefined' && stream instanceof Blob) {
    const arrayBuffer = await stream.arrayBuffer()
    return new Uint8Array(arrayBuffer)
  }

  // Response
  if (typeof Response !== 'undefined' && stream instanceof Response) {
    const arrayBuffer = await stream.arrayBuffer()
    return new Uint8Array(arrayBuffer)
  }

  throw new DaytonaError('Unsupported stream type for byte collection.')
}

/**
 * Checks if value is a File object (browser environment)
 */
export function isFile(value: any): boolean {
  const FileConstructor = (globalThis as any).File
  return typeof FileConstructor !== 'undefined' && value instanceof FileConstructor
}
