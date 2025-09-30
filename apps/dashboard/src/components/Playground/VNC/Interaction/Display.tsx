/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { DisplayActions, PlaygroundActionFormDataBasic } from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { PlaygroundActionInvokeApi } from '@/contexts/PlaygroundContext'
import PlaygroundActionForm from '../../ActionForm'

const VNCDisplayOperations: React.FC = () => {
  const { runPlaygroundActionWithoutParams } = usePlayground()

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

  //TODO -> Implementation
  const displayActionAPICall: PlaygroundActionInvokeApi = async (displayActionFormData) => {
    return
  }

  return (
    <div className="space-y-6">
      {displayActionsFormData.map((displayActionFormData) => (
        <div key={displayActionFormData.methodName} className="space-y-4">
          <PlaygroundActionForm<DisplayActions>
            actionFormItem={displayActionFormData}
            onRunActionClick={() => runPlaygroundActionWithoutParams(displayActionFormData, displayActionAPICall)}
          />
        </div>
      ))}
    </div>
  )
}

export default VNCDisplayOperations
