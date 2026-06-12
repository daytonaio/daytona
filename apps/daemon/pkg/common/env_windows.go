//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"strings"
)

// Windows environment variables are case-insensitive: if a request carries
// both FOO and foo, Go map iteration order decides which one survives
// deduplication, making execution nondeterministic. Reject such requests.
func validateEnvKeysPlatform(envs map[string]string) error {
	lowerKeys := make(map[string]string, len(envs))
	for key := range envs {
		lower := strings.ToLower(key)
		if existing, ok := lowerKeys[lower]; ok {
			return fmt.Errorf("environment variable names '%s' and '%s' collide on case-insensitive Windows", existing, key)
		}
		lowerKeys[lower] = key
	}
	return nil
}
