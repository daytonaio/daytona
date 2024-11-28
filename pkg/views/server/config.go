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

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("FRPS Domain: "), config.Frps.Domain) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("FRPS Port: "), config.Frps.Port) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("FRPS Protocol: "), config.Frps.Protocol) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("Headscale Port: "), config.HeadscalePort) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Binaries Path: "), config.BinariesPath) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Log File Path: "), config.LogFile.Path) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("Log File Max Size: "), config.LogFile.MaxSize) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("Log File Max Backups: "), config.LogFile.MaxBackups) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("Log File Max Age: "), config.LogFile.MaxAge) + "\n\n"

	output += fmt.Sprintf("%s %t", views.GetPropertyKey("Log File Local Time: "), config.LogFile.LocalTime) + "\n\n"

	output += fmt.Sprintf("%s %t", views.GetPropertyKey("Log File Compress: "), config.LogFile.Compress) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Builder Image: "), config.BuilderImage) + "\n\n"

	if config.BuilderRegistryServer == "local" {
		output += fmt.Sprintf("%s %d", views.GetPropertyKey("Local Builder Registry Port: "), config.LocalBuilderRegistryPort) + "\n\n"

		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Local Builder Registry Image: "), config.LocalBuilderRegistryImage) + "\n\n"
	} else {
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("Builder Registry: "), config.BuilderRegistryServer) + "\n\n"
	}

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Build Image Namespace: "), config.BuildImageNamespace) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Providers Dir: "), config.ProvidersDir) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Registry URL: "), config.RegistryUrl) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Server Download URL: "), config.ServerDownloadUrl) + "\n\n"

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
