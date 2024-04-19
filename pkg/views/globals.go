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
	Padding(1, 2).
	BorderForeground(LightGray).
	Border(lipgloss.RoundedBorder())

var DefaultLayoutMarginTop = 1

var CheckmarkSymbol = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")

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

func GetListFooter(profileName string) string {
	return lipgloss.NewStyle().Bold(true).PaddingLeft(2).Render("\n\nActive profile: " + profileName)
}

func RenderBorderedMessage(message string) {
	fmt.Println(GetBorderedMessage(message))
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

func GetStyledMainTitle(content string) string {
	return lipgloss.NewStyle().Foreground(Black).Background(lipgloss.Color("#fff")).Padding(1, 2).Render(content)

	sidePadding := 2
	topBorder := ""
	bottomBorder := ""

	title := lipgloss.NewStyle().Foreground(Black).Bold(true).
		Background(lipgloss.Color("#fff")).Padding(0, sidePadding).Render(content)

	for i := 0; i < sidePadding+len(content)+sidePadding; i++ {
		topBorder = topBorder + "▔"
		// topBorder = topBorder + "▀"
		bottomBorder = bottomBorder + "▁"
		// bottomBorder = bottomBorder + "▂"
		// bottomBorder = bottomBorder + "▄"
	}

	title = title + "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#fff")).Render(topBorder)
	title = lipgloss.NewStyle().Foreground(lipgloss.Color("#fff")).Render(bottomBorder) + "\n" + title

	return title
}

func GetSeparatorString() string {
	return lipgloss.NewStyle().Foreground(LightGray).Render("===")
}

func RenderDefaultIdeUpdated(ide string) {

	content := fmt.Sprintf("%s %s", lipgloss.NewStyle().Foreground(LightGray).Render("Default IDE: "), ide)

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(content)
		return
	}

	fmt.Println(lipgloss.
		NewStyle().
		Margin(1, 0).
		Padding(1, 0, 1, 2).
		BorderForeground(LightGray).
		Border(lipgloss.RoundedBorder()).
		Width(terminalWidth - 10).
		Render(content))
}
