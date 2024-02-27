// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
)

func RenderMainTitle(title string) {
	fmt.Println(lipgloss.NewStyle().Foreground(views.Green).Bold(true).Padding(1, 0, 1, 0).Render(title))
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
