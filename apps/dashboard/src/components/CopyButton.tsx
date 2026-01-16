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
  ...props
}: { value: string; tooltipText?: string } & Omit<ComponentProps<typeof TooltipButton>, 'tooltipText'>) {
  const [copied, copy] = useCopyToClipboard()

  return (
    <TooltipButton
      tooltipText={tooltipText || (copied ? 'Copied' : 'Copy')}
      onClick={() => copy(value)}
      className={cn('w-8 h-8 font-sans', className)}
      {...props}
    >
      <AnimatePresence initial={false} mode="wait">
        {copied ? (
          <MotionCheckIcon className="size-4" key="copied" {...iconProps} />
        ) : (
          <MotionCopyIcon className="size-4" key="copy" {...iconProps} />
        )}
      </AnimatePresence>
    </TooltipButton>
  )
}

export { CopyButton }
