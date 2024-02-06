// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views_util

import (
	"fmt"

	"github.com/daytonaio/daytona/cmd/views"

	"github.com/charmbracelet/lipgloss"
)

func RenderMainTitle(title string) {
	fmt.Println(lipgloss.NewStyle().Foreground(views.Green).Bold(true).Padding(1, 0, 1, 0).Render(title))
}

func RenderInfoMessage(message string) {
	fmt.Println(lipgloss.NewStyle().Padding(1, 0, 1, 1).Render(message))
}

func RenderInfoMessageBold(message string) {
	fmt.Println(lipgloss.NewStyle().Bold(true).Padding(1, 0, 1, 1).Render(message))
}
