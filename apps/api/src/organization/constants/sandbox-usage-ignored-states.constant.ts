/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { SandboxState } from '../../sandbox/enums/sandbox-state.enum'

export const SANDBOX_USAGE_IGNORED_STATES: SandboxState[] = [
  SandboxState.DESTROYED,
  SandboxState.ARCHIVED,
  SandboxState.ERROR,
  SandboxState.BUILD_FAILED,
]
