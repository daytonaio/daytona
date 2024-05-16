// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

const propertyNameWidth = 16

var propertyNameStyle = lipgloss.NewStyle().
	Foreground(views.LightGray)

var propertyValueStyle = lipgloss.NewStyle().
	Foreground(views.Light).
	Bold(true)

func Render(workspace *serverapiclient.WorkspaceDTO, ide string, forceUnstyled bool) {
	var isCreationView bool
	var output string
	nameLabel := "Name"

	if ide != "" {
		isCreationView = true
	}

	if isCreationView {
		nameLabel = "Workspace"
	}

	output += "\n"
	output += getInfoLine(nameLabel, *workspace.Name) + "\n"

	output += getInfoLine("ID", *workspace.Id) + "\n"

	if isCreationView {
		output += getInfoLine("Editor", ide) + "\n"
	}

	if len(workspace.Projects) == 1 {
		output += getSingleProjectOutput(&workspace.Projects[0], isCreationView)
	} else {
		output += getProjectsOutputs(workspace.Projects, isCreationView)
	}

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(output)
		return
	}
	if terminalWidth < views.TUITableMinimumWidth || forceUnstyled {
		renderUnstyledInfo(output)
		return
	}

	if !isCreationView {
		output = views.GetStyledMainTitle("Workspace Info") + "\n" + output
	}

	renderTUIView(output, views.GetContainerBreakpointWidth(terminalWidth), isCreationView)
}

func renderUnstyledInfo(output string) {
	fmt.Println(output)
}

func renderTUIView(output string, width int, isCreationView bool) {
	output = lipgloss.NewStyle().PaddingLeft(3).Render(output)

	content := lipgloss.
		NewStyle().Width(width).
		Render(output)

	if !isCreationView {
		content = lipgloss.NewStyle().Margin(1, 0).Render(content)
	}

	fmt.Println(content)
}

func getSingleProjectOutput(project *serverapiclient.Project, isCreationView bool) string {
	var output string
	var repositoryUrl string

	if project.Repository != nil {
		repositoryUrl = *project.Repository.Url
		repositoryUrl = strings.TrimPrefix(repositoryUrl, "https://")
		repositoryUrl = strings.TrimPrefix(repositoryUrl, "http://")
	}

	if project.State != nil {
		output += getInfoLineState("State", project.State) + "\n"
		if project.State.GitStatus != nil {
			output += getInfoLineGitStatus("Branch", project.State.GitStatus) + "\n"
		}
	}

	if project.Target != nil && !isCreationView {
		output += getInfoLine("Target", *project.Target) + "\n"
	}
	output += getInfoLine("Repository", repositoryUrl)

	if project.Name != nil && !isCreationView {
		output += "\n"
		output += getInfoLine("Project", *project.Name)
	}

	return output
}

func getProjectsOutputs(projects []serverapiclient.Project, isCreationView bool) string {
	var output string
	for i, project := range projects {
		output += getInfoLine(fmt.Sprintf("Project #%d", i+1), *project.Name)
		output += getInfoLineState("State", project.State)
		output += getInfoLineGitStatus("Branch", project.State.GitStatus)
		if project.Target != nil && !isCreationView {
			output += getInfoLine("Target", *project.Target)
		}
		if project.Repository != nil {
			output += getInfoLine("Repository", *project.Repository.Url)
		}
		if project.Name != projects[len(projects)-1].Name {
			output += "\n"
		}
	}
	return output
}

func getInfoLine(key, value string) string {
	return propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key)) + propertyValueStyle.Render(value) + "\n"
}

func getInfoLineState(key string, state *serverapiclient.ProjectState) string {
	var uptime int
	var stateProperty string

	if state == nil || state.Uptime == nil {
		uptime = 0
	} else {
		uptime = int(*state.Uptime)
	}

	if uptime == 0 {
		stateProperty = propertyValueStyle.Foreground(views.Gray).Render("STOPPED")
	} else {
		stateProperty = propertyValueStyle.Foreground(views.Green).Render("RUNNING")
	}

	return propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key)) + stateProperty + propertyValueStyle.Foreground(views.Light).Render("\n")
}

func getInfoLineGitStatus(key string, status *serverapiclient.GitStatus) string {
	output := propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key))
	if status.CurrentBranch == nil {
		return output + propertyValueStyle.Foreground(views.Gray).Render("No branch") + "\n"
	}
	output += propertyNameStyle.Foreground(views.Gray).Render(fmt.Sprintf("%-*s", propertyNameWidth, *status.CurrentBranch))

	changesOutput := ""
	if status.FileStatus == nil {
		return output + "\n"
	}

	filesNum := len(status.FileStatus)
	if filesNum == 1 {
		changesOutput = " (" + fmt.Sprint(filesNum) + " uncommited change)"
	} else if filesNum > 1 {
		changesOutput = " (" + fmt.Sprint(filesNum) + " uncommited changes)"
	}
	output += changesOutput + propertyValueStyle.Foreground(views.Light).Render("\n")

	return output
}
