/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { GitOperationsActions, PlaygroundActionFormDataBasic } from '@/enums/Playground'
import PlaygroundActionForm from '../../ActionForm'

const SandboxGitOperations: React.FC = () => {
  const gitOperationsActionsFormData: PlaygroundActionFormDataBasic<GitOperationsActions>[] = [
    {
      methodName: GitOperationsActions.GIT_CLONE,
      label: 'clone()',
      description: 'Clones a Git repository into the specified path',
    },
    {
      methodName: GitOperationsActions.GIT_STATUS,
      label: 'status()',
      description: 'Gets the current Git repository status',
    },
    {
      methodName: GitOperationsActions.GIT_BRANCHES_LIST,
      label: 'branches()',
      description: 'Lists branches in the repository',
    },
    {
      methodName: GitOperationsActions.CREATE_BRANCH,
      label: 'createBranch()',
      description: 'Creates branch in the repository',
    },
    {
      methodName: GitOperationsActions.CHECKOUT_BRANCH,
      label: 'checkoutBranch()',
      description: 'Checkout branch in the repository',
    },
    {
      methodName: GitOperationsActions.DELETE_BRANCH,
      label: 'deleteBranch()',
      description: 'Deletes branch in the repository',
    },
    {
      methodName: GitOperationsActions.GIT_ADD,
      label: 'add()',
      description: 'Stages the specified files for the next commit',
    },
    {
      methodName: GitOperationsActions.GIT_COMMIT,
      label: 'commit()',
      description: 'Creates a new commit with the staged changes',
    },
    {
      methodName: GitOperationsActions.GIT_PUSH,
      label: 'push()',
      description: 'Pushes all local commits on the current branch to the remote repository',
    },
    {
      methodName: GitOperationsActions.GIT_PULL,
      label: 'pull()',
      description: 'Pulls changes from the remote repository',
    },
  ]

  return (
    <div className="space-y-6">
      {gitOperationsActionsFormData.map((gitOperationsActionFormData) => (
        <div key={gitOperationsActionFormData.methodName} className="space-y-4">
          <PlaygroundActionForm<GitOperationsActions>
            actionFormItem={gitOperationsActionFormData}
            hideRunActionButton
          />
        </div>
      ))}
    </div>
  )
}

export default SandboxGitOperations
