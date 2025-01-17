// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

// Use generics to create a pointer to a value
func Pointer[T any](d T) *T {
	return &d
}

func ArrayMap[T, U any](data []T, f func(T) U) []U {
	res := make([]U, 0, len(data))

	for _, e := range data {
		res = append(res, f(e))
	}

	return res
}

func StringsToInterface(slice []string) []interface{} {
	args := make([]interface{}, len(slice))
	for i, v := range slice {
		args[i] = v
	}
	return args
}
