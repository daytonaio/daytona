/*
 * Copyright Daytona Platforms Inc.
 * SPDX-License-Identifier: AGPL-3.0
 */

package utils

import "strings"

// Helper function to quote shell commands
func ShellQuote(s string) string {
	// Simple shell quoting - wrap in single quotes and escape existing single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// Helper function to quote and join multiple shell arguments
func ShellQuoteJoin(args []string) string {
	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = ShellQuote(arg)
	}
	return strings.Join(quotedArgs, " ")
}
