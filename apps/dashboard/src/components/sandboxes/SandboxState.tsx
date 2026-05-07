/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { SandboxState as SandboxStateType } from '@daytona/api-client'
import { Loader2 } from 'lucide-react'
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import type React from 'react'
import { getStateLabel } from '../SandboxTable/constants'
import { Tooltip, TooltipContent, TooltipTrigger } from '../ui/tooltip'

interface SandboxStateProps {
  state?: SandboxStateType
  errorReason?: string
  recoverable?: boolean
  animate?: boolean
}

interface StatusDotProps {
  className: string
  ping?: boolean
}

function StatusDot({ className, ping }: StatusDotProps) {
  return (
    <div className="relative w-4 h-4 p-1">
      {ping && (
        <div
          className={cn(
            'absolute inset-[3px] rounded-full opacity-80 [animation:ping_1000ms_cubic-bezier(0,0,0.2,1)_1] motion-reduce:animate-none',
            className,
          )}
        />
      )}
      <div className={cn('size-2 rounded-full', className)} />
    </div>
  )
}

type StateIndicatorKey = 'unknown' | 'loading' | 'success' | 'muted' | 'destructive' | 'warning'

function StateIndicator({ indicatorKey, animate }: { indicatorKey: StateIndicatorKey; animate: boolean }) {
  switch (indicatorKey) {
    case 'loading':
      return <Loader2 className="size-3 animate-spin" />
    case 'success':
      return <StatusDot className="bg-emerald-500 dark:bg-success" ping={animate} />
    case 'muted':
      return <StatusDot className="bg-muted-foreground/50" />
    case 'destructive':
      return <StatusDot className="bg-destructive" />
    case 'warning':
      return <StatusDot className="bg-warning" />
    default:
      return <StatusDot className="bg-muted-foreground/20" />
  }
}

function getStateIndicatorKey(state: SandboxStateType, recoverable?: boolean): StateIndicatorKey {
  if (recoverable) {
    return 'warning'
  }

  switch (state) {
    case SandboxStateType.CREATING:
    case SandboxStateType.STARTING:
    case SandboxStateType.STOPPING:
    case SandboxStateType.DESTROYING:
    case SandboxStateType.BUILDING_SNAPSHOT:
    case SandboxStateType.PULLING_SNAPSHOT:
    case SandboxStateType.ARCHIVING:
    case SandboxStateType.RESTORING:
    case SandboxStateType.RESIZING:
    case SandboxStateType.SNAPSHOTTING:
    case SandboxStateType.FORKING:
      return 'loading'
    case SandboxStateType.STARTED:
      return 'success'
    case SandboxStateType.STOPPED:
      return 'muted'
    case SandboxStateType.ERROR:
    case SandboxStateType.BUILD_FAILED:
      return 'destructive'
    default:
      return 'unknown'
  }
}

function getStateTextColor(state: SandboxStateType, recoverable?: boolean) {
  if (recoverable) {
    return 'text-warning-foreground'
  }

  if (state === SandboxStateType.ERROR || state === SandboxStateType.BUILD_FAILED) {
    return 'text-destructive-foreground'
  }

  if (state === SandboxStateType.ARCHIVED) {
    return 'text-muted-foreground'
  }

  return undefined
}

function AnimatedStateIndicator({ indicatorKey, animate }: { indicatorKey: StateIndicatorKey; animate: boolean }) {
  const reduceMotion = useReducedMotion()
  const isMuted = indicatorKey === 'muted'

  if (!animate) {
    return (
      <div className="w-4 h-4 flex items-center justify-center flex-shrink-0">
        <StateIndicator indicatorKey={indicatorKey} animate={false} />
      </div>
    )
  }

  const motionProps = reduceMotion
    ? {
        initial: false,
        animate: { opacity: 1 },
        exit: { opacity: 0 },
        transition: { duration: 0.01 },
      }
    : {
        initial: { opacity: 0, scale: isMuted ? 1.2 : 0.8 },
        animate: { opacity: 1, scale: 1 },
        exit: { opacity: 0, scale: 0.8 },
        transition: { duration: isMuted ? 0.28 : 0.2 },
      }

  return (
    <div className="w-4 h-4 flex items-center justify-center flex-shrink-0">
      <AnimatePresence mode="popLayout">
        <motion.div key={indicatorKey} className="origin-center" {...motionProps}>
          <StateIndicator indicatorKey={indicatorKey} animate />
        </motion.div>
      </AnimatePresence>
    </div>
  )
}

function AnimatedStateLabel({ label, animate }: { label: string; animate: boolean }) {
  const reduceMotion = useReducedMotion()

  if (!animate) {
    return <span className="min-w-0 truncate">{label}</span>
  }

  const motionProps = reduceMotion
    ? {
        initial: false,
        animate: { opacity: 1 },
        exit: { opacity: 0 },
        transition: { duration: 0.01 },
      }
    : {
        initial: { opacity: 0, x: -5, filter: 'blur(1px)' },
        animate: { opacity: 1, x: 0, filter: 'blur(0)' },
        exit: { opacity: 0, x: 5, filter: 'blur(1px)' },
        transition: { duration: 0.25 },
      }

  return (
    <span className="min-w-0 overflow-hidden">
      <AnimatePresence mode="popLayout" initial={false}>
        <motion.span key={label} className="block truncate origin-left" {...motionProps}>
          {label}
        </motion.span>
      </AnimatePresence>
    </span>
  )
}

function SandboxStateContent({
  indicatorKey,
  label,
  className,
  animate,
}: {
  indicatorKey: StateIndicatorKey
  label: string
  className?: string
  animate: boolean
}) {
  return (
    <div className={cn('flex items-center gap-1', className)}>
      <AnimatedStateIndicator indicatorKey={indicatorKey} animate={animate} />
      <AnimatedStateLabel label={label} animate={animate} />
    </div>
  )
}

export function SandboxState({ state, errorReason, recoverable, animate = false }: SandboxStateProps) {
  if (!state) return null
  const indicatorKey = getStateIndicatorKey(state, recoverable)
  const label = getStateLabel(state)
  const textColor = getStateTextColor(state, recoverable)

  if (state === SandboxStateType.ERROR || state === SandboxStateType.BUILD_FAILED) {
    const errorContent = (
      <SandboxStateContent indicatorKey={indicatorKey} label={label} className={textColor} animate={animate} />
    )

    if (!errorReason) {
      return errorContent
    }

    return (
      <Tooltip delayDuration={100}>
        <TooltipTrigger asChild>
          <div className="inline-flex">{errorContent}</div>
        </TooltipTrigger>
        <TooltipContent>
          <p className="max-w-[300px]">{errorReason}</p>
        </TooltipContent>
      </Tooltip>
    )
  }

  return <SandboxStateContent indicatorKey={indicatorKey} label={label} className={textColor} animate={animate} />
}
