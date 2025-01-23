// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/views"
)

const FIRST_WORKSPACE_INDEX = 0
const STATIC_INDEX = -1

var minimumLongestPrefixLength = 4

const maxPrefixLength = 20
const prefixDelimiter = " | "
const prefixPadding = " "

func DisplayLogsFromReader(reader io.Reader) {
	for {
		buf := make([]byte, 1024)
		n, err := reader.Read(buf)
		if err != nil {
			break
		}
		fmt.Print(string(buf[:n]))
	}
}

func DisplayLogs(logEntriesChan <-chan logs.LogEntry, index int) {
	for logEntry := range logEntriesChan {
		DisplayLogEntry(logEntry, index)
	}
}

func DisplayLogEntry(logEntry logs.LogEntry, index int) {
	line := logEntry.Msg

	prefixColor := getPrefixColor(index, logEntry.Source)
	prefixText := logEntry.Label

	if prefixText == "" {
		prefixText = strings.ToUpper(logEntry.Source)
	}

	prefix := lipgloss.NewStyle().Foreground(prefixColor).Bold(true).Render(formatPrefixText(prefixText))

	if index == STATIC_INDEX {
		fmt.Printf("%s%s\033[1m%s\033[0m", prefixPadding, prefix, line)
		return
	}

	// Ensure the cursor moving never overwrites the prefix
	cursorOffset := minimumLongestPrefixLength + len(prefixDelimiter) + 2*len(prefixPadding)
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

func SetupLongestPrefixLength(workspaceNames []string) {
	minimumLongestPrefixLength = len(slices.MaxFunc(workspaceNames, func(a, b string) int {
		return len(a) - len(b)
	}))
}

func formatPrefixText(input string) string {
	prefixLength := minimumLongestPrefixLength
	if prefixLength > maxPrefixLength {
		prefixLength = maxPrefixLength
		minimumLongestPrefixLength = maxPrefixLength
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
