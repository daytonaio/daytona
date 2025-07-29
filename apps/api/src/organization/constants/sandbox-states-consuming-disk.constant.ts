/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxState } from '../../sandbox/enums/sandbox-state.enum'
import { SANDBOX_STATES_CONSUMING_COMPUTE } from './sandbox-states-consuming-compute.constant'

export const SANDBOX_STATES_CONSUMING_DISK: SandboxState[] = [
  ...SANDBOX_STATES_CONSUMING_COMPUTE,
  SandboxState.STOPPED,
  SandboxState.ARCHIVING,
]
