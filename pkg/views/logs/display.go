// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package logs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/views"
)

var FIRST_PROJECT_INDEX = 0
var WORKSPACE_INDEX = -1
var WORKSPACE_PREFIX = "WORKSPACE"

var longestPrefixLength = len(WORKSPACE_PREFIX)
var maxPrefixLength = 20
var prefixDelimiter = " | "
var prefixPadding = " "

var workspaceLogsCursorStart int
var workspaceLogsCursorPosition int
var workspaceLogsCursorThreshold int
var workspaceLogsBoundaries = 1
var projectLogsCursorPosition int

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
		fmt.Print("\033[u")
		if workspaceLogsCursorPosition >= workspaceLogsCursorThreshold {
			workspaceLogsCursorPosition = workspaceLogsCursorStart
			projectLogsCursorPosition += workspaceLogsCursorThreshold - workspaceLogsCursorStart
			fmt.Printf("\033[%dA", workspaceLogsCursorThreshold-workspaceLogsCursorStart)
		}

		// fmt.Printf("\033[%d;0H", workspaceLogsCursorPosition)

		line = fmt.Sprintf("%s%s%s \033[1m%s\033[0m", prefixPadding, prefix, views.CheckmarkSymbol, line)
		fmt.Print(line)

		workspaceLogsCursorPosition += 1
		fmt.Print("\033[s")

		projectLogsCursorPosition -= 1
		fmt.Printf("\033[%dB", projectLogsCursorPosition)
		return
	}

	// Ensure the cursor moving never overwrites the prefix
	cursorOffset := longestPrefixLength + len(prefixDelimiter) + 2*len(prefixPadding)
	line = strings.ReplaceAll(line, "\r", fmt.Sprintf("\u001b[%dG", cursorOffset))
	line = strings.ReplaceAll(line, "\u001b[0G", fmt.Sprintf("\u001b[%dG", cursorOffset))

	if line == "\n" {
		fmt.Print(line)
		// fmt.Print("\033[s")
		projectLogsCursorPosition += 1
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
	// fmt.Print("\033[s")
	projectLogsCursorPosition += 1
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

func getPrefixColor(index int) lipgloss.AdaptiveColor {
	if index == WORKSPACE_INDEX {
		return views.Green
	}
	return views.LogPrefixColors[index%len(views.LogPrefixColors)]
}

func SetCursorPositions() {
	row, _, _ := getCursorPosition()
	workspaceLogsCursorStart = row
	workspaceLogsCursorPosition = row
	workspaceLogsCursorThreshold = row + workspaceLogsBoundaries
	projectLogsCursorPosition = row + workspaceLogsBoundaries
	// fmt.Printf("\033[7;0H")
	fmt.Print("\033[s")
	fmt.Printf("\033[%dB", projectLogsCursorPosition)
}

func getCursorPosition() (row, col int, err error) {
	// Save the terminal state
	var oldState syscall.Termios
	termFd := int(os.Stdin.Fd())
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(termFd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&oldState)), 0, 0, 0); err != 0 {
		return 0, 0, err
	}

	// Disable input buffering
	newState := oldState
	newState.Lflag &^= syscall.ICANON | syscall.ECHO
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(termFd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&newState)), 0, 0, 0); err != 0 {
		return 0, 0, err
	}

	// Restore terminal state afterwards
	defer syscall.Syscall6(syscall.SYS_IOCTL, uintptr(termFd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)

	// Send the cursor position query
	fmt.Print("\x1b[6n")

	// Read the response
	buf := make([]byte, 32)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return 0, 0, err
	}

	// Parse the response
	response := string(buf[:n])
	if response[0] != '\x1b' || response[1] != '[' {
		return 0, 0, fmt.Errorf("unexpected response format")
	}
	response = strings.TrimSuffix(response[2:], "R")
	coords := strings.Split(response, ";")
	if len(coords) != 2 {
		return 0, 0, fmt.Errorf("unexpected response format")
	}

	row, err = strconv.Atoi(coords[0])
	if err != nil {
		return 0, 0, err
	}

	col, err = strconv.Atoi(coords[1])
	if err != nil {
		return 0, 0, err
	}

	return row, col, nil
}
