/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { ProcessCodeExecutionActions, PlaygroundActionFormDataBasic } from '@/enums/Playground'
import PlaygroundActionForm from '../../ActionForm'

const SandboxProcessCodeExecution: React.FC = () => {
  const processCodeExecutionActionsFormData: PlaygroundActionFormDataBasic<ProcessCodeExecutionActions>[] = [
    {
      methodName: ProcessCodeExecutionActions.CODE_RUN,
      label: 'codeRun()',
      description: 'Executes code in the Sandbox using the appropriate language runtime',
    },
    {
      methodName: ProcessCodeExecutionActions.SHELL_COMMANDS_RUN,
      label: 'executeCommand()',
      description: 'Executes a shell command in the Sandbox',
    },
    {
      methodName: ProcessCodeExecutionActions.CREATE_SESSION,
      label: 'createSession()',
      description: 'Creates a new long-running background session in the Sandbox',
    },
    {
      methodName: ProcessCodeExecutionActions.GET_SESSION,
      label: 'getSession()',
      description: 'Gets a session in the Sandbox',
    },
    {
      methodName: ProcessCodeExecutionActions.LIST_SESSIONS,
      label: 'listSessions()',
      description: 'Lists all sessions in the Sandbox',
    },
    {
      methodName: ProcessCodeExecutionActions.DELETE_SESSION,
      label: 'deleteSession()',
      description: 'Terminates and removes a session from the Sandbox, cleaning up any resources associated with it',
    },
  ]

  return (
    <div className="space-y-6">
      {processCodeExecutionActionsFormData.map((processCodeExecutionActionFormData) => (
        <div key={processCodeExecutionActionFormData.methodName} className="space-y-4">
          <PlaygroundActionForm<ProcessCodeExecutionActions>
            actionFormItem={processCodeExecutionActionFormData}
            hideRunActionButton
          />
        </div>
      ))}
    </div>
  )
}

export default SandboxProcessCodeExecution
