// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"golang.org/x/term"
)

const propertyNameWidth = 16
const minTUIWidth = 80
const maxTUIWidth = 140

var propertyNameStyle = lipgloss.NewStyle().
	Foreground(views.LightGray)

var propertyValueStyle = lipgloss.NewStyle().
	Foreground(views.White).
	Bold(true)

func Render(workspace *serverapiclient.WorkspaceDTO, ide string) {
	var isCreationView bool
	nameLabel := "Name"

	if ide != "" {
		isCreationView = true
	}

	output := ""

	if !isCreationView {
		output += view_util.GetStyledMainTitle("Workspace info") + "\n"
	} else {
		nameLabel = "Workspace"
	}

	output += "\n"

	output += getInfoLine(nameLabel, *workspace.Info.Name) + "\n"

	if isCreationView {
		output += getInfoLine("Editor", ide) + "\n"
	} else {
		output += getInfoLine("ID", *workspace.Id) + "\n"
	}

	if len(workspace.Projects) == 1 {
		output += renderSingleProject(&workspace.Projects[0], isCreationView)
	} else {
		output += renderProjects(workspace.Projects, workspace.Info.Projects, isCreationView)
	}

	var width int
	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || terminalWidth < minTUIWidth {
		fmt.Println(output)
		return
	}
	width = terminalWidth - 20
	if width > maxTUIWidth {
		width = maxTUIWidth
	}
	renderTUIView(output, width, isCreationView)
}

func renderTUIView(output string, terminalWidth int, isCreationView bool) {
	output = lipgloss.NewStyle().PaddingLeft(3).Render(output)

	content := lipgloss.
		NewStyle().
		BorderForeground(views.LightGray).
		Border(lipgloss.RoundedBorder()).Width(terminalWidth).
		Render(output)

	if !isCreationView {
		content = lipgloss.NewStyle().Margin(1, 0).Render(content)
	}

	fmt.Println(content)
}

func renderSingleProject(project *serverapiclient.Project, isCreationView bool) string {
	var output string

	repositoryUrl := *project.Repository.Url
	repositoryUrl = strings.TrimPrefix(repositoryUrl, "https://")
	repositoryUrl = strings.TrimPrefix(repositoryUrl, "http://")

	if project.State != nil {
		output += getInfoLineState("State", strconv.Itoa(int(*project.State.Uptime))) + "\n"
	}
	if project.Target != nil && !isCreationView {
		output += getInfoLine("Target", *project.Target) + "\n"
	}
	output += getInfoLine("Repository", repositoryUrl) + "\n"

	if project.Name != nil && !isCreationView {
		output += getInfoLine("Project", *project.Name)
	}

	return output
}

func renderProjects(projects []serverapiclient.Project, projectInfos []serverapiclient.ProjectInfo, isCreationView bool) string {
	var output string
	for i, project := range projects {
		for _, projectInfo := range projectInfos {
			if *projectInfo.Name == *project.Name {
				output += getInfoLine(fmt.Sprintf("Project #%d", i+1), *project.Name)
				if project.State != nil {
					output += getInfoLineState("State", strconv.Itoa(int(*project.State.Uptime)))
				}
				if project.Target != nil && !isCreationView {
					output += getInfoLine("Target", *project.Target)
				}
				if project.Repository != nil {
					output += getInfoLine("Repository", *project.Repository.Url)
				}
				break
			}
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

func getInfoLineState(key, value string) string {
	var state string
	if value == "0" {
		state = propertyValueStyle.Foreground(views.Red).Render("STOPPED")
	} else {
		state = propertyValueStyle.Foreground(views.Green).Render("RUNNING")
	}
	return propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key)) + state + propertyValueStyle.Foreground(views.White).Render("\n")
}
