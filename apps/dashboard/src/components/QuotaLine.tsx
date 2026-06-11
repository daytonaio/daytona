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
  const percentage = total > 0 ? Math.min(Math.max((current / total) * 100, 0), 100) : current > 0 ? 100 : 0
  const hiddenPercentage = 100 - percentage

  return (
    <div className={cn('w-full h-2 bg-muted rounded-full overflow-clip relative', className)}>
      <motion.div
        className="absolute inset-0 bg-[linear-gradient(90deg,#22c55e_0%,#22c55e_60%,#facc15_75%,#ef4444_100%)]"
        initial={{ clipPath: 'inset(0 100% 0 0)' }}
        animate={{ clipPath: `inset(0 ${hiddenPercentage}% 0 0)` }}
        transition={transition}
      />
    </div>
  )
}

export default QuotaLine
