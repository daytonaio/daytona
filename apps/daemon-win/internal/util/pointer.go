// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

// Pointer returns a pointer to the given value
func Pointer[T any](v T) *T {
	return &v
}
