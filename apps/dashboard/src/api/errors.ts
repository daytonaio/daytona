/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export class DaytonaError extends Error {
  public static fromError(error: Error): DaytonaError {
    const errorString = String(error)

    if (errorString.includes('Organization is suspended')) {
      return new OrganizationSuspendedError(error.message)
    }

    // Check for "has active child sandbox(es)" error pattern
    const childrenMatch = errorString.match(/it has (\d+) active child sandbox\(es\)/)
    if (childrenMatch) {
      const childCount = parseInt(childrenMatch[1], 10)
      return new HasChildrenError(error.message, childCount)
    }

    return new DaytonaError(error.message)
  }

  public static fromString(error: string): DaytonaError {
    return DaytonaError.fromError(new Error(error))
  }
}

export class OrganizationSuspendedError extends DaytonaError {}

export class HasChildrenError extends DaytonaError {
  public childCount: number

  constructor(message: string, childCount: number) {
    super(message)
    this.childCount = childCount
  }
}
