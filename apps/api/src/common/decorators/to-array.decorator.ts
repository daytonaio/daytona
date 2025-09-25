/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

import { Transform } from 'class-transformer'

/**
 * Decorator that transforms a value to an array. Useful for query parameters that can be a single value or an array of values.
 *
 * If the value is a primitive, it will return a single element array. If the value is already an array, it will return it as is.
 */
export function ToArray() {
  return Transform(({ value }) => {
    return value ? (Array.isArray(value) ? value : [value]) : undefined
  })
}
