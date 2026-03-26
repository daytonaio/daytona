'use client'

import { clsx as cn } from 'clsx'
import { AnimatePresence, motion } from 'framer-motion'
import { CheckIcon, CopyIcon } from 'lucide-react'
import type { ComponentProps } from 'react'
import * as React from 'react'

import styles from './Button.module.scss'

function useCopyToClipboard({
  timeout = 2000,
  onCopy,
}: {
  timeout?: number
  onCopy?: () => void
} = {}) {
  const [isCopied, setIsCopied] = React.useState(false)

  const copyToClipboard = (value: string) => {
    if (typeof window === 'undefined' || !navigator.clipboard.writeText) {
      return
    }

    if (!value) {
      return
    }

    navigator.clipboard.writeText(value).then(() => {
      setIsCopied(true)

      onCopy?.()

      if (timeout !== 0) {
        setTimeout(() => {
          setIsCopied(false)
        }, timeout)
      }
    }, console.error)
  }

  return { isCopied, copyToClipboard }
}

const iconProps = {
  initial: { opacity: 0, y: 5 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -5 },
  transition: { duration: 0.125 },
}

const MotionCopyIcon = motion(CopyIcon)
const MotionCheckIcon = motion(CheckIcon)

interface Props {
  value: string
  variant?: 'default' | 'ghost'
}
export function CopyButton({
  value,
  children,
  className,
  variant = 'default',
  ...props
}: Props & ComponentProps<'button'>) {
  const { isCopied, copyToClipboard } = useCopyToClipboard()

  return (
    <button
      className={cn(
        styles.button,
        {
          [styles.ghost]: variant === 'ghost',
          [styles.default]: variant === 'default',
        },
        className
      )}
      type="button"
      onClick={() => copyToClipboard(value)}
      {...props}
    >
      <AnimatePresence initial={false} mode="wait">
        {isCopied ? (
          <MotionCheckIcon size={14} key="copied" {...iconProps} />
        ) : (
          <MotionCopyIcon size={14} key="copy" {...iconProps} />
        )}
      </AnimatePresence>
      {children}
    </button>
  )
}
