// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/views"
)

var WORKSPACE_INDEX = -1
var WORKSPACE_PREFIX = "WORKSPACE"

var longestPrefixLength = len(WORKSPACE_PREFIX)
var maxPrefixLength = 20

func DisplayLogs(logEntriesChan <-chan logs.LogEntry, index int) {
	for logEntry := range logEntriesChan {
		DisplayLogEntry(logEntry, index)
	}
}

func DisplayLogEntry(logEntry logs.LogEntry, index int) {
	line := logEntry.Msg

	prefixColor := getPrefixColor(index)
	prefixText := logEntry.ProjectName

	if index == WORKSPACE_INDEX {
		prefixText = WORKSPACE_PREFIX
	}

	prefix := lipgloss.NewStyle().Foreground(prefixColor).Bold(true).Render(formatPrefixText(prefixText))

	if index == WORKSPACE_INDEX {
		line = fmt.Sprintf(" %s%s \033[1m%s\033[0m", prefix, views.CheckmarkSymbol, line)
	} else {
		// Check if carriage return exists and if it does, remove the characters before it unless it is the last character of the line
		lastIndex := strings.LastIndex(line, "\r")
		if lastIndex != -1 {
			if !strings.HasSuffix(line, "\r") && !strings.HasSuffix(line, "\r\n") {
				line = line[lastIndex+1:]
			}
		}

		line = fmt.Sprintf("\r %s%s", prefix, line)
	}

	fmt.Print(line)
}

func CalculateLongestPrefixLength(projectNames []string) {
	for _, projectName := range projectNames {
		if len(projectName) > longestPrefixLength {
			longestPrefixLength = len(projectName)
		}
	}
}

func formatPrefixText(input string) string {
	prefixLength := longestPrefixLength
	if prefixLength > maxPrefixLength {
		prefixLength = maxPrefixLength
		longestPrefixLength = maxPrefixLength
	}

	// Trim input if longer than maxPrefixLength
	if len(input) > prefixLength {
		input = input[:prefixLength-3]
		input += "..."
	}

	// Pad input with spaces if shorter than maxPrefixLength
	for len(input) < prefixLength {
		input += " "
	}

	input += " | "
	return input
}

func getPrefixColor(index int) lipgloss.AdaptiveColor {
	if index == WORKSPACE_INDEX {
		return views.Green
	}
	return views.LogPrefixColors[index%len(views.LogPrefixColors)]
}
