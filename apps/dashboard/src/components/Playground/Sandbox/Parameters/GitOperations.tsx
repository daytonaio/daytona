/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import {
  GitBranchesParams,
  GitCloneParams,
  GitOperationsActionFormData,
  GitStatusParams,
  ParameterFormData,
  ParameterFormItem,
} from '@/contexts/PlaygroundContext'
import { GitOperationsActions } from '@/enums/Playground'
import { usePlayground } from '@/hooks/usePlayground'
import { useState } from 'react'
import PlaygroundActionForm from '../../ActionForm'
import InlineInputFormControl from '../../Inputs/InlineInputFormControl'
import FormTextInput from '../../Inputs/TextInput'

const SandboxGitOperations: React.FC = () => {
  const { sandboxParametersState, playgroundActionParamValueSetter } = usePlayground()
  const [gitCloneParams, setGitCloneParams] = useState<GitCloneParams>(sandboxParametersState['gitCloneParams'])
  const [gitStatusParams, setGitStatusParams] = useState<GitStatusParams>(sandboxParametersState['gitStatusParams'])
  const [gitBranchesParams, setGitBranchesParams] = useState<GitBranchesParams>(
    sandboxParametersState['gitBranchesParams'],
  )

  const gitCloneParamsFormData: ParameterFormData<GitCloneParams> = [
    { label: 'URL', key: 'repositoryURL', placeholder: 'Repository URL to clone from', required: true },
    {
      label: 'Destination',
      key: 'cloneDestinationPath',
      placeholder: 'Path where the repository should be cloned',
      required: true,
    },
    { label: 'Branch', key: 'branchToClone', placeholder: 'Specific branch to clone' },
    { label: 'Commit', key: 'commitToClone', placeholder: 'Specific commit to clone' },
    { label: 'Username', key: 'authUsername', placeholder: 'Git username for authentication' },
    { label: 'Password', key: 'authPassword', placeholder: 'Git password or token for authentication' },
  ]

  const gitRepoLocationFormData: ParameterFormItem & { key: 'repositoryPath' } = {
    label: 'Repo location',
    key: 'repositoryPath',
    placeholder: 'Path to the Git repository root',
    required: true,
  }

  const gitOperationsActionsFormData: GitOperationsActionFormData<
    GitCloneParams | GitStatusParams | GitBranchesParams
  >[] = [
    {
      methodName: GitOperationsActions.GIT_CLONE,
      label: 'clone()',
      description: 'Clones a Git repository into the specified path',
      parametersFormItems: gitCloneParamsFormData,
      parametersState: gitCloneParams,
    },
    {
      methodName: GitOperationsActions.GIT_STATUS,
      label: 'status()',
      description: 'Gets the current Git repository status',
      parametersFormItems: [gitRepoLocationFormData],
      parametersState: gitStatusParams,
    },
    {
      methodName: GitOperationsActions.GIT_BRANCHES_LIST,
      label: 'branches()',
      description: 'Lists branches in the repository',
      parametersFormItems: [gitRepoLocationFormData],
      parametersState: gitBranchesParams,
    },
  ]

  return (
    <div className="space-y-6">
      {gitOperationsActionsFormData.map((gitOperationsAction) => (
        <div key={gitOperationsAction.methodName} className="space-y-4">
          <PlaygroundActionForm<GitOperationsActions> actionFormItem={gitOperationsAction} hideRunActionButton />
          <div className="space-y-2">
            {gitOperationsAction.methodName === GitOperationsActions.GIT_CLONE && (
              <>
                {gitCloneParamsFormData.map((gitCloneParamFormItem) => (
                  <InlineInputFormControl key={gitCloneParamFormItem.key} formItem={gitCloneParamFormItem}>
                    <FormTextInput
                      formItem={gitCloneParamFormItem}
                      textValue={gitCloneParams[gitCloneParamFormItem.key]}
                      onChangeHandler={(value) =>
                        playgroundActionParamValueSetter(
                          gitOperationsAction,
                          gitCloneParamFormItem,
                          setGitCloneParams,
                          'gitCloneParams',
                          value,
                        )
                      }
                    />
                  </InlineInputFormControl>
                ))}
              </>
            )}
            {gitOperationsAction.methodName === GitOperationsActions.GIT_STATUS && (
              <InlineInputFormControl formItem={gitRepoLocationFormData}>
                <FormTextInput
                  formItem={gitRepoLocationFormData}
                  textValue={gitStatusParams[gitRepoLocationFormData.key]}
                  onChangeHandler={(value) =>
                    playgroundActionParamValueSetter(
                      gitOperationsAction,
                      gitRepoLocationFormData,
                      setGitStatusParams,
                      'gitStatusParams',
                      value,
                    )
                  }
                />
              </InlineInputFormControl>
            )}
            {gitOperationsAction.methodName === GitOperationsActions.GIT_BRANCHES_LIST && (
              <InlineInputFormControl formItem={gitRepoLocationFormData}>
                <FormTextInput
                  formItem={gitRepoLocationFormData}
                  textValue={gitBranchesParams[gitRepoLocationFormData.key]}
                  onChangeHandler={(value) =>
                    playgroundActionParamValueSetter(
                      gitOperationsAction,
                      gitRepoLocationFormData,
                      setGitBranchesParams,
                      'gitBranchesParams',
                      value,
                    )
                  }
                />
              </InlineInputFormControl>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

export default SandboxGitOperations
