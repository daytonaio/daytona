// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	list_view "github.com/daytonaio/daytona/pkg/views/workspace/list"
)

var NewWorkspaceIdentifier = "<NEW_WORKSPACE>"

type ActionVerb string

var StartActionVerb ActionVerb = "Start"
var StopActionVerb ActionVerb = "Stop"
var RestartActionVerb ActionVerb = "Restart"
var DeleteActionVerb ActionVerb = "Delete"

func generateWorkspaceList(workspaces []apiclient.WorkspaceDTO, isMultipleSelect bool, action ActionVerb) []list.Item {
	// Initialize an empty list of items.
	items := []list.Item{}
	enabledItems := []list.Item{}
	disabledItems := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, workspace := range workspaces {
		workspaceName := workspace.Name
		if workspace.Name == "" {
			workspaceName = "Unnamed Workspace"
		}

		stateLabel := views.GetStateLabel(workspace.State.Name)

		isDisabled := false
		switch action {
		case StartActionVerb:
			if workspace.State.Name == apiclient.ResourceStateNameStarted {
				isDisabled = true
			}
		case StopActionVerb, RestartActionVerb:
			if workspace.State.Name == apiclient.ResourceStateNameStopped {
				isDisabled = true
			}
		}

		if workspace.Metadata != nil {
			views_util.CheckAndAppendTimeLabel(&stateLabel, workspace.State, workspace.Metadata.Uptime)
		}

		newItem := item[apiclient.WorkspaceDTO]{
			title:          workspaceName,
			id:             workspace.Id,
			desc:           "",
			targetName:     workspace.Target.Name,
			repository:     workspace.Repository.Url,
			state:          stateLabel,
			workspace:      &workspace,
			choiceProperty: workspace,
			isDisabled:     isDisabled,
		}

		if isMultipleSelect {
			newItem.isMultipleSelect = true
			newItem.action = string(action)
		}

		if isDisabled {
			disabledItems = append(disabledItems, newItem)
		} else {
			enabledItems = append(enabledItems, newItem)
		}
	}

	items = append(items, enabledItems...)
	items = append(items, disabledItems...)

	return items
}

func getWorkspaceProgramEssentials(modelTitle string, actionVerb ActionVerb, workspaces []apiclient.WorkspaceDTO, footerText string, isMultipleSelect bool) tea.Model {

	items := generateWorkspaceList(workspaces, isMultipleSelect, actionVerb)

	d := ItemDelegate[apiclient.WorkspaceDTO]{}

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	m := model[apiclient.WorkspaceDTO]{list: l}

	m.list.Title = views.GetStyledMainTitle(modelTitle + string(actionVerb))
	m.list.Styles.Title = lipgloss.NewStyle().Foreground(views.Green).Bold(true)
	m.footer = footerText

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()

	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	return p
}

func selectWorkspacePrompt(workspaces []apiclient.WorkspaceDTO, actionVerb ActionVerb, choiceChan chan<- *apiclient.WorkspaceDTO) {
	list_view.SortWorkspaces(&workspaces)

	p := getWorkspaceProgramEssentials("Select a Workspace To ", actionVerb, workspaces, "", false)
	if m, ok := p.(model[apiclient.WorkspaceDTO]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}

func GetWorkspaceFromPrompt(workspaces []apiclient.WorkspaceDTO, actionVerb ActionVerb) *apiclient.WorkspaceDTO {
	choiceChan := make(chan *apiclient.WorkspaceDTO)

	go selectWorkspacePrompt(workspaces, actionVerb, choiceChan)

	return <-choiceChan
}

func selectWorkspacesFromPrompt(workspaces []apiclient.WorkspaceDTO, actionVerb ActionVerb, choiceChan chan<- []*apiclient.WorkspaceDTO) {
	list_view.SortWorkspaces(&workspaces)

	footerText := lipgloss.NewStyle().Bold(true).PaddingLeft(2).Render(fmt.Sprintf("\n\nPress 'x' to mark a workspace.\nPress 'enter' to %s the current/marked workspaces.", actionVerb))
	p := getWorkspaceProgramEssentials("Select Workspaces To ", actionVerb, workspaces, footerText, true)

	m, ok := p.(model[apiclient.WorkspaceDTO])
	if ok && m.choices != nil {
		choiceChan <- m.choices
	} else if ok && m.choice != nil {
		choiceChan <- []*apiclient.WorkspaceDTO{m.choice}
	} else {
		choiceChan <- nil
	}
}

func GetWorkspacesFromPrompt(workspaces []apiclient.WorkspaceDTO, actionVerb ActionVerb) []*apiclient.WorkspaceDTO {
	choiceChan := make(chan []*apiclient.WorkspaceDTO)

	go selectWorkspacesFromPrompt(workspaces, actionVerb, choiceChan)

	return <-choiceChan
}
