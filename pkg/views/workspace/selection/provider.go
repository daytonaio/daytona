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

func selectProviderPrompt(gitProviders []gitprovider_view.GitProviderView, projectOrder int, choiceChan chan<- string) {
	items := []list.Item{}

	for _, provider := range gitProviders {
		newItem := item[gitprovider_view.GitProviderView]{
			id:             provider.Id,
			title:          fmt.Sprintf("%s (%s)", provider.Name, provider.Identity),
			choiceProperty: provider,
		}
		items = append(items, newItem)
	}

	newItem := item[string]{id: CustomRepoIdentifier, title: "Enter a custom repository URL", choiceProperty: CustomRepoIdentifier}
	items = append(items, newItem)

	l := views.GetStyledSelectList(items)

	title := "Choose a Provider"
	if projectOrder > 1 {
		title += fmt.Sprintf(" (Project #%d)", projectOrder)
	}
	l.Title = views.GetStyledMainTitle(title)
	l.Styles.Title = titleStyle
	m := model[string]{list: l}

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

func GetProviderIdFromPrompt(gitProviders []gitprovider_view.GitProviderView, projectOrder int) string {
	choiceChan := make(chan string)

	go selectProviderPrompt(gitProviders, projectOrder, choiceChan)

	return <-choiceChan
}
