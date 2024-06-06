// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/views"
)

func RenderConfig(config *server.Config) {
	apiUrl := util.GetFrpcApiUrl(config.Frps.Protocol, config.Id, config.Frps.Domain)

	output := views.GetStyledMainTitle("Daytona Server Config") + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Server ID: "), config.Id) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("API URL: "), apiUrl) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("API Port: "), config.ApiPort) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Default Project Image: "), config.DefaultProjectImage) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Default Project User: "), config.DefaultProjectUser) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Default Project Post Start Commands: "), config.DefaultProjectPostStartCommands) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("FRPS Domain: "), config.Frps.Domain) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("Headscale Port: "), config.HeadscalePort) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Binaries Path: "), config.BinariesPath) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Logs Path: "), config.LogFilePath) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("Build Registry Port: "), config.RegistryPort) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Builder Image: "), config.BuilderImage) + "\n\n"

	output += views.SeparatorString + "\n\n"

	output += fmt.Sprintf("To edit these values run: %s", lipgloss.NewStyle().Foreground(views.Green).Render("daytona server configure")) + "\n\n"

	output += views.SeparatorString + "\n\n"

	output += "If you want to connect to the server remotely:\n\n"

	output += "1. Create an API key on this machine: "
	output += lipgloss.NewStyle().Foreground(views.Green).Render("daytona api-key new") + "\n"
	output += "2. Add a profile on the client machine: \n\t"
	output += lipgloss.NewStyle().Foreground(views.Green).Render(fmt.Sprintf("daytona profile add -a %s -k API_KEY", apiUrl))

	views.RenderContainerLayout(views.GetInfoMessage(output))
}
