// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import "strings"

// shellQuoteJoin quotes each argument for safe use in a shell command string.
// Each arg is wrapped in single quotes, with any internal single quotes escaped.
func ShellQuoteJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = "'" + strings.ReplaceAll(arg, "'", "'\\''") + "'"
	}
	return strings.Join(quoted, " ")
}
