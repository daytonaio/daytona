// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package coderun

import "strings"

func formatArgv(argv []string) string {
	if len(argv) == 0 {
		return ""
	}

	return strings.Join(argv, " ")
}
