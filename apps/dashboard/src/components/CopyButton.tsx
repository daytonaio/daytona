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

const iconProps = {
  initial: { opacity: 0, y: 5 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -5 },
  transition: { duration: 0.1 },
}

function CopyButton({
  value,
  className,
  tooltipText,
  variant = 'ghost',
  autoHide,
  onClick,
  ...props
}: { value: string; tooltipText?: string; autoHide?: boolean } & Omit<
  ComponentProps<typeof TooltipButton>,
  'tooltipText'
>) {
  const [copied, copy] = useCopyToClipboard()

  return (
    <TooltipButton
      tooltipText={tooltipText || (copied ? 'Copied' : 'Copy')}
      onClick={(e) => {
        copy(value)
        onClick?.(e)
      }}
      className={cn(
        'font-sans text-muted-foreground hover:text-foreground',
        {
          'opacity-0 -translate-x-1': autoHide && !copied,
          'group-hover/copy-button:opacity-100 group-hover/copy-button:translate-x-0 group-focus-within/copy-button:opacity-100 group-focus-within/copy-button:translate-x-0':
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
            className={cn('size-4 text-success', { 'size-3.5': props.size === 'icon-xs' })}
            key="copied"
            {...iconProps}
          />
        ) : (
          <MotionCopyIcon className={cn('size-4', { 'size-3': props.size === 'icon-xs' })} key="copy" {...iconProps} />
        )}
      </AnimatePresence>
    </TooltipButton>
  )
}

export { CopyButton }
