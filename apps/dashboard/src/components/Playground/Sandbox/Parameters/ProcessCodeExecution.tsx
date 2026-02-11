/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import CodeBlock from '@/components/CodeBlock'
import {
  CodeRunParams,
  ParameterFormItem,
  ProcessCodeExecutionOperationsActionFormData,
  ShellCommandRunParams,
} from '@/contexts/PlaygroundContext'
import { ProcessCodeExecutionActions } from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { CodeLanguage } from '@daytonaio/sdk'
import PlaygroundActionForm from '../../ActionForm'
import StackedInputFormControl from '../../Inputs/StackedInputFormControl'

const SandboxProcessCodeExecution: React.FC = () => {
  const { sandboxParametersState, setSandboxParameterValue } = usePlayground()
  const codeRunParams = sandboxParametersState['codeRunParams']
  const shellCommandRunParams = sandboxParametersState['shellCommandRunParams']

  const codeRunLanguageCodeFormData: ParameterFormItem & { key: 'languageCode' } = {
    label: 'Code to execute',
    key: 'languageCode',
    placeholder: 'Write the code you want to execute inside the sandbox',
    required: true,
  }

  const shellCommandFormData: ParameterFormItem & { key: 'shellCommand' } = {
    label: 'Shell command',
    key: 'shellCommand',
    placeholder: 'Enter a shell command to run inside the sandbox',
    required: true,
  }

  const processCodeExecutionActionsFormData: ProcessCodeExecutionOperationsActionFormData<
    CodeRunParams | ShellCommandRunParams
  >[] = [
    {
      methodName: ProcessCodeExecutionActions.CODE_RUN,
      label: 'codeRun()',
      description: 'Executes code in the Sandbox using the appropriate language runtime',
      parametersFormItems: [codeRunLanguageCodeFormData],
      parametersState: codeRunParams,
    },
    {
      methodName: ProcessCodeExecutionActions.SHELL_COMMANDS_RUN,
      label: 'executeCommand()',
      description: 'Executes a shell command in the Sandbox',
      parametersFormItems: [shellCommandFormData],
      parametersState: shellCommandRunParams,
    },
  ]

  //TODO -> Currently codeRun and executeCommand values are fixed -> when we enable user to define them implement onChange handlers with validatePlaygroundActionWithParams logic
  return (
    <div className="space-y-6">
      {processCodeExecutionActionsFormData.map((processCodeExecutionAction) => (
        <div key={processCodeExecutionAction.methodName} className="space-y-4">
          <PlaygroundActionForm<ProcessCodeExecutionActions>
            actionFormItem={processCodeExecutionAction}
            hideRunActionButton
          />
          <div className="space-y-2">
            {processCodeExecutionAction.methodName === ProcessCodeExecutionActions.CODE_RUN && (
              <StackedInputFormControl formItem={codeRunLanguageCodeFormData}>
                <CodeBlock
                  language={sandboxParametersState.language || CodeLanguage.PYTHON} // Python is default language if none specified
                  code={codeRunParams[codeRunLanguageCodeFormData.key] || ''}
                />
              </StackedInputFormControl>
            )}
            {processCodeExecutionAction.methodName === ProcessCodeExecutionActions.SHELL_COMMANDS_RUN && (
              <StackedInputFormControl formItem={shellCommandFormData}>
                <CodeBlock language="bash" code={shellCommandRunParams[shellCommandFormData.key] || ''} />
              </StackedInputFormControl>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

export default SandboxProcessCodeExecution
