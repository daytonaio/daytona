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
  const percentage = Math.min(Math.max((current / total) * 100, 0), 100)

  const greenWidth = Math.min(percentage, 60)
  const yellowWidth = Math.min(Math.max(percentage - 60, 0), 30)
  const redWidth = Math.min(Math.max(percentage - 90, 0), 10)

  return (
    <div className={cn('w-full h-2 bg-muted rounded-full overflow-clip flex relative', className)}>
      <motion.div
        className="h-full bg-green-500"
        initial={{ width: 0 }}
        animate={{ width: `${greenWidth}%` }}
        transition={transition}
      />
      <motion.div
        className="h-full bg-yellow-400"
        initial={{ width: 0 }}
        animate={{ width: `${yellowWidth}%` }}
        transition={transition}
      />
      <motion.div
        className="h-full bg-red-500"
        initial={{ width: 0 }}
        animate={{ width: `${redWidth}%` }}
        transition={transition}
      />
    </div>
  )
}

export default QuotaLine
