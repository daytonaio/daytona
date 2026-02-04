/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Label } from '@/components/ui/label'
import { PlaygroundActionFormDataBasic, PlaygroundActions } from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import PlaygroundActionRunButton from './ActionRunButton'

type PlaygroundActionFormProps<A> = {
  actionFormItem: PlaygroundActionFormDataBasic<A>
  onRunActionClick?: () => Promise<void>
  disable?: boolean
  hideRunActionButton?: boolean
}

function PlaygroundActionForm<A extends PlaygroundActions>({
  actionFormItem,
  onRunActionClick,
  disable,
  hideRunActionButton,
}: PlaygroundActionFormProps<A>) {
  const { runningActionMethod, actionRuntimeError } = usePlayground()

  return (
    <>
      <div className="flex items-center justify-between">
        <div>
          <Label htmlFor={actionFormItem.methodName as string}>{actionFormItem.label}</Label>
          <p id={actionFormItem.methodName as string} className="text-sm text-muted-foreground mt-1">
            {actionFormItem.description}
          </p>
        </div>
        {!hideRunActionButton && (
          <PlaygroundActionRunButton
            isDisabled={disable || !!runningActionMethod}
            isRunning={runningActionMethod === actionFormItem.methodName}
            onRunActionClick={onRunActionClick}
          />
        )}
      </div>
      <div className="empty:hidden">
        {actionRuntimeError[actionFormItem.methodName] && (
          <p className="text-sm text-red-500 mt-2">{actionRuntimeError[actionFormItem.methodName]}</p>
        )}
      </div>
    </>
  )
}

export default PlaygroundActionForm
