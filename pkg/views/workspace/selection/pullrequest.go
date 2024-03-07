// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func selectPullRequestPrompt(pullRequests []gitprovider.GitPullRequest, secondaryProjectOrder int, choiceChan chan<- string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, pr := range pullRequests {
		newItem := item[string]{id: pr.Name, title: pr.Name, choiceProperty: pr.Name}
		if pr.Branch != "" {
			newItem.desc = fmt.Sprintf("Branch: %s", pr.Branch)
		}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)
	m := model[string]{list: l}
	m.list.Title = "CHOOSE A PULL/MERGE REQUEST"
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

func GetPullRequestFromPrompt(pullRequests []gitprovider.GitPullRequest, secondaryProjectOrder int) gitprovider.GitPullRequest {
	choiceChan := make(chan string)

	go selectPullRequestPrompt(pullRequests, secondaryProjectOrder, choiceChan)

	pullRequestName := <-choiceChan

	for _, pr := range pullRequests {
		if pr.Name == pullRequestName {
			return pr
		}
	}
	return gitprovider.GitPullRequest{}
}
