// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
)

var DocStyle = lipgloss.NewStyle().Margin(3, 2, 1, 2)

func RenderMainTitle(title string) {
	fmt.Println(lipgloss.NewStyle().Foreground(views.Green).Bold(true).Padding(1, 0, 1, 0).Render(title))
}

func RenderLine(message string) {
	fmt.Println(lipgloss.NewStyle().PaddingLeft(1).Render(message))
}

func RenderInfoMessage(message string) {
	fmt.Println(lipgloss.NewStyle().Padding(1, 0, 1, 1).Render(message))
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
		Padding(1, 1, 1, 1).
		BorderForeground(views.Green).
		Border(lipgloss.RoundedBorder()).
		Render(message)
}

func GetStyledMainTitle(content string) string {
	return lipgloss.NewStyle().Foreground(views.Green).Bold(true).Render(content)
}
