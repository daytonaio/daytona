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

func selectPullRequestPrompt(pullRequests []serverapiclient.GitPullRequest, additionalProjectOrder int, choiceChan chan<- string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, pr := range pullRequests {
		newItem := item[string]{id: *pr.Name, title: *pr.Name, choiceProperty: *pr.Name}
		if *pr.Branch != "" {
			newItem.desc = fmt.Sprintf("Branch: %s", *pr.Branch)
		}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)

	title := "Choose a Pull/Merge Request"
	if additionalProjectOrder > 0 {
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

func GetPullRequestFromPrompt(pullRequests []serverapiclient.GitPullRequest, additionalProjectOrder int) *serverapiclient.GitPullRequest {
	choiceChan := make(chan string)

	go selectPullRequestPrompt(pullRequests, additionalProjectOrder, choiceChan)

	pullRequestName := <-choiceChan

	for _, pr := range pullRequests {
		if *pr.Name == pullRequestName {
			return &pr
		}
	}
	return nil
}
