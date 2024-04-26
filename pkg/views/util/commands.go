// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import "strings"

func GetSplitCommands(commands string) []string {
	splitCommands := []string{}

	for _, command := range splitEscaped(commands, ',') {
		splitCommand := strings.ReplaceAll(command, "\\,", ",")
		splitCommand = strings.TrimLeft(splitCommand, " ")
		splitCommands = append(splitCommands, splitCommand)
	}

	return splitCommands
}

func GetJoinedCommands(commands []string) string {
	joinedCommands := ""

	for _, command := range commands {
		joinedCommands += strings.ReplaceAll(command, ",", "\\,") + ","
	}

	joinedCommands = strings.TrimRight(joinedCommands, ",")
	return joinedCommands
}

func splitEscaped(s string, sep rune) []string {
	var result []string
	var builder strings.Builder
	escaping := false

	for _, c := range s {
		if c == '\\' && !escaping {
			escaping = true
			continue
		}

		if c == sep && !escaping {
			result = append(result, builder.String())
			builder.Reset()
			continue
		}

		builder.WriteRune(c)
		escaping = false
	}

	result = append(result, builder.String())
	return result
}
