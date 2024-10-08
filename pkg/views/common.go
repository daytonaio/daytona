// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type Padding struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

var DocStyle = lipgloss.
	NewStyle().
	Margin(3, 2, 1, 2).
	Padding(1, 2)

var BasicLayout = lipgloss.
	NewStyle().
	Margin(1, 0).
	PaddingLeft(2)

var DefaultListFooterPadding = &Padding{Left: 2}

var DefaultLayoutMarginTop = 1

var DefaultHorizontalMargin = 1

var TUITableMinimumWidth = 80

var CheckmarkSymbol = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")

var SeparatorString = lipgloss.NewStyle().Foreground(LightGray).Render("===")

var (
	minimumWidth     = 40
	maximumWidth     = 160
	widthBreakpoints = []int{60, 80, 100, 120, 140, 160}
)

func RenderMainTitle(title string) {
	fmt.Println(lipgloss.NewStyle().Foreground(Green).Bold(true).Padding(1, 0, 1, 0).Render(title))
}

func RenderLine(message string) {
	fmt.Println(lipgloss.NewStyle().PaddingLeft(1).Render(message))
}

func RenderInfoMessage(message string) {
	fmt.Println(lipgloss.NewStyle().Padding(1, 0, 1, 1).Render(message))
}

func RenderViewBuildLogsMessage(buildId string) {
	RenderInfoMessage(fmt.Sprintf("The build has been scheduled for running. Use `daytona build logs %s -f` to view the progress.", buildId))
}

func RenderCreationInfoMessage(message string) {
	fmt.Println(lipgloss.NewStyle().Foreground(Gray).Padding(1, 0, 1, 1).Render(message))
}

func RenderListLine(message string) {
	fmt.Println(lipgloss.NewStyle().Padding(0, 0, 1, 1).Render(message))
}

func RenderInfoMessageBold(message string) {
	fmt.Println(lipgloss.NewStyle().Bold(true).Padding(1, 0, 1, 1).Render(message))
}

func RenderBorderedMessage(message string) {
	fmt.Println(GetBorderedMessage(message))
}

func GetListFooter(profileName string, padding *Padding) string {
	style := lipgloss.NewStyle().Bold(true)
	style = style.Padding(padding.Top, padding.Right, padding.Bottom, padding.Left)

	return style.Render("\n\nActive profile: " + profileName)
}

func GetStyledMainTitle(content string) string {
	return lipgloss.NewStyle().Foreground(Dark).Background(Light).Padding(0, 1).Render(content)
}

func GetInfoMessage(message string) string {
	return lipgloss.NewStyle().Padding(1, 0, 1, 1).Render(message)
}

func GetBoldedInfoMessage(message string) string {
	return lipgloss.NewStyle().Bold(true).Padding(1, 0, 1, 1).Render(message)
}

func GetListLine(message string) string {
	return lipgloss.NewStyle().Padding(0, 0, 1, 1).Render(message)
}

func GetPropertyKey(key string) string {
	return lipgloss.NewStyle().Foreground(LightGray).Render(key)
}

func GetBranchNameLabel(branch string) string {
	if branch == "" {
		return "Default branch"
	}
	return branch
}

func GetBorderedMessage(message string) string {
	return lipgloss.
		NewStyle().
		Margin(1, 0).
		Padding(1, 2).
		BorderForeground(LightGray).
		Border(lipgloss.RoundedBorder()).
		Render(message)
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

func RenderContainerLayout(output string) {
	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(DocStyle.Render("Error: Unable to get terminal size"))
		return
	}

	fmt.Println(BasicLayout.Width(GetContainerBreakpointWidth(terminalWidth)).Render(output))
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
