/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: Apache-2.0
 */

import { createHmac } from 'crypto'
import { DaytonaError } from '../errors/DaytonaError'

const SIGNATURE_V1_PREFIX = 'v1_'
const DEFAULT_TTL_SECONDS = 3600

export function computeFileUrlSignature(signingKey: string, method: string, path: string, expires: number): string {
  const canonical = `v1:files:${method}:${path}:${expires}`
  const digest = createHmac('sha256', signingKey).update(canonical).digest('base64url')
  return `${SIGNATURE_V1_PREFIX}${digest}`
}

export function resolveExpires(ttlSeconds: number | undefined): number {
  if (ttlSeconds === undefined) {
    return Math.floor(Date.now() / 1000) + DEFAULT_TTL_SECONDS
  }
  if (!Number.isFinite(ttlSeconds)) {
    throw new DaytonaError('ttlSeconds must be a finite number')
  }
  if (ttlSeconds <= 0) {
    return 0
  }
  return Math.floor(Date.now() / 1000) + Math.floor(ttlSeconds)
}

export function buildSignedFileUrl(
  toolboxProxyUrl: string,
  sandboxId: string,
  operationPath: string,
  method: string,
  filePath: string,
  signingKey: string,
  ttlSeconds?: number,
): string {
  if (!signingKey) {
    throw new DaytonaError(
      'Sandbox signing key is not available. Call refreshData() or fetch the sandbox by ID to load it.',
    )
  }

  const expires = resolveExpires(ttlSeconds)
  const signature = computeFileUrlSignature(signingKey, method, filePath, expires)
  const query = new URLSearchParams({
    path: filePath,
    expires: String(expires),
    signature,
  })

  return `${toolboxProxyUrl}/${sandboxId}${operationPath}?${query.toString()}`
}
