// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func GetWorkspaceFromPrompt(workspaces []apiclient.WorkspaceDTO, actionVerb string) *apiclient.WorkspaceDTO {
	choiceChan := make(chan *apiclient.WorkspaceDTO)
	go selectWorkspacePrompt(workspaces, actionVerb, choiceChan)
	return <-choiceChan
}

func selectWorkspacePrompt(workspaces []apiclient.WorkspaceDTO, actionVerb string, choiceChan chan<- *apiclient.WorkspaceDTO) {
	items := []list.Item{}

	for _, workspace := range workspaces {
		workspaceName := workspace.Name
		if workspace.Name == "" {
			workspaceName = "Unnamed Workspace"
		}

		newItem := item[apiclient.WorkspaceDTO]{title: workspaceName, desc: "", choiceProperty: workspace}
		items = append(items, newItem)
	}

	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(views.Green).
		Foreground(views.Green).
		Bold(true).
		Padding(0, 0, 0, 1)

	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Foreground(views.DimmedGreen)

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	m := model[apiclient.WorkspaceDTO]{list: l}

	m.list.Title = views.GetStyledMainTitle("Select a Workspace To " + actionVerb)
	m.list.Styles.Title = lipgloss.NewStyle().Foreground(views.Green).Bold(true)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[apiclient.WorkspaceDTO]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}
