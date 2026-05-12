/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { AnimatePresence, motion } from 'framer-motion'
import { ArrowDownIcon, ArrowUpDownIcon, ArrowUpIcon } from 'lucide-react'
import { useEffect, useRef } from 'react'

type SortState = 'asc' | 'desc' | null

interface Props {
  sort: SortState
  hideDefaultState?: boolean
  className?: string
}

type SortMotionCustom = {
  sort: SortState
  previousSort: SortState
  nextSort: SortState
}

const getUnsortedY = (sort: SortState) => {
  if (sort === 'desc') return 6
  if (sort === 'asc') return -6
  return 0
}

const sortIconVariants = {
  initial: ({ sort, previousSort }: SortMotionCustom) => {
    if (!sort) {
      const y = getUnsortedY(previousSort)
      return y ? { opacity: 0, y } : { opacity: 0, scale: 0.8 }
    }

    return sort === 'desc' ? { opacity: 0, y: -6 } : { opacity: 0, y: 6 }
  },
  animate: { opacity: 1, y: 0, scale: 1 },
  exit: ({ sort, nextSort }: SortMotionCustom) => {
    if (!sort) {
      const y = getUnsortedY(nextSort)
      return y ? { opacity: 0, y } : { opacity: 0, scale: 0.8 }
    }

    return sort === 'desc' ? { opacity: 0, y: 6 } : { opacity: 0, y: -6 }
  },
}

const getMotionCustom = (sort: SortState, previousSort: SortState): SortMotionCustom => ({
  sort,
  previousSort,
  nextSort: sort,
})

const getMotionProps = (sort: SortState, previousSort: SortState) => {
  return {
    custom: getMotionCustom(sort, previousSort),
    variants: sortIconVariants,
    initial: 'initial',
    animate: 'animate',
    exit: 'exit',
    transition: { duration: 0.15 },
  }
}

const PlaceholderIcon = () => <span className="size-4 inline-block" />

export const SortOrderIcon = ({ hideDefaultState = false, sort, className }: Props) => {
  const previousSortRef = useRef<SortState>(null)
  const previousSort = previousSortRef.current
  const Icon =
    sort === 'asc'
      ? ArrowUpIcon
      : sort === 'desc'
        ? ArrowDownIcon
        : hideDefaultState
          ? PlaceholderIcon
          : ArrowUpDownIcon

  useEffect(() => {
    previousSortRef.current = sort
  }, [sort])

  return (
    <AnimatePresence mode="wait" initial={false} custom={getMotionCustom(sort, previousSort)}>
      <motion.span
        key={sort || 'none'}
        {...getMotionProps(sort, previousSort)}
        data-sort={sort}
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
