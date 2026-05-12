/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { motion } from 'framer-motion'
import React from 'react'

interface QuotaLineProps {
  current: number
  total: number
  className?: string
}

const transition = {
  type: 'spring',
  stiffness: 60,
  damping: 15,
  mass: 1,
} as const

const QuotaLine: React.FC<QuotaLineProps> = ({ current, total, className }) => {
  const percentage = total > 0 ? Math.min(Math.max((current / total) * 100, 0), 100) : 0
  const fillGradientWidth = percentage > 0 ? `${10000 / percentage}%` : '100%'

  return (
    <div className={cn('relative h-2 w-full overflow-clip rounded-full bg-muted', className)}>
      <motion.div
        className="h-full overflow-hidden rounded-full"
        initial={{ width: 0 }}
        animate={{ width: `${percentage}%` }}
        transition={transition}
      >
        <div
          className="h-full bg-[linear-gradient(to_right,hsl(var(--success-foreground))_0%,hsl(var(--success-foreground))_36%,hsl(var(--warning-foreground))_64%,hsl(var(--destructive-foreground))_100%)]"
          style={{ width: fillGradientWidth }}
        />
      </motion.div>
    </div>
  )
}

export default QuotaLine
