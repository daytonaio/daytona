/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { cn } from '@/lib/utils'
import { Loader2, Play } from 'lucide-react'
import TooltipButton from '../TooltipButton'

type PlaygroundActionRunButtonProps = {
  isDisabled: boolean
  isRunning: boolean
  onRunActionClick?: () => Promise<void>
  className?: string
}

const PlaygroundActionRunButton: React.FC<PlaygroundActionRunButtonProps> = ({
  isDisabled,
  isRunning,
  onRunActionClick,
  className,
}) => {
  return (
    <TooltipButton
      disabled={isDisabled}
      variant="outline"
      tooltipText="Run"
      onClick={onRunActionClick}
      className={cn('w-8 h-8', className)}
    >
      {isRunning ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="w-4 h-4" />}
    </TooltipButton>
  )
}

export default PlaygroundActionRunButton
