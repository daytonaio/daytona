/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { AnimatePresence, motion } from 'framer-motion'
import { ArrowDownIcon, ArrowUpDownIcon, ArrowUpIcon } from 'lucide-react'

interface Props {
  sort: 'asc' | 'desc' | null
  hideDefaultState?: boolean
  className?: string
}

const motionProps = {
  initial: { opacity: 0, y: 6 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -6 },
  transition: { duration: 0.15 },
}

const PlaceholderIcon = () => <span className="size-4 inline-block" />

export const SortOrderIcon = ({ hideDefaultState = false, sort, className }: Props) => {
  const Icon =
    sort === 'asc'
      ? ArrowUpIcon
      : sort === 'desc'
        ? ArrowDownIcon
        : hideDefaultState
          ? PlaceholderIcon
          : ArrowUpDownIcon

  return (
    <AnimatePresence mode="wait" initial={false}>
      <motion.span
        key={sort || 'none'}
        {...motionProps}
        className={cn(
          'flex items-center justify-center text-muted-foreground/60 transition-colors duration-150',
          'group-hover/sort-header:text-current group-focus-visible/sort-header:text-current',
          { 'text-foreground': sort },
          className,
        )}
        aria-hidden="true"
      >
        <Icon className="size-4" />
      </motion.span>
    </AnimatePresence>
  )
}
