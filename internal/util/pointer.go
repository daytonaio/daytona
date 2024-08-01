// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

// Use generics to create a pointer to a value
func Pointer[T any](d T) *T {
	return &d
}
