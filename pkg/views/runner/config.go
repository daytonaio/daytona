// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/runner"
	"github.com/daytonaio/daytona/pkg/views"
)

func RenderConfig(config *runner.Config, showKey bool) {
	output := views.GetStyledMainTitle("Daytona Runner Config") + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Runner ID: "), config.Id) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Runner Name: "), config.Name) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("API URL: "), config.ServerApiUrl) + "\n\n"

	if showKey {
		output += fmt.Sprintf("%s %s", views.GetPropertyKey("API Key: "), config.ServerApiKey) + "\n\n"
	}

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Providers Dir: "), config.ProvidersDir) + "\n\n"

	output += fmt.Sprintf("%s %s", views.GetPropertyKey("Log File Path: "), config.LogFile.Path) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("Log File Max Size: "), config.LogFile.MaxSize) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("Log File Max Backups: "), config.LogFile.MaxBackups) + "\n\n"

	output += fmt.Sprintf("%s %d", views.GetPropertyKey("Log File Max Age: "), config.LogFile.MaxAge) + "\n\n"

	output += fmt.Sprintf("%s %t", views.GetPropertyKey("Log File Local Time: "), config.LogFile.LocalTime) + "\n\n"

	output += fmt.Sprintf("%s %t", views.GetPropertyKey("Log File Compress: "), config.LogFile.Compress) + "\n\n"

	output += views.SeparatorString + "\n\n"

	output += fmt.Sprintf("To edit these values run: %s", lipgloss.NewStyle().Foreground(views.Green).Render("daytona runner configure")) + "\n\n"

	output += views.SeparatorString

	views.RenderContainerLayout(views.GetInfoMessage(output))
}
