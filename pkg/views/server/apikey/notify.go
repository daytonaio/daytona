// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
)

func Render(key, apiUrl string) {
	var output string

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Generated API key: "), key) + "\n\n"

	output += "You can connect to the Daytona Server from a client machine by running:"

	views.RenderContainerLayout(views.GetInfoMessage(output))

	command := fmt.Sprintf("daytona profile add -a %s -k %s", apiUrl, key)
	fmt.Println(lipgloss.NewStyle().Padding(0).Foreground(views.Green).Render(command) + "\n\n")

	if err := clipboard.WriteAll(command); err == nil {
		output = "The command has been copied to your clipboard."
	} else {
		output = "Make sure to copy it as you will not be able to see it again."
	}

	views.RenderContainerLayout(views.GetInfoMessage(output))
}
