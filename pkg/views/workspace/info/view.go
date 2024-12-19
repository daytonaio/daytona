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
	views_util "github.com/daytonaio/daytona/pkg/views/util"
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
		nameLabel = "Workspace"
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

	output += getWorkspaceDataOutput(workspace, isCreationView)

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

func getWorkspaceDataOutput(workspace *apiclient.WorkspaceDTO, isCreationView bool) string {
	var output string
	var repositoryUrl string

	repositoryUrl = workspace.Repository.Url
	repositoryUrl = strings.TrimPrefix(repositoryUrl, "https://")
	repositoryUrl = strings.TrimPrefix(repositoryUrl, "http://")

	output += getInfoLineState("State", workspace.State, workspace.Metadata) + "\n"
	if workspace.State.Error != nil {
		output += getInfoLine("Error", *workspace.State.Error) + "\n"
	}

	if workspace.Metadata != nil {
		output += getInfoLineGitStatus("Branch", &workspace.Metadata.GitStatus) + "\n"
	}

	output += getInfoLinePrNumber(workspace.Repository.PrNumber, workspace.Repository, workspace.Metadata)

	if !isCreationView {
		output += getInfoLine("Target", fmt.Sprintf("%s (%s)", workspace.Target.Name, workspace.TargetId)) + "\n"
	}

	output += getInfoLine("Repository", repositoryUrl)

	return output
}

func getInfoLine(key, value string) string {
	return propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key)) + propertyValueStyle.Render(value) + "\n"
}

func getInfoLineState(key string, state apiclient.ResourceState, metadata *apiclient.WorkspaceMetadata) string {
	stateLabel := views.GetStateLabel(state.Name)

	if metadata != nil {
		views_util.CheckAndAppendTimeLabel(&stateLabel, state, metadata.Uptime)
	}

	return propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key)) + stateLabel + propertyValueStyle.Foreground(views.Light).Render("\n")
}

func getInfoLineGitStatus(key string, status *apiclient.GitStatus) string {
	output := propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key))

	if status == nil {
		return output + propertyValueStyle.Foreground(views.Light).Render("\n")
	}

	currentBranch := status.CurrentBranch
	if status.CurrentBranch == "" {
		currentBranch = "/"
	}

	output += propertyNameStyle.Foreground(views.Gray).Render(currentBranch)

	detailsOutput := ""
	if status.FileStatus != nil {
		filesNum := len(status.FileStatus)
		if filesNum == 1 {
			detailsOutput = " (1 uncommitted change)"
		} else if filesNum > 1 {
			detailsOutput = fmt.Sprintf(" (%d uncommitted changes)", filesNum)
		}
	}

	if status.Ahead != nil && *status.Ahead > 0 {
		if *status.Ahead == 1 {
			detailsOutput += " (1 commit ahead)"
		} else {
			detailsOutput += fmt.Sprintf(" (%d commits ahead)", *status.Ahead)
		}
	}

	if status.Behind != nil && *status.Behind > 0 {
		if *status.Behind == 1 {
			detailsOutput += " (1 commit behind)"
		} else {
			detailsOutput += fmt.Sprintf(" (%d commits behind)", *status.Behind)
		}
	}

	if !*status.BranchPublished {
		detailsOutput += " (branch not published)"
	}

	output += detailsOutput + propertyValueStyle.Foreground(views.Light).Render("\n")

	return output
}

func getInfoLinePrNumber(PrNumber *int32, repo apiclient.GitRepository, metadata *apiclient.WorkspaceMetadata) string {
	if PrNumber != nil && (metadata == nil || metadata.GitStatus.CurrentBranch == repo.Branch) {
		return getInfoLine("PR Number", fmt.Sprintf("#%d", *PrNumber)) + "\n"
	}
	return ""
}
