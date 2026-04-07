/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export interface SandboxMutationVariables {
  sandboxId: string
}

const sandboxMutationKeyAll = ['sandboxes'] as const

export const mutationKeys = {
  sandboxes: {
    all: sandboxMutationKeyAll,
    start: [...sandboxMutationKeyAll, 'start'] as const,
    recover: [...sandboxMutationKeyAll, 'recover'] as const,
    stop: [...sandboxMutationKeyAll, 'stop'] as const,
    archive: [...sandboxMutationKeyAll, 'archive'] as const,
    delete: [...sandboxMutationKeyAll, 'delete'] as const,
  },
} as const
