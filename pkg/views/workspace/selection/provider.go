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

func selectProviderPrompt(gitProviders []gitprovider_view.GitProviderView, additionalProjectOrder int, selectedReposGitProviders map[string]bool, choiceChan chan<- string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, provider := range gitProviders {
		id := provider.Id
		title := provider.Name
		isDisabled := false

		// additionalProjectOrder > 1 indicates use of 'multi-project' command
		if additionalProjectOrder > 1 && len(selectedReposGitProviders) > 0 && selectedReposGitProviders[id] {
			title += statusMessageDangerStyle(" (All repositories under this are already selected)")
			// isDisabled property helps in skipping over this specific git provider option, refer
			// handling of up/down key press under update method in ./view.go file
			isDisabled = true
		}

		newItem := item[string]{id: id, title: title, choiceProperty: id, isDisabled: isDisabled}
		items = append(items, newItem)
	}

	newItem := item[string]{id: CustomRepoIdentifier, title: "Enter a custom repository URL", choiceProperty: CustomRepoIdentifier}
	items = append(items, newItem)

	l := views.GetStyledSelectList(items)

	title := "Choose a Provider"
	if additionalProjectOrder > 1 {
		title += fmt.Sprintf(" (Project #%d)", additionalProjectOrder)
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

func GetProviderIdFromPrompt(gitProviders []gitprovider_view.GitProviderView, additionalProjectOrder int, selectedReposGitProviders map[string]bool) string {
	choiceChan := make(chan string)

	go selectProviderPrompt(gitProviders, additionalProjectOrder, selectedReposGitProviders, choiceChan)

	return <-choiceChan
}
