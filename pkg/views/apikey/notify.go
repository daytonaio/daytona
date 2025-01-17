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

	views.RenderContainerLayout(views.GetInfoMessage("You have successfully generated a new API Key and Server URL, which you can find below:"))

	fmt.Println(lipgloss.NewStyle().Padding(0).Render(fmt.Sprintf("%s %s", views.GetPropertyKey("DAYTONA_API_KEY="), key)))
	fmt.Println(lipgloss.NewStyle().Padding(0).Render(fmt.Sprintf("%s%s", views.GetPropertyKey("DAYTONA_SERVER_URL="), apiUrl)))

	output += "You can also connect to the Daytona Server instantly from a client machine by running the command below:"

	views.RenderContainerLayout(views.GetInfoMessage(output))

	command := fmt.Sprintf("daytona profile create -a %s -k %s", apiUrl, key)
	fmt.Println(lipgloss.NewStyle().Padding(0).Foreground(views.Green).Render(command))

	if err := clipboard.WriteAll(command); err == nil {
		output = "The command has been copied to your clipboard."
	} else {
		output = "Make sure to copy it as you will not be able to see it again."
	}

	views.RenderContainerLayout(views.GetInfoMessage(output))
}
