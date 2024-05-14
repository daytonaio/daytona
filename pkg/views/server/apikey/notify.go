// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
)

func Render(key, apiUrl string) {
	var output string

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Generated API key: "), key) + "\n\n"

	output += "Make sure to copy it as you will not be able to see it again." + "\n\n"

	output += views.SeparatorString + "\n\n"

	output += "You can connect to the Daytona Server from a client machine by running:\n\n"

	output += lipgloss.NewStyle().Foreground(views.Green).Render(fmt.Sprintf("daytona profile add -a %s -k %s", apiUrl, key))

	views.RenderContainerLayout(views.GetInfoMessage(output))
}
