//go:build linux

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

// validateEnvKeysPlatform is a no-op on Linux: environment variable names
// are case-sensitive, so no cross-key collisions exist beyond the pattern
// check in ValidateEnvKeys.
func validateEnvKeysPlatform(map[string]string) error {
	return nil
}
