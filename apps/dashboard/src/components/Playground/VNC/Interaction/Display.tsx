/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { DisplayActions } from '@/enums/Playground'
import { Loader2, Play } from 'lucide-react'
import { useState } from 'react'

const VNCDisplayOperations: React.FC = () => {
  const [runningDisplayActionMethod, setRunningDisplayActionMethod] = useState<DisplayActions | null>(null)

  const onDisplayActionRunClick = (displayActionMethodName: DisplayActions) => {
    setRunningDisplayActionMethod(displayActionMethodName)
    //TODO -> API call + set API response as responseText if present
    setRunningDisplayActionMethod(null)
  }

  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <Label htmlFor={DisplayActions.GET_INFO}>getInfo()</Label>
            <p id={DisplayActions.GET_INFO} className="text-sm text-muted-foreground mt-1 pl-1">
              Gets information about the displays
            </p>
          </div>
          <div>
            <Button
              disabled={!!runningDisplayActionMethod}
              variant="outline"
              title="Run"
              onClick={() => onDisplayActionRunClick(DisplayActions.GET_INFO)}
            >
              {runningDisplayActionMethod === DisplayActions.GET_INFO ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Play className="w-4 h-4" />
              )}
            </Button>
          </div>
        </div>
      </div>
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <Label htmlFor={DisplayActions.GET_WINDOWS}>getWindows()</Label>
            <p id={DisplayActions.GET_WINDOWS} className="text-sm text-muted-foreground mt-1 pl-1">
              Gets the list of open windows
            </p>
          </div>
          <div>
            <Button
              disabled={!!runningDisplayActionMethod}
              variant="outline"
              title="Run"
              onClick={() => onDisplayActionRunClick(DisplayActions.GET_WINDOWS)}
            >
              {runningDisplayActionMethod === DisplayActions.GET_WINDOWS ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Play className="w-4 h-4" />
              )}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default VNCDisplayOperations
