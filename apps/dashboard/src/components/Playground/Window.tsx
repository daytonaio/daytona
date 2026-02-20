/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { motion } from 'framer-motion'
import { ComponentProps, ReactNode } from 'react'

const Window = ({ children, className = '' }: ComponentProps<'div'>) => {
  return (
    <motion.div
      layoutId="window"
      className={cn(
        'flex flex-col bg-card/80 backdrop-blur-xl dark:border dark:border-border rounded-xl overflow-hidden ring-1 ring-ring/5 shadow-lg shadow-black/5 dark:shadow-black/20',
        className,
      )}
    >
      {children}
    </motion.div>
  )
}

const WindowTitleBar = ({
  children,
  right,
  hideControls = false,
}: {
  children?: ReactNode
  right?: ReactNode
  hideControls?: boolean
} & ComponentProps<'div'>) => {
  return (
    <motion.div
      layout
      layoutId="window-title-bar"
      className="relative items-center h-8 px-4 border-b border-border select-none shrink-0 grid grid-cols-[auto_1fr_auto] bg-muted/50"
    >
      <motion.div className="flex gap-2 z-10" layout>
        {!hideControls && (
          <>
            <div className="w-3 h-3 rounded-full bg-muted-foreground/30" />
            <div className="w-3 h-3 rounded-full bg-muted-foreground/30" />
            <div className="w-3 h-3 rounded-full bg-muted-foreground/30" />
          </>
        )}
      </motion.div>

      <motion.div
        layoutId="window-title-bar-content"
        className="absolute inset-0 flex items-center justify-center pointer-events-none"
      >
        <span className="text-sm text-muted-foreground">{children}</span>
      </motion.div>

      <div className="ml-auto z-10">{right}</div>
    </motion.div>
  )
}

const WindowContent = ({ children, className = '' }: ComponentProps<'div'>) => {
  return (
    <motion.div
      layoutId="window-content"
      layout="position"
      className={cn(
        'flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-border scrollbar-track-background p-4 pb-3.5 text-foreground',
        className,
      )}
    >
      {children}
    </motion.div>
  )
}

export { Window, WindowContent, WindowTitleBar }
