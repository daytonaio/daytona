/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { DisplayActions, PlaygroundActionFormDataBasic } from '@/enums/Playground'
import PlaygroundActionForm from '../../ActionForm'
import { useState } from 'react'

const VNCDisplayOperations: React.FC = () => {
  const [runningDisplayActionMethod, setRunningDisplayActionMethod] = useState<DisplayActions | null>(null)

  const displayActionsFormData: PlaygroundActionFormDataBasic<DisplayActions>[] = [
    {
      methodName: DisplayActions.GET_INFO,
      label: 'getInfo()',
      description: 'Gets information about the displays',
    },
    {
      methodName: DisplayActions.GET_WINDOWS,
      label: 'getWindows()',
      description: 'Gets the list of open windows',
    },
  ]

  const onDisplayActionRunClick = async (displayActionMethodName: DisplayActions) => {
    setRunningDisplayActionMethod(displayActionMethodName)
    //TODO -> API call + set API response as responseText if present
    setRunningDisplayActionMethod(null)
  }

  return (
    <div className="space-y-6">
      {displayActionsFormData.map((displayActionFormData) => (
        <div key={displayActionFormData.methodName} className="space-y-4">
          <PlaygroundActionForm<DisplayActions>
            actionFormItem={displayActionFormData}
            onRunActionClick={() => onDisplayActionRunClick(displayActionFormData.methodName)}
            runningActionMethodName={runningDisplayActionMethod}
          />
        </div>
      ))}
    </div>
  )
}

export default VNCDisplayOperations
