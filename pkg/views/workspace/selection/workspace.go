// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func selectWorkspacePrompt(workspaces []serverapiclient.WorkspaceDTO, actionVerb string, choiceChan chan<- *serverapiclient.WorkspaceDTO) {
	// Initialize an empty list of items.
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, workspace := range workspaces {
		var projectNames []string
		for _, project := range workspace.Projects {
			projectNames = append(projectNames, *project.Name)
		}

		// Get the time if available
		statusTime := ""
		createdTime := ""
		if workspace.Info != nil && workspace.Info.Projects != nil && len(workspace.Info.Projects) > 0 && workspace.Info.Projects[0].Created != nil {
			createdTime = util.FormatCreatedTime(*workspace.Info.Projects[0].Created)
		}
		if workspace.Info != nil && workspace.Info.Projects != nil && len(workspace.Info.Projects) > 0 && workspace.Info.Projects[0].Started != nil {
			statusTime = util.FormatStatusTime(*workspace.Info.Projects[0].Started)
		}

		newItem := item[serverapiclient.WorkspaceDTO]{
			title:          *workspace.Name,
			id:             *workspace.Id,
			desc:           strings.Join(projectNames, ", "),
			createdTime:    createdTime,
			statusTime:     statusTime,
			target:         *workspace.Target,
			choiceProperty: workspace,
		}

		items = append(items, newItem)
	}

	d := ItemDelegate[serverapiclient.WorkspaceDTO]{}

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	m := model[serverapiclient.Workspace]{list: l}

	m.list.Title = "SELECT A WORKSPACE TO " + strings.ToUpper(actionVerb)
	m.list.Styles.Title = lipgloss.NewStyle().Foreground(views.Green).Bold(true)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[serverapiclient.WorkspaceDTO]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}

func GetWorkspaceFromPrompt(workspaces []serverapiclient.WorkspaceDTO, actionVerb string) *serverapiclient.WorkspaceDTO {
	choiceChan := make(chan *serverapiclient.WorkspaceDTO)

	go selectWorkspacePrompt(workspaces, actionVerb, choiceChan)

	return <-choiceChan
}
