/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { MoreThanOrEqual, LessThanOrEqual, Between } from 'typeorm'

/**
 * Creates a TypeORM range filter from min/max values
 * @param minValue - Minimum value (inclusive)
 * @param maxValue - Maximum value (inclusive)
 * @returns TypeORM comparison operator (Between, MoreThanOrEqual, LessThanOrEqual, or undefined)
 */
export function createRangeFilter<T>(minValue?: T, maxValue?: T) {
  if (minValue !== undefined && maxValue !== undefined) {
    return Between(minValue, maxValue)
  } else if (minValue !== undefined) {
    return MoreThanOrEqual(minValue)
  } else if (maxValue !== undefined) {
    return LessThanOrEqual(maxValue)
  }
  return undefined
}
