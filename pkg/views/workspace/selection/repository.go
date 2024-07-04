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

const redColor = "\033[31m"

func selectRepositoryPrompt(repositories []apiclient.GitRepository, index int, choiceChan chan<- string, selectedRepos map[string]bool) {
	isDuplicateEntry := false
	duplicateEntryErrorMessage := "DUPLICATE ENTRY detected, Retry with different option.\n"
	for {
		items := []list.Item{}

		// Populate items with titles and descriptions from workspaces.
		for _, repository := range repositories {
			title := *repository.Name
			// Index > 1 indicates 'multi-project' use
			if index > 1 && len(selectedRepos) > 0 && selectedRepos[*repository.Url] {
				// Hack to set red as font color
				title += fmt.Sprintf(" %s(ALREADY SELECTED)", redColor)
			}
			newItem := item[string]{id: *repository.Url, title: title, choiceProperty: *repository.Url, desc: *repository.Url}
			items = append(items, newItem)
		}

		l := views.GetStyledSelectList(items)

		title := ""
		if isDuplicateEntry {
			title = duplicateEntryErrorMessage
		}

		title += "Choose a Repository"
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
			if index > 1 && len(selectedRepos) > 0 && selectedRepos[choice] {
				// Duplicate entry, loop continues until user selects some other repo option
				isDuplicateEntry = true
			} else {
				selectedRepos[choice] = true
				choiceChan <- choice
				break
			}
		} else {
			choiceChan <- ""
			break
		}
	}
}

func GetRepositoryFromPrompt(repositories []apiclient.GitRepository, index int, selectedRepos map[string]bool) *apiclient.GitRepository {
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
