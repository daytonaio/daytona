/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { GpuType } from '../enums/gpu-type.enum'
import { BadRequestError } from '../../exceptions/bad-request.exception'

/**
 * Reconciles a request's GPU type preferences against the region's
 * `allowedGpuTypes` allowlist. Call once at the start of every create flow,
 * before invoking the scheduler.
 *
 * Region allowlist semantics:
 *  - `null` (or `undefined`): no restriction.
 *  - `[]`: empty allowlist — all GPU types are blocked in this region.
 *  - non-empty array: only the listed types are permitted.
 *
 * @returns Effective preference list to pass to the scheduler, or `undefined`
 *   when no GPU type filter should be applied.
 * @throws {BadRequestError} When the region blocks all GPU types, or none of
 *   the requested preferences intersect with the region's allowlist.
 */
export function resolveGpuTypePreferences(
  gpu: number,
  gpuTypePreferences: GpuType[] | undefined,
  allowedGpuTypes: GpuType[] | null | undefined,
): GpuType[] | undefined {
  if (gpu <= 0) return undefined

  if (allowedGpuTypes == null) {
    return gpuTypePreferences && gpuTypePreferences.length > 0 ? gpuTypePreferences : undefined
  }

  if (allowedGpuTypes.length === 0) {
    throw new BadRequestError('No GPU types are allowed in this region.')
  }

  if (!gpuTypePreferences || gpuTypePreferences.length === 0) {
    return allowedGpuTypes
  }

  const permitted = gpuTypePreferences.filter((t) => allowedGpuTypes.includes(t))
  if (permitted.length === 0) {
    throw new BadRequestError(
      `Requested GPU type(s) ${gpuTypePreferences.join(', ')} not permitted in this region. Allowed: ${allowedGpuTypes.join(', ')}.`,
    )
  }
  return permitted
}
