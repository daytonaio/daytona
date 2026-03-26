// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var DefaultLayoutMarginTop = 1

var DefaultHorizontalMargin = 1

var TUITableMinimumWidth = 80

var SeparatorString = lipgloss.NewStyle().Foreground(LightGray).Render("===")

var Checkmark = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓").String()

var (
	minimumWidth     = 40
	maximumWidth     = 160
	widthBreakpoints = []int{60, 80, 100, 120, 140, 160}
)

func RenderMainTitle(title string) {
	fmt.Println(lipgloss.NewStyle().Foreground(Green).Bold(true).Padding(1, 0, 1, 0).Render(title))
}

func RenderTip(message string) {
	fmt.Println(lipgloss.NewStyle().Padding(0, 0, 1, 1).Render(message))
}

func RenderInfoMessage(message string) {
	fmt.Println(lipgloss.NewStyle().PaddingLeft(1).Render(message))
}

func RenderInfoMessageBold(message string) {
	fmt.Println(lipgloss.NewStyle().Bold(true).Padding(1, 0, 1, 1).Render(message))
}

func GetStyledMainTitle(content string) string {
	return lipgloss.NewStyle().Foreground(Dark).Background(Light).Padding(0, 1).MarginTop(1).Render(content)
}

func GetInfoMessage(message string) string {
	return lipgloss.NewStyle().Padding(1, 0, 1, 1).Render(message)
}

func GetContainerBreakpointWidth(terminalWidth int) int {
	if terminalWidth < minimumWidth {
		return 0
	}
	for _, width := range widthBreakpoints {
		if terminalWidth < width {
			return width - 20 - DefaultHorizontalMargin - DefaultHorizontalMargin
		}
	}
	return maximumWidth
}

func GetEnvVarsInput(envVars *map[string]string) *huh.Text {
	if envVars == nil {
		return nil
	}

	var inputText string
	for key, value := range *envVars {
		inputText += fmt.Sprintf("%s=%s\n", key, value)
	}
	inputText = strings.TrimSuffix(inputText, "\n")

	return huh.NewText().
		Title("Environment Variables").
		Description("Enter environment variables in the format KEY=VALUE\nTo pass machine env variables at runtime, use $VALUE").
		CharLimit(-1).
		Value(&inputText).
		Validate(func(str string) error {
			tempEnvVars := map[string]string{}
			for i, line := range strings.Split(str, "\n") {
				if line == "" {
					continue
				}

				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid format: %s on line %d", line, i+1)
				}

				tempEnvVars[parts[0]] = parts[1]
			}
			*envVars = tempEnvVars

			return nil
		})
}

// Bolds the message and prepends a checkmark
func GetPrettyLogLine(message string) string {
	return fmt.Sprintf("%s \033[1m%s\033[0m\n", lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓").String(), message)
}
