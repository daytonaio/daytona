// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"
	"strconv"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func selectRepositoryPrompt(repositories []apiclient.GitRepository, index int, choiceChan chan<- string, selectedRepos map[string]int) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, repository := range repositories {
		url := *repository.Url

		// Index > 1 indicates use of 'multi-project' command
		// Append occurence number so as to keep unique names for duplicate entries.
		if index > 1 && len(selectedRepos) > 0 && selectedRepos[url] > 0 {
			*repository.Name += strconv.Itoa(selectedRepos[url] + 1)
		}

		newItem := item[string]{id: url, title: *repository.Name, choiceProperty: url, desc: url}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)

	title := "Choose a Repository"
	if index > 1 {
		title += fmt.Sprintf(" (Project #%d)", index)
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
		choice := *m.choice

		selectedRepos[choice]++

		choiceChan <- choice
	} else {
		choiceChan <- ""
	}
}

func GetRepositoryFromPrompt(repositories []apiclient.GitRepository, index int, selectedRepos map[string]int) *apiclient.GitRepository {
	choiceChan := make(chan string)

	go selectRepositoryPrompt(repositories, index, choiceChan, selectedRepos)

	choice := <-choiceChan

	for _, repository := range repositories {
		if *repository.Url == choice {
			return &repository
		}
	}

	return nil
}
