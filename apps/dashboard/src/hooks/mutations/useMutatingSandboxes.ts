/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { usePendingMutationKeys } from '@/hooks/usePendingMutationKeys'
import { mutationKeys } from './mutationKeys'

type SandboxMutationVariables = {
  sandboxId?: string
}

const sandboxMutationSelectors = [
  {
    mutationKey: mutationKeys.sandboxes.start(),
    getKey: (variables?: SandboxMutationVariables) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.stop(),
    getKey: (variables?: SandboxMutationVariables) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.archive(),
    getKey: (variables?: SandboxMutationVariables) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.recover(),
    getKey: (variables?: SandboxMutationVariables) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.remove(),
    getKey: (variables?: SandboxMutationVariables) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.fork(),
    getKey: (variables?: SandboxMutationVariables) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.createSnapshot(),
    getKey: (variables?: SandboxMutationVariables) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.vnc(),
    getKey: (variables?: SandboxMutationVariables) => variables?.sandboxId,
  },
  {
    mutationKey: mutationKeys.sandboxes.screenRecordings(),
    getKey: (variables?: SandboxMutationVariables) => variables?.sandboxId,
  },
]

export function useMutatingSandboxes(): Set<string> {
  return usePendingMutationKeys<string, SandboxMutationVariables>(sandboxMutationSelectors)
}
