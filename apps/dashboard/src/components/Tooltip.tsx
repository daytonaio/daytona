/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { TooltipContent, TooltipTrigger, Tooltip as UiTooltip } from '@/components/ui/tooltip'
import React from 'react'

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
    <UiTooltip>
      <TooltipTrigger asChild>{label}</TooltipTrigger>
      <TooltipContent side={side}>{content}</TooltipContent>
    </UiTooltip>
  )
}
