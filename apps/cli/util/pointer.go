// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

// Use generics to create a pointer to a value
func Pointer[T any](d T) *T {
	return &d
}
