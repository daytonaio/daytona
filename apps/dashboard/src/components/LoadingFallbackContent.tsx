/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { useQueryClient } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { useEffect, useState } from 'react'
import { Spinner } from './ui/spinner'

type LoadingFallbackContentProps = {
  className?: string
  source?: string
}

export function LoadingFallbackContent({ className, source = 'unknown' }: LoadingFallbackContentProps) {
  const [showLongLoadingMessage, setShowLongLoadingMessage] = useState(false)
  const queryClient = useQueryClient()

  useEffect(() => {
    const timer = setTimeout(() => {
      setShowLongLoadingMessage(true)
      console.warn('[dashboard] Loading fallback still mounted after 5s', { source })
    }, 5_000)

    return () => clearTimeout(timer)
  }, [queryClient, source])

  return (
    <div className={cn('flex items-center justify-center flex-col gap-2', className)} data-loading-source={source}>
      <Spinner className="w-8 h-8 animate-spin" />
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={showLongLoadingMessage ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
        transition={{ duration: 0.35 }}
      >
        <p className="text-sm text-muted-foreground text-center">This is taking longer than expected...</p>
        <p className="text-sm text-muted-foreground text-center">
          If this issue persists, contact us at{' '}
          <a href="mailto:support@daytona.io" className="text-primary underline">
            support@daytona.io
          </a>
          .
        </p>
      </motion.div>
    </div>
  )
}
