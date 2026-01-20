/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import isEqual from 'fast-deep-equal'
import { useRef } from 'react'

export function useDeepCompareMemo<T>(value: T) {
  const ref = useRef<T>(value)

  if (!isEqual(value, ref.current)) {
    ref.current = value
  }

  return ref.current
}
