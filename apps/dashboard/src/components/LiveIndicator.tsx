/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useQueryCountdown } from '@/hooks/useQueryCountdown'
import { cn } from '@/lib/utils'
import { Tooltip } from './Tooltip'

export function LiveIndicator({
  isUpdating,
  intervalMs,
  lastUpdatedAt,
}: {
  isUpdating: boolean
  intervalMs: number
  lastUpdatedAt: number
}) {
  const refreshingIn = useQueryCountdown(lastUpdatedAt, intervalMs)

  return (
    <Tooltip
      content={
        <div className="relative flex flex-col items-center">
          <div className="text-xs text-muted-foreground">Data is refreshed every {intervalMs / 1000} seconds.</div>
          <span className={cn('text-xs text-muted-foreground/70')}>
            Refreshing{' '}
            {isUpdating ? (
              ''
            ) : (
              <>
                in <span className="min-w-[2ch] inline-block tabular-nums text-center">{refreshingIn}</span>s
              </>
            )}
            ...
          </span>
        </div>
      }
      label={
        <div className="flex items-center gap-2">
          <div
            className={cn('w-2 h-2 bg-green-500 rounded-full transition-all', {
              'opacity-70': isUpdating,
            })}
          >
            <div
              className={cn('w-full h-full bg-green-500 rounded-full', {
                'animate-ping': !isUpdating,
              })}
            />
          </div>
          <span
            className={cn('text-xs text-muted-foreground transition-all', {
              'opacity-70': isUpdating,
            })}
          >
            Live
          </span>
        </div>
      }
    />
  )
}
