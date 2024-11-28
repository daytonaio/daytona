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
var NewWorkspaceTemplateIdentifier = "<NEW_WORKSPACE_CONFIG>"

func GetWorkspaceTemplateFromPrompt(workspaceTemplates []apiclient.WorkspaceTemplate, workspaceOrder int, showBlankOption, withNewWorkspaceTemplate bool, actionVerb string) *apiclient.WorkspaceTemplate {
	choiceChan := make(chan *apiclient.WorkspaceTemplate)
	go selectWorkspaceTemplatePrompt(workspaceTemplates, workspaceOrder, showBlankOption, withNewWorkspaceTemplate, actionVerb, choiceChan)
	return <-choiceChan
}

func selectWorkspaceTemplatePrompt(workspaceTemplates []apiclient.WorkspaceTemplate, workspaceOrder int, showBlankOption, withNewWorkspaceTemplate bool, actionVerb string, choiceChan chan<- *apiclient.WorkspaceTemplate) {
	items := []list.Item{}

	if showBlankOption {
		newItem := item[apiclient.WorkspaceTemplate]{title: "Make a blank workspace", desc: "(default workspace configuration)", choiceProperty: apiclient.WorkspaceTemplate{
			Name: BlankWorkspaceIdentifier,
		}}
		items = append(items, newItem)
	}

	for _, wt := range workspaceTemplates {
		workspaceTemplateName := wt.Name
		if wt.Name == "" {
			workspaceTemplateName = "Unnamed Workspace Template"
		}

		newItem := item[apiclient.WorkspaceTemplate]{title: workspaceTemplateName, desc: wt.RepositoryUrl, choiceProperty: wt}
		items = append(items, newItem)
	}

	if withNewWorkspaceTemplate {
		newItem := item[apiclient.WorkspaceTemplate]{title: "+ Create a new workspace template", desc: "", choiceProperty: apiclient.WorkspaceTemplate{
			Name: NewWorkspaceTemplateIdentifier,
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

	title := "Select a Workspace Template To " + actionVerb
	if workspaceOrder > 1 {
		title += fmt.Sprintf(" (Workspace #%d)", workspaceOrder)
	}
	l.Title = views.GetStyledMainTitle(title)
	l.Styles.Title = titleStyle

	m := model[apiclient.WorkspaceTemplate]{list: l}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[apiclient.WorkspaceTemplate]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}
