/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export enum SandboxClass {
  SMALL = 'small',
  MEDIUM = 'medium',
  LARGE = 'large',
}

export const SandboxClassData = {
  [SandboxClass.SMALL]: {
    cpu: 4,
    memory: 8,
    disk: 30,
  },
  [SandboxClass.MEDIUM]: {
    cpu: 8,
    memory: 16,
    disk: 60,
  },
  [SandboxClass.LARGE]: {
    cpu: 12,
    memory: 24,
    disk: 90,
  },
}
