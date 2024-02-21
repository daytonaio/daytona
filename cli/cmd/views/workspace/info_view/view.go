// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package info_view_view

// A simple program that counts down from 5 and then exits.

import (
	"github.com/daytonaio/daytona/cli/cmd/views"

	"github.com/daytonaio/daytona/common/api_client"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var colors = views.ColorGrid(5, 5)

var workspaceInfoStyle = lipgloss.NewStyle()

var workspaceNameStyle = lipgloss.NewStyle().
	Foreground(views.Green).
	Bold(true).
	MarginLeft(2).
	MarginBottom(1)

var projectViewStyle = lipgloss.NewStyle().
	MarginTop(1).
	MarginBottom(0).
	PaddingLeft(2).
	PaddingRight(2)

var projectNameStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(views.Blue).
	PaddingLeft(2)

var projectStatusStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(colors[0][4])).
	PaddingLeft(2)

func projectRender(project *api_client.ProjectInfo) string {
	projectState := ""
	extensions := [][]string{}
	extensionsTable := ""

	if !*project.IsRunning && *project.Created == "" {
		projectState = projectStatusStyle.Foreground(lipgloss.Color(colors[0][4])).Render("Unavailable")
	} else if !*project.IsRunning {
		projectState = projectStatusStyle.Render("Stopped")
	} else {
		projectState = projectStatusStyle.Foreground(lipgloss.Color(colors[4][4])).Render("Running")
		// for _, extension := range project.Extensions {
		// 	extensions = append(extensions, []string{extension.Name /*extension.State*/, "", extension.Info})
		// }

		extensionsTable = table.New().
			Border(lipgloss.HiddenBorder()).
			Rows(extensions...).Render()
	}

	projectView := "Project" + projectNameStyle.Render(*project.Name) + "\n" + "State  " + projectState + "\n" + extensionsTable

	return projectViewStyle.Render(projectView)
}

func Render(wsInfo *api_client.WorkspaceInfo) {
	var output string
	output = "\n"
	output += workspaceInfoStyle.Render("Workspace" + workspaceNameStyle.Render(*wsInfo.Name))
	if len(wsInfo.Projects) > 1 {
		output += "\n" + "Projects"
	}
	for _, project := range wsInfo.Projects {
		output += projectRender(&project)
	}

	output = lipgloss.NewStyle().PaddingLeft(3).Render(output)

	println(output)
}
