/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxState } from '@daytonaio/api-client'
import { Loader2 } from 'lucide-react'

interface SquareProps {
  color: string
}

function Square({ color }: SquareProps) {
  return (
    <div className="w-4 h-4 p-1">
      <div className={`w-2 h-2 ${color} rounded-[2px]`} />
    </div>
  )
}

export const STATE_ICONS: Record<SandboxState, React.ReactNode> = {
  [SandboxState.UNKNOWN]: <Square color="bg-muted-foreground/20" />,
  [SandboxState.CREATING]: <Loader2 className="w-3 h-3 animate-spin" />,
  [SandboxState.STARTING]: <Loader2 className="w-3 h-3 animate-spin" />,
  [SandboxState.STARTED]: <Square color="bg-green-600" />,
  [SandboxState.STOPPING]: <Loader2 className="w-3 h-3 animate-spin" />,
  [SandboxState.STOPPED]: <Square color="bg-muted-foreground/50" />,
  [SandboxState.DESTROYING]: <Loader2 className="w-3 h-3 animate-spin" />,
  [SandboxState.DESTROYED]: <Square color="bg-muted-foreground/20" />,
  [SandboxState.ERROR]: <Square color="bg-destructive" />,
  [SandboxState.BUILD_FAILED]: <Square color="bg-destructive" />,
  [SandboxState.BUILDING_SNAPSHOT]: <Loader2 className="w-3 h-3 animate-spin" />,
  [SandboxState.PULLING_SNAPSHOT]: <Loader2 className="w-3 h-3 animate-spin" />,
  [SandboxState.PENDING_BUILD]: <Square color="bg-muted-foreground/20" />,
  [SandboxState.ARCHIVING]: <Loader2 className="w-3 h-3 animate-spin" />,
  [SandboxState.ARCHIVED]: <Square color="bg-muted-foreground/20" />,
  [SandboxState.RESTORING]: <Loader2 className="w-3 h-3 animate-spin" />,
}
