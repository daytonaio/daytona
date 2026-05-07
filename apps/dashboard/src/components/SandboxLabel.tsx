/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { Separator } from './ui/separator'

interface SandboxLabelProps {
  labelKey: string
  value: string
  className?: string
}

export function SandboxLabel({ labelKey, value, className }: SandboxLabelProps) {
  return (
    <code
      className={cn(
        'flex items-center gap-2 bg-muted rounded px-2 py-0.5 text-xs font-mono border border-border',
        className,
      )}
    >
      <span className="text-muted-foreground">{labelKey}</span>
      <Separator orientation="vertical" className="self-stretch -my-0.5 h-[calc(100%+0.25rem)]" />
      <span>{value}</span>
    </code>
  )
}
