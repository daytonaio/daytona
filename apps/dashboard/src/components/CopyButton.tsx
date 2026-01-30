/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'
import { cn } from '@/lib/utils'
import { AnimatePresence, motion } from 'framer-motion'
import { CheckIcon, CopyIcon } from 'lucide-react'
import { ComponentProps } from 'react'
import TooltipButton from './TooltipButton'

const MotionCopyIcon = motion(CopyIcon)
const MotionCheckIcon = motion(CheckIcon)

function CopyButton({
  value,
  copied: controlledCopied,
  copy: controlledCopy,
  className,
  tooltipText,
  variant = 'ghost',
  autoHide,
  ...props
}: {
  value: string
  tooltipText?: string
  autoHide?: boolean
  copied?: string | null
  copy?: (text: string) => void
} & Omit<ComponentProps<typeof TooltipButton>, 'tooltipText'>) {
  const uncontrolledCopy = useCopyToClipboard()
  const isControlled = controlledCopied !== undefined && controlledCopy !== undefined
  const [copied, copy] = isControlled ? ([controlledCopied, controlledCopy] as const) : uncontrolledCopy

  return (
    <TooltipButton
      tooltipText={tooltipText || (copied ? 'Copied' : 'Copy')}
      onClick={() => copy(value)}
      data-state={copied ? 'copied' : 'copy'}
      className={cn(
        'font-sans text-muted-foreground hover:text-foreground duration-150 transition-all',
        {
          'opacity-0 -translate-x-1': autoHide && !copied,
          'group-hover/copy-button:opacity-100 group-hover/copy-button:translate-x-0 focus:opacity-100 focus:translate-x-0':
            autoHide,
          'opacity-100 translate-x-0': autoHide && copied,
        },
        className,
      )}
      variant={variant}
      {...props}
    >
      <AnimatePresence initial={false} mode="wait">
        {copied ? (
          <MotionCheckIcon
            className={cn('size-4 text-success [filter:drop-shadow(0_0_2px_currentColor)]', {
              'size-3.5': props.size === 'icon-xs',
            })}
            key="copied"
            initial={{ opacity: 0 }}
            animate={{
              opacity: [1, 0, 1],
            }}
            exit={{ opacity: 0, transition: { duration: 0.2 } }}
            transition={{ duration: 0.3, times: [0, 0.5, 1] }}
          />
        ) : (
          <MotionCopyIcon
            className={cn('size-4', { 'size-3': props.size === 'icon-xs' })}
            key="copy"
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.8 }}
            transition={{ duration: 0.1 }}
          />
        )}
      </AnimatePresence>
    </TooltipButton>
  )
}

export { CopyButton }
