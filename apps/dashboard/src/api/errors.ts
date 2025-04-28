/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class DaytonaError extends Error {
  public static fromError(error: Error): DaytonaError {
    if (String(error).includes('Organization is suspended')) {
      return new OrganizationSuspendedError(error.message)
    }

    return new DaytonaError(error.message)
  }

  public static fromString(error: string): DaytonaError {
    return DaytonaError.fromError(new Error(error))
  }
}

export class OrganizationSuspendedError extends DaytonaError {}
