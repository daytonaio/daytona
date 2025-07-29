/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxState } from '../../sandbox/enums/sandbox-state.enum'
import { SANDBOX_USAGE_IGNORED_STATES } from './sandbox-usage-ignored-states.constant'

export const SANDBOX_USAGE_INACTIVE_STATES: SandboxState[] = [
  ...SANDBOX_USAGE_IGNORED_STATES,
  SandboxState.STOPPED,
  SandboxState.ARCHIVING,
]
