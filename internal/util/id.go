// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"unicode"
)

func GenerateIdFromName(name string) string {
	var result strings.Builder

	for _, char := range name {
		if unicode.IsLetter(char) || unicode.IsNumber(char) || char == '-' || char == '_' {
			result.WriteRune(char)
		} else if char == ' ' {
			result.WriteRune('_')
		}
	}

	return strings.ToLower(result.String())
}
