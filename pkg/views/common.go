// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var DocStyle = lipgloss.
	NewStyle().
	Margin(3, 2, 1, 2).
	Padding(1, 2)

var BasicLayout = lipgloss.
	NewStyle().
	Margin(1, 0).
	PaddingLeft(2)

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

func GetListFooter(profileName string) string {
	return lipgloss.NewStyle().Bold(true).PaddingLeft(2).Render("\n\nActive profile: " + profileName)
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
