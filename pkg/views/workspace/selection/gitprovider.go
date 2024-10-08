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

func selectGitProviderPrompt(gitProviders []gitprovider_view.GitProviderView, projectOrder int, choiceChan chan<- string, samplesEnabled bool) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, provider := range gitProviders {
		newItem := item[string]{id: provider.Id, title: provider.Name, choiceProperty: provider.Id}
		items = append(items, newItem)
	}

	newItem := item[string]{id: CustomRepoIdentifier, title: "Enter a custom repository URL", choiceProperty: CustomRepoIdentifier}
	items = append(items, newItem)

	if samplesEnabled {
		newItem := item[string]{id: CREATE_FROM_SAMPLE, title: "Create from Sample", choiceProperty: CREATE_FROM_SAMPLE}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)

	title := "Choose a Git Provider"
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

func GetProviderIdFromPrompt(gitProviders []gitprovider_view.GitProviderView, projectOrder int, samplesEnabled bool) string {
	choiceChan := make(chan string)

	go selectGitProviderPrompt(gitProviders, projectOrder, choiceChan, samplesEnabled)

	return <-choiceChan
}
