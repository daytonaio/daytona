/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ComponentProps, ReactNode } from 'react'
import { Button } from './ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'

type Props = ComponentProps<typeof Button> & {
  tooltipText: string
  tooltipContent?: ReactNode
  tooltipContainer?: HTMLElement
  side?: ComponentProps<typeof TooltipContent>['side']
}

function TooltipButton({
  tooltipText,
  tooltipContent,
  side = 'top',
  tooltipContainer,
  ref,
  size = 'icon-sm',
  ...props
}: Props) {
  return (
    <Tooltip delayDuration={0}>
      <TooltipTrigger asChild>
        <Button ref={ref} {...props} size={size} aria-label={tooltipText} />
      </TooltipTrigger>
      <TooltipContent side={side}>{tooltipContent || <div>{tooltipText}</div>}</TooltipContent>
    </Tooltip>
  )
}

export default TooltipButton
