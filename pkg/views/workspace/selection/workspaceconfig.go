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

var BlankWorkspaceIdentifier = "<BLANK_WORKSPACE>"
var NewWorkspaceConfigIdentifier = "<NEW_WORKSPACE_CONFIG>"

func GetWorkspaceConfigFromPrompt(workspaceConfigs []apiclient.WorkspaceConfig, workspaceOrder int, showBlankOption, withNewWorkspaceConfig bool, actionVerb string) *apiclient.WorkspaceConfig {
	choiceChan := make(chan *apiclient.WorkspaceConfig)
	go selectWorkspaceConfigPrompt(workspaceConfigs, workspaceOrder, showBlankOption, withNewWorkspaceConfig, actionVerb, choiceChan)
	return <-choiceChan
}

func selectWorkspaceConfigPrompt(workspaceConfigs []apiclient.WorkspaceConfig, workspaceOrder int, showBlankOption, withNewWorkspaceConfig bool, actionVerb string, choiceChan chan<- *apiclient.WorkspaceConfig) {
	items := []list.Item{}

	if showBlankOption {
		newItem := item[apiclient.WorkspaceConfig]{title: "Make a blank workspace", desc: "(default workspace configuration)", choiceProperty: apiclient.WorkspaceConfig{
			Name: BlankWorkspaceIdentifier,
		}}
		items = append(items, newItem)
	}

	for _, wc := range workspaceConfigs {
		workspaceConfigName := wc.Name
		if wc.Name == "" {
			workspaceConfigName = "Unnamed Workspace Config"
		}

		newItem := item[apiclient.WorkspaceConfig]{title: workspaceConfigName, desc: wc.RepositoryUrl, choiceProperty: wc}
		items = append(items, newItem)
	}

	if withNewWorkspaceConfig {
		newItem := item[apiclient.WorkspaceConfig]{title: "+ Create a new workspace configuration", desc: "", choiceProperty: apiclient.WorkspaceConfig{
			Name: NewWorkspaceConfigIdentifier,
		}}
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

	title := "Select a Workspace Config To " + actionVerb
	if workspaceOrder > 1 {
		title += fmt.Sprintf(" (Workspace #%d)", workspaceOrder)
	}
	l.Title = views.GetStyledMainTitle(title)
	l.Styles.Title = titleStyle

	m := model[apiclient.WorkspaceConfig]{list: l}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[apiclient.WorkspaceConfig]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}
