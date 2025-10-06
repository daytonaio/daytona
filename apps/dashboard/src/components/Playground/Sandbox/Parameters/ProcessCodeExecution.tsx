/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  ProcessCodeExecutionActions,
  PlaygroundActionFormDataBasic,
  CodeRunParams,
  ShellCommandRunParams,
  ParameterFormItem,
} from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { CodeLanguage } from '@daytonaio/sdk'
import PlaygroundActionForm from '../../ActionForm'
import StackedInputFormControl from '../../Inputs/StackedInputFormControl'
import CodeBlock from '@/components/CodeBlock'
import { getLanguageCodeToRun } from '@/lib/playground'
import { useEffect, useState } from 'react'

const SandboxProcessCodeExecution: React.FC = () => {
  const { sandboxParametersState, setSandboxParameterValue } = usePlayground()
  const [codeRunParams, setCodeRunParams] = useState<CodeRunParams>(sandboxParametersState['codeRunParams'])
  const [shellCommandRunParams, setShellCommandRunParams] = useState<ShellCommandRunParams>(
    sandboxParametersState['shellCommandRunParams'],
  )

  const codeRunLanguageCodeFormData: ParameterFormItem & { key: 'languageCode' } = {
    label: 'Code to execute',
    key: 'languageCode',
    placeholder: 'Write the code you want to execute inside the sandbox',
  }

  const shellCommandFormData: ParameterFormItem & { key: 'shellCommand' } = {
    label: 'Shell command',
    key: 'shellCommand',
    placeholder: 'Enter a shell command to run inside the sandbox',
  }

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
  ]

  // Change code to run based on selected sandbox language
  useEffect(() => {
    setCodeRunParams((prev) => {
      const codeRunParamsNew = { ...prev, languageCode: getLanguageCodeToRun(sandboxParametersState.language) }
      setSandboxParameterValue('codeRunParams', codeRunParamsNew)
      return codeRunParamsNew
    })
  }, [sandboxParametersState.language, setSandboxParameterValue])

  return (
    <div className="space-y-6">
      {processCodeExecutionActionsFormData.map((processCodeExecutionAction) => (
        <div key={processCodeExecutionAction.methodName} className="space-y-4">
          <PlaygroundActionForm<ProcessCodeExecutionActions>
            actionFormItem={processCodeExecutionAction}
            hideRunActionButton
          />
          <div className="px-4 space-y-2">
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
