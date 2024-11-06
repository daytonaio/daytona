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

const FIRST_WORKSPACE_INDEX = 0
const STATIC_INDEX = -1
const TARGET_PREFIX = "TARGET"
const PROVIDER_PREFIX = "PROVIDER"

var longestPrefixLength = len(TARGET_PREFIX)

const maxPrefixLength = 20
const prefixDelimiter = " | "
const prefixPadding = " "

func DisplayLogs(logEntriesChan <-chan logs.LogEntry, index int) {
	for logEntry := range logEntriesChan {
		DisplayLogEntry(logEntry, index)
	}
}

func DisplayLogEntry(logEntry logs.LogEntry, index int) {
	line := logEntry.Msg

	prefixColor := getPrefixColor(index, logEntry.Source)
	var prefixText string

	if logEntry.WorkspaceName != nil {
		prefixText = *logEntry.WorkspaceName
	} else if logEntry.BuildId != nil {
		prefixText = *logEntry.BuildId
	} else if logEntry.TargetName != nil {
		prefixText = *logEntry.TargetName
	}

	if index == STATIC_INDEX && prefixText == "" {
		if logEntry.Source == string(logs.LogSourceProvider) {
			prefixText = PROVIDER_PREFIX
		} else {
			prefixText = TARGET_PREFIX
		}
	}

	prefix := lipgloss.NewStyle().Foreground(prefixColor).Bold(true).Render(formatPrefixText(prefixText))

	if index == STATIC_INDEX {
		fmt.Printf("%s%s\033[1m%s\033[0m", prefixPadding, prefix, line)
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

func CalculateLongestPrefixLength(workspaceNames []string) {
	for _, workspaceName := range workspaceNames {
		if len(workspaceName) > longestPrefixLength {
			longestPrefixLength = len(workspaceName)
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
