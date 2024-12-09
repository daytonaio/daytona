// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
)

func Notify(apiUrl, key, id string) {
	var output string

	output += "You can connect the Runner to the Daytona Server by running this command on Runner's machine:"

	views.RenderContainerLayout(views.GetInfoMessage(output))

	command := fmt.Sprintf("daytona runner configure --api-url %s --api-key %s --id %s", apiUrl, key, id)
	fmt.Println(lipgloss.NewStyle().Padding(0).Foreground(views.Green).Render(command))

	if err := clipboard.WriteAll(command); err == nil {
		output = "The command has been copied to your clipboard."
	} else {
		output = "Make sure to copy it as you will not be able to see it again."
	}

	views.RenderContainerLayout(views.GetInfoMessage(output))
}
