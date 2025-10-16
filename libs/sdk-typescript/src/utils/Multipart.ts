/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { concatUint8Arrays, findAllBytes, findBytesInRange, indexAfterSequence, utf8Decode, isFile } from './Binary'

export interface MultipartPart {
  name: string | undefined
  filename: string | undefined
  headers: Record<string, string>
  data: Uint8Array
}

/**
 * Extracts the boundary from a Content-Type header
 */
export function extractBoundary(contentType: string): string | null {
  const match = /boundary="?([^";]+)"?/i.exec(contentType || '')
  return match ? match[1] : null
}

/**
 * Extracts a parameter value from Content-Disposition header
 */
function getDispositionParam(disposition: string, paramName: 'name' | 'filename'): string | undefined {
  const match = disposition.match(new RegExp(`${paramName}\\*?=([^;]+)`, 'i'))
  if (!match) return undefined
  return match[1].replace(/^"|"$/g, '').trim()
}

/**
 * Parses multipart/form-data or multipart/mixed response body
 */
export function parseMultipart(body: Uint8Array, boundary: string): MultipartPart[] {
  const encoder = new TextEncoder()
  const dashBoundary = encoder.encode(`--${boundary}`)
  const crlf = encoder.encode('\r\n')
  const boundaryLine = concatUint8Arrays([dashBoundary, crlf])

  const boundaryPositions = findAllBytes(body, dashBoundary)
  if (boundaryPositions.length === 0) return []

  const parts: MultipartPart[] = []

  for (let i = 0; i < boundaryPositions.length; i++) {
    const start = boundaryPositions[i]

    // Headers start after "--boundary\r\n"
    const headerStart = indexAfterSequence(body, start, boundaryLine)
    if (headerStart < 0) continue

    // Part ends before next boundary
    const nextBoundary = boundaryPositions[i + 1] ?? body.length
    let partEnd = nextBoundary - 2 // Remove trailing CRLF
    if (partEnd < headerStart) partEnd = headerStart

    // Find headers/body separator (\r\n\r\n)
    const separator = findBytesInRange(body, headerStart, partEnd, encoder.encode('\r\n\r\n'))
    if (separator < 0) continue

    // Parse headers
    const headersText = utf8Decode(body.subarray(headerStart, separator))
    const headers: Record<string, string> = {}

    headersText.split(/\r\n/).forEach((line) => {
      const [key, ...valueParts] = line.split(':')
      if (valueParts.length > 0) {
        headers[key.trim().toLowerCase()] = valueParts.join(':').trim()
      }
    })

    // Extract body
    const dataStart = separator + 4
    const data = body.subarray(dataStart, partEnd)

    // Extract name and filename from Content-Disposition
    const disposition = headers['content-disposition'] || ''
    const name = getDispositionParam(disposition, 'name')
    const filename = getDispositionParam(disposition, 'filename')

    parts.push({ name, filename, headers, data })
  }

  return parts
}

/**
 * Parses multipart response using browser's native FormData API
 * This is more reliable than manual parsing when available
 */
export async function parseMultipartWithFormData(
  bodyBytes: Uint8Array,
  contentType: string,
): Promise<Map<string, { filename: string; data: Uint8Array }>> {
  const result = new Map<string, { filename: string; data: Uint8Array }>()

  // Create a Blob and parse with FormData API
  const blob = new Blob([bodyBytes.slice()], { type: contentType })
  const formData = await new Response(blob).formData()

  // Process FormData entries (forEach is more universally supported than entries())
  const filePromises: Promise<void>[] = []
  formData.forEach((value, fieldName) => {
    if (isFile(value)) {
      filePromises.push(
        (async () => {
          const file = value as File
          const arrayBuffer = await file.arrayBuffer()
          result.set(fieldName, {
            filename: file.name,
            data: new Uint8Array(arrayBuffer),
          })
        })(),
      )
    }
  })

  await Promise.all(filePromises)
  return result
}

/**
 * Extracts a header value from response headers (case-insensitive)
 */
export function getHeader(headers: any, key: string): string | undefined {
  if (!headers) return undefined
  const headerKey = Object.keys(headers).find((h) => h.toLowerCase() === key.toLowerCase())
  if (!headerKey) return undefined
  const value = headers[headerKey]
  return Array.isArray(value) ? value[0] : value
}
