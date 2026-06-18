/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'

type MiddleTruncateProps = Omit<React.ComponentProps<'span'>, 'children'> & {
  end?: number
  start?: number
  value?: string | null
}

function MiddleTruncate({ className, end = 4, start = 8, title, value, ...props }: MiddleTruncateProps) {
  const text = value ?? ''
  const startLength = Math.max(0, Math.floor(start))
  const endLength = Math.max(0, Math.floor(end))
  const shouldSplit = text.length > startLength + endLength && startLength + endLength > 0

  const startText = shouldSplit ? text.slice(0, startLength) : text
  const middleText = shouldSplit ? text.slice(startLength, text.length - endLength) : ''
  const endText = shouldSplit && endLength > 0 ? text.slice(-endLength) : ''

  return (
    <span
      className={cn('inline-flex min-w-0 max-w-full overflow-hidden whitespace-nowrap align-bottom', className)}
      data-slot="middle-truncate"
      title={title ?? text}
      {...props}
    >
      <span className="shrink-0">{startText}</span>
      {middleText ? <span className="min-w-0 flex-1 overflow-hidden text-ellipsis">{middleText}</span> : null}
      {endText ? <span className="shrink-0">{endText}</span> : null}
    </span>
  )
}

export { MiddleTruncate }
