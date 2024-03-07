// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func selectProviderPrompt(gitProviders []config.GitProvider, secondaryProjectOrder int, choiceChan chan<- string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, provider := range gitProviders {
		newItem := item[string]{id: provider.Id, title: provider.Name, choiceProperty: provider.Id}
		items = append(items, newItem)
	}

	newItem := item[string]{id: CustomRepoIdentifier, title: "Enter a custom repository URL", choiceProperty: CustomRepoIdentifier}
	items = append(items, newItem)

	l := views.GetStyledSelectList(items)
	m := model[string]{list: l}
	m.list.Title = "CHOOSE YOUR PROVIDER"
	if secondaryProjectOrder > 0 {
		m.list.Title += fmt.Sprintf(" (Secondary Project #%d)", secondaryProjectOrder)
	}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[string]); ok && m.choice != nil {
		choiceChan <- *m.choice
	} else {
		choiceChan <- ""
	}
}

func GetProviderIdFromPrompt(gitProviders []config.GitProvider, secondaryProjectOrder int) string {
	choiceChan := make(chan string)

	go selectProviderPrompt(gitProviders, secondaryProjectOrder, choiceChan)

	return <-choiceChan
}
