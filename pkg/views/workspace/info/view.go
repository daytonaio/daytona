// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package info

// A simple program that counts down from 5 and then exits.

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var colors = views.ColorGrid(5, 5)

var workspaceNameStyle = lipgloss.NewStyle().
	Foreground(views.Green).
	Bold(true).
	MarginLeft(2)

var repositoryURLStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("227")).
	MarginLeft(2)

var repositoryBranchStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("227")).
	MarginLeft(2)

var viewStyle = lipgloss.NewStyle().
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

func projectRender(projectInfo *serverapiclient.ProjectInfo, project *serverapiclient.Project) string {
	projectState := ""
	extensions := [][]string{}
	extensionsTable := ""
	repoBranch := ""

	if !*projectInfo.IsRunning && *projectInfo.Created == "" {
		projectState = projectStatusStyle.Foreground(lipgloss.Color(colors[0][4])).Render("Unavailable")
	} else if !*projectInfo.IsRunning {
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

	if (project.Repository.Branch == nil){
		repoBranch = ""
	} else {
		repoBranch = "Branch" + repositoryBranchStyle.Render(*project.Repository.Branch)
	}

	repoView := "Url" + (repositoryURLStyle.Render(*project.Repository.Url)) + "\n" + repoBranch
	repoView = "Repository" + viewStyle.Render(repoView)
	
	projectView := "Project" + projectNameStyle.Render(*projectInfo.Name) + "\n" + "State  " + projectState + extensionsTable

	return viewStyle.Render(projectView) + viewStyle.Render(repoView)
}

func Render(workspace *serverapiclient.Workspace) {
	var output string
	output = "\n"
	output += "Workspace" + workspaceNameStyle.Render(*workspace.Info.Name) + "\n"
	output += "ID" + workspaceNameStyle.Render(*workspace.Id) + "\n"
	output += "Target" + workspaceNameStyle.Render(*workspace.Target) + "\n"

	if len(workspace.Projects) > 1 {
		output += "\n" + "Projects"
	}

	for _, project := range workspace.Projects { 
		for _, projectInfo := range workspace.Info.Projects {
			output += projectRender(&projectInfo, &project)
		}
	}

	output = lipgloss.NewStyle().PaddingLeft(3).Render(output)

	println(output)
	fmt.Println()
}
