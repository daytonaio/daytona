// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "crypto/sha256"

func Hash(value string) string {
	hash := sha256.Sum256([]byte(value))
	return string(hash[:])
}
