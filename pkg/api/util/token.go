// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "strings"

func ExtractToken(bearerToken string) string {
	if !strings.HasPrefix(bearerToken, "Bearer ") {
		return ""
	}

	return strings.TrimPrefix(bearerToken, "Bearer ")
}
