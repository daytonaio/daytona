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
)

func selectRepositoryPrompt(repositories []apiclient.GitRepository, projectOrder int, choiceChan chan<- string, navChan chan<- string, selectedRepos map[string]int, parentIdentifier string, isPaginationDisabled bool, curPage, perPage int32) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, repository := range repositories {
		newItem := item[string]{
			id:             repository.Url,
			title:          repository.Name,
			choiceProperty: repository.Url,
			desc:           repository.Url,
		}
		items = append(items, newItem)
	}

	if !isPaginationDisabled {
		items = AddLoadMoreOptionToList(items, len(repositories), curPage, perPage)
	}

	l := views.GetStyledSelectList(items, parentIdentifier)

	title := "Choose a Repository"
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

	// Return either the choice or navigation
	if m, ok := p.(model[string]); ok && m.choice != nil {
		choice := *m.choice
		if choice == views.ListNavigationText {
			navChan <- choice
		} else {
			selectedRepos[choice]++
			choiceChan <- choice
		}
	} else {
		choiceChan <- ""
	}
}

func GetRepositoryFromPrompt(repositories []apiclient.GitRepository, projectOrder int, selectedRepos map[string]int, parentIdentifier string, isPaginationDisabled bool, curPage, perPage int32) (*apiclient.GitRepository, string) {
	choiceChan := make(chan string)
	navChan := make(chan string)

	go selectRepositoryPrompt(repositories, projectOrder, choiceChan, navChan, selectedRepos, parentIdentifier, isPaginationDisabled, curPage, perPage)

	select {
	case choice := <-choiceChan:
		for _, repository := range repositories {
			if repository.Url == choice {
				return &repository, ""
			}
		}
		return nil, ""
	case navigate := <-navChan:
		return nil, navigate
	}
}
