/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */
import { SandboxClass } from '../enums/sandbox-class.enum'

/**
 * Temporary utility function to allow Android snapshots/sandboxes on Container runners
 */
export function getRunnerSandboxClass(sandboxClass: SandboxClass): SandboxClass {
  if (sandboxClass !== SandboxClass.ANDROID) {
    return sandboxClass
  }

  return SandboxClass.CONTAINER
}
