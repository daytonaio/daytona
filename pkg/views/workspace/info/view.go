// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

const propertyNameWidth = 16

var propertyNameStyle = lipgloss.NewStyle().
	Foreground(views.LightGray)

var propertyValueStyle = lipgloss.NewStyle().
	Foreground(views.Light).
	Bold(true)

func Render(workspace *apiclient.WorkspaceDTO, ide string, forceUnstyled bool) {
	var isCreationView bool
	var output string
	nameLabel := "Name"

	if ide != "" {
		isCreationView = true
	}

	if isCreationView {
		nameLabel = "Worskpace"
	}

	output += "\n"
	output += getInfoLine(nameLabel, workspace.Name) + "\n"

	output += getInfoLine("ID", workspace.Id) + "\n"

	if isCreationView {
		output += getInfoLine("Editor", ide) + "\n"
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

	output = getSingleWorkspaceOutput(workspace, isCreationView)

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

func getSingleWorkspaceOutput(workspace *apiclient.WorkspaceDTO, isCreationView bool) string {
	var output string
	var repositoryUrl string

	repositoryUrl = workspace.Repository.Url
	repositoryUrl = strings.TrimPrefix(repositoryUrl, "https://")
	repositoryUrl = strings.TrimPrefix(repositoryUrl, "http://")

	if workspace.State != nil {
		output += getInfoLineState("State", workspace.State) + "\n"
		output += getInfoLineGitStatus("Branch", &workspace.State.GitStatus) + "\n"
	}

	output += getInfoLinePrNumber(workspace.Repository.PrNumber, workspace.Repository, workspace.State)

	if !isCreationView {
		output += getInfoLine("Target ID", workspace.TargetId) + "\n"
	}
	output += getInfoLine("Repository", repositoryUrl)

	if !isCreationView {
		output += "\n"
		output += getInfoLine("Workspace", workspace.Name)
	}

	return output
}

func getInfoLine(key, value string) string {
	return propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key)) + propertyValueStyle.Render(value) + "\n"
}

func getInfoLineState(key string, state *apiclient.WorkspaceState) string {
	var uptime int
	var stateProperty string

	if state == nil {
		uptime = 0
	} else {
		uptime = int(state.Uptime)
	}

	if uptime == 0 {
		stateProperty = propertyValueStyle.Foreground(views.Gray).Render("STOPPED")
	} else {
		stateProperty = propertyValueStyle.Foreground(views.Green).Render("RUNNING")
	}

	return propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key)) + stateProperty + propertyValueStyle.Foreground(views.Light).Render("\n")
}

func getInfoLineGitStatus(key string, status *apiclient.GitStatus) string {
	output := propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key))
	output += propertyNameStyle.Foreground(views.Gray).Render(fmt.Sprintf("%-*s", propertyNameWidth, status.CurrentBranch))

	changesOutput := ""
	if status.FileStatus != nil {
		filesNum := len(status.FileStatus)
		if filesNum == 1 {
			changesOutput = " (1 uncommitted change)"
		} else if filesNum > 1 {
			changesOutput = fmt.Sprintf(" (%d uncommitted changes)", filesNum)
		}
	}

	unpushedOutput := ""

	if status.Ahead != nil && *status.Ahead > 0 {
		if *status.Ahead == 1 {
			unpushedOutput += " (1 commit ahead)"
		} else {
			unpushedOutput += fmt.Sprintf(" (%d commits ahead)", *status.Ahead)
		}
	}

	if status.Behind != nil && *status.Behind > 0 {
		if *status.Behind == 1 {
			unpushedOutput += " (1 commit behind)"
		} else {
			unpushedOutput += fmt.Sprintf(" (%d commits behind)", *status.Behind)
		}
	}

	branchPublishedOutput := ""
	if !*status.BranchPublished {
		branchPublishedOutput = " (branch not published)"
	}

	output += changesOutput + unpushedOutput + branchPublishedOutput + propertyValueStyle.Foreground(views.Light).Render("\n")

	return output
}

func getInfoLinePrNumber(PrNumber *int32, repo apiclient.GitRepository, state *apiclient.WorkspaceState) string {
	if PrNumber != nil && (state == nil || state.GitStatus.CurrentBranch == repo.Branch) {
		return getInfoLine("PR Number", fmt.Sprintf("#%d", *PrNumber)) + "\n"
	}
	return ""
}
