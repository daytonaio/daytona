// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
)

var titleStyle = lipgloss.NewStyle()

func selectProviderPrompt(gitProviders []gitprovider_view.GitProviderView, projectOrder int, choiceChan chan<- map[string]string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, provider := range gitProviders {
		newItem := item[map[string]string]{id: provider.Id, title: provider.Name, desc: provider.TokenScopeIdentity, choiceProperty: map[string]string{"id": provider.Id, "idenitity": provider.TokenScopeIdentity}}
		items = append(items, newItem)
	}

	newItem := item[map[string]string]{id: CustomRepoIdentifier, title: "Enter a custom repository URL", choiceProperty: map[string]string{"id": CustomRepoIdentifier, "idenitity": ""}}
	items = append(items, newItem)

	l := views.GetStyledSelectList(items)

	title := "Choose a Provider"
	if projectOrder > 1 {
		title += fmt.Sprintf(" (Project #%d)", projectOrder)
	}
	l.Title = views.GetStyledMainTitle(title)
	l.Styles.Title = titleStyle
	m := model[map[string]string]{list: l}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[map[string]string]); ok && m.choice != nil {
		choiceChan <- *m.choice
	} else {
		choiceChan <- map[string]string{"id": "", "idenitity": ""}
	}
}

func GetProviderIdFromPrompt(gitProviders []gitprovider_view.GitProviderView, projectOrder int) map[string]string {
	choiceChan := make(chan map[string]string)

	go selectProviderPrompt(gitProviders, projectOrder, choiceChan)

	return <-choiceChan
}
