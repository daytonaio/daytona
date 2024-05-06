// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func selectRepositoryPrompt(repositories []serverapiclient.GitRepository, secondaryProjectOrder int, choiceChan chan<- string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, repository := range repositories {
		newItem := item[string]{id: *repository.Url, title: *repository.Name, choiceProperty: *repository.Url, desc: *repository.Url}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)

	title := "Choose a Repository"
	if secondaryProjectOrder > 0 {
		title += fmt.Sprintf(" (Secondary Project #%d)", secondaryProjectOrder)
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

func GetRepositoryFromPrompt(repositories []serverapiclient.GitRepository, secondaryProjectOrder int) serverapiclient.GitRepository {
	choiceChan := make(chan string)

	go selectRepositoryPrompt(repositories, secondaryProjectOrder, choiceChan)

	choice := <-choiceChan

	for _, repository := range repositories {
		if *repository.Url == choice {
			return repository
		}
	}

	return serverapiclient.GitRepository{}
}
