/*
 * Copyright 2025 Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

export function fromAxiosError(error: any): Error {
  return new Error(error.response?.data?.message || error.response?.data || error.message || error)
}
