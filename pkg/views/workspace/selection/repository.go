// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/types"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func selectRepositoryPrompt(repositories []types.Repository, secondaryProjectOrder int, choiceChan chan<- string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, repository := range repositories {
		newItem := item[string]{id: repository.Url, title: repository.Name, choiceProperty: repository.Url, desc: repository.Url}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)
	m := model[string]{list: l}
	m.list.Title = "CHOOSE A REPOSITORY"
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

func GetRepositoryFromPrompt(repositories []types.Repository, secondaryProjectOrder int) types.Repository {
	choiceChan := make(chan string)

	go selectRepositoryPrompt(repositories, secondaryProjectOrder, choiceChan)

	choice := <-choiceChan

	for _, repository := range repositories {
		if repository.Url == choice {
			return repository
		}
	}

	return types.Repository{}
}
