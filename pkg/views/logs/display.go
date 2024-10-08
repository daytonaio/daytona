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

var FIRST_PROJECT_INDEX = 0
var STATIC_INDEX = -1
var WORKSPACE_PREFIX = "WORKSPACE"
var PROVIDER_PREFIX = "PROVIDER"

var longestPrefixLength = len(WORKSPACE_PREFIX)
var maxPrefixLength = 20
var prefixDelimiter = " | "
var prefixPadding = " "

func DisplayLogs(logEntriesChan <-chan logs.LogEntry, index int) {
	for logEntry := range logEntriesChan {
		DisplayLogEntry(logEntry, index)
	}
}

func DisplayLogEntry(logEntry logs.LogEntry, index int) {
	line := logEntry.Msg

	prefixColor := getPrefixColor(index, logEntry.Source)
	var prefixText string

	if logEntry.ProjectName != nil {
		prefixText = *logEntry.ProjectName
	}

	if logEntry.BuildId != nil {
		prefixText = *logEntry.BuildId
	}

	if index == STATIC_INDEX {
		if logEntry.Source == string(logs.LogSourceProvider) {
			prefixText = PROVIDER_PREFIX
		} else {
			prefixText = WORKSPACE_PREFIX
		}
	}

	prefix := lipgloss.NewStyle().Foreground(prefixColor).Bold(true).Render(formatPrefixText(prefixText))

	if index == STATIC_INDEX {
		if logEntry.Source == string(logs.LogSourceProvider) {
			line = fmt.Sprintf("%s%s\033[1m%s\033[0m", prefixPadding, prefix, line)
		} else {
			line = fmt.Sprintf("%s%s%s \033[1m%s\033[0m", prefixPadding, prefix, views.CheckmarkSymbol, line)
		}
		fmt.Print(line)
		return
	}

	// Ensure the cursor moving never overwrites the prefix
	cursorOffset := longestPrefixLength + len(prefixDelimiter) + 2*len(prefixPadding)
	line = strings.ReplaceAll(line, "\r", fmt.Sprintf("\u001b[%dG", cursorOffset))
	line = strings.ReplaceAll(line, "\u001b[0G", fmt.Sprintf("\u001b[%dG", cursorOffset))

	if line == "\n" {
		fmt.Print(line)
		return
	}

	parts := strings.Split(line, "\n")

	var result string
	for _, part := range parts {
		if part == "" {
			continue
		}
		result = fmt.Sprintf("%s\r%s%s%s", result, prefixPadding, prefix, part)
		if len(parts) > 1 {
			result = fmt.Sprintf("%s\n", result)
		}
	}

	fmt.Print(result)
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
		input += prefixPadding
	}

	input += prefixDelimiter
	return input
}

func getPrefixColor(index int, source string) lipgloss.AdaptiveColor {
	if index == STATIC_INDEX {
		if source == string(logs.LogSourceProvider) {
			return views.Yellow
		} else {
			return views.Green
		}
	}
	return views.LogPrefixColors[index%len(views.LogPrefixColors)]
}
