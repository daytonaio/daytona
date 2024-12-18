// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

func Notify(runner *apiclient.RegisterRunnerResultDTO, apiUrl string) {
	var output string

	output += fmt.Sprintf("You can connect the Runner %s to the Daytona Server by running this command on the Runner's machine:", runner.Name)

	views.RenderContainerLayout(views.GetInfoMessage(output))

	command := fmt.Sprintf("daytona runner configure --api-url %s --api-key %s --id %s", apiUrl, runner.ApiKey, runner.Id)
	fmt.Println(lipgloss.NewStyle().Padding(0).Foreground(views.Green).Render(command))

	if err := clipboard.WriteAll(command); err == nil {
		output = "The command has been copied to your clipboard."
	} else {
		output = "Make sure to copy it as you will not be able to see it again."
	}

	views.RenderContainerLayout(views.GetInfoMessage(output))
}
