/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useEffect, useState } from 'react'

export function useQueryCountdown(
  dataUpdatedAt: number,
  intervalMs: number,
  {
    updateInterval,
  }: {
    updateInterval?: number
  } = {},
) {
  const [seconds, setSeconds] = useState(intervalMs / 1000)

  useEffect(() => {
    if (!dataUpdatedAt) return

    const timer = setInterval(() => {
      const timePassed = Date.now() - dataUpdatedAt
      const remaining = Math.max(0, intervalMs - timePassed)
      setSeconds(Math.ceil(remaining / 1000))
    }, updateInterval || 1000)

    return () => clearInterval(timer)
  }, [dataUpdatedAt, intervalMs, updateInterval])

  return seconds
}
