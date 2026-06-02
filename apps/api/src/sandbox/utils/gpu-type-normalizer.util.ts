/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { GpuType } from '../enums/gpu-type.enum'

const GPU_TYPE_PATTERNS: ReadonlyArray<readonly [RegExp, GpuType]> = [
  [/h100/i, GpuType.H100],
  [/rtx\s*pro\s*6000/i, GpuType.RTX_PRO_6000],
]

export function normalizeGpuType(raw: string | null | undefined): GpuType | null {
  if (!raw) return null
  for (const [pattern, type] of GPU_TYPE_PATTERNS) {
    if (pattern.test(raw)) return type
  }
  return null
}
