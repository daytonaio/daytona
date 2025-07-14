/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxState as SandboxStateType } from '@daytonaio/api-client'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '../ui/tooltip'
import { getStateLabel } from './constants'
import { STATE_ICONS } from './state-icons'

interface SandboxStateProps {
  state?: SandboxStateType
  errorReason?: string
}

export function SandboxState({ state, errorReason }: SandboxStateProps) {
  if (!state) return null
  const stateIcon = STATE_ICONS[state] || STATE_ICONS[SandboxStateType.UNKNOWN]
  const label = getStateLabel(state)

  if (state === SandboxStateType.ERROR || state === SandboxStateType.BUILD_FAILED) {
    const errorContent = (
      <div className={`flex items-center gap-1 text-red-600 dark:text-red-400`}>
        {stateIcon}
        {label}
      </div>
    )

    if (!errorReason) {
      return errorContent
    }

    return (
      <TooltipProvider delayDuration={100}>
        <Tooltip>
          <TooltipTrigger asChild>{errorContent}</TooltipTrigger>
          <TooltipContent>
            <p className="max-w-[300px]">{errorReason}</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    )
  }

  return (
    <div className={`flex items-center gap-1 ${state === SandboxStateType.ARCHIVED ? 'text-muted-foreground' : ''}`}>
      <div className="w-4 h-4 flex items-center justify-center">{stateIcon}</div>
      <span>{label}</span>
    </div>
  )
}
