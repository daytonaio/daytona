/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import React from 'react'
import { TooltipProvider, Tooltip as UiTooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip'

export function Tooltip({
  label,
  content,
  side = 'top',
}: {
  label: React.ReactNode
  content: React.ReactNode
  side?: 'right' | 'left' | 'top' | 'bottom'
}) {
  return (
    <TooltipProvider>
      <UiTooltip>
        <TooltipTrigger asChild>{label}</TooltipTrigger>
        <TooltipContent side={side}>{content}</TooltipContent>
      </UiTooltip>
    </TooltipProvider>
  )
}
