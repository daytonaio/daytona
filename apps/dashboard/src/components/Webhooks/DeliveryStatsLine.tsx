/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import type { EndpointStats } from 'svix'
import { motion } from 'framer-motion'
import React from 'react'

const transition = {
  type: 'spring',
  stiffness: 60,
  damping: 15,
  mass: 1,
} as const

const SEGMENTS = [
  { key: 'success', label: 'Success', color: 'bg-green-500', dotColor: 'bg-green-500' },
  { key: 'fail', label: 'Failed', color: 'bg-red-500', dotColor: 'bg-red-500' },
  { key: 'pending', label: 'Pending', color: 'bg-muted-foreground/50', dotColor: 'bg-muted-foreground/50' },
  { key: 'sending', label: 'Sending', color: 'bg-white', dotColor: 'bg-white border border-border' },
] as const

interface DeliveryStatsLineProps {
  stats: EndpointStats
  className?: string
}

const DeliveryStatsLine: React.FC<DeliveryStatsLineProps> = ({ stats, className }) => {
  const total = stats.success + stats.fail + stats.pending + stats.sending

  if (total === 0) {
    return (
      <div className={cn('flex flex-col gap-2', className)}>
        <div className="w-full h-2 bg-muted rounded-full" />
        <Legend stats={stats} total={total} />
      </div>
    )
  }

  return (
    <div className={cn('flex flex-col gap-2', className)}>
      <div className="w-full h-2 bg-muted rounded-full overflow-clip flex">
        {SEGMENTS.map(({ key, color }) => {
          const value = stats[key]
          if (value === 0) return null
          const pct = (value / total) * 100
          return (
            <motion.div
              key={key}
              className={cn('h-full', color)}
              initial={{ width: 0 }}
              animate={{ width: `${pct}%` }}
              transition={transition}
            />
          )
        })}
      </div>
      <Legend stats={stats} total={total} />
    </div>
  )
}

function Legend({ stats, total }: { stats: EndpointStats; total: number }) {
  return (
    <div className="flex items-center gap-4 flex-wrap">
      {SEGMENTS.map(({ key, label, dotColor }) => (
        <div key={key} className="flex items-center gap-1.5 text-xs text-muted-foreground">
          <div className={cn('w-2 h-2 rounded-full', dotColor)} />
          <span>
            {label} {stats[key]}
            {total > 0 && <span className="ml-0.5">({Math.round((stats[key] / total) * 100)}%)</span>}
          </span>
        </div>
      ))}
    </div>
  )
}

export default DeliveryStatsLine
