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

func selectBranchPrompt(branches []apiclient.GitBranch, additionalProjectOrder int, choiceChan chan<- string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, branch := range branches {
		newItem := item[string]{id: *branch.Name, title: *branch.Name, choiceProperty: *branch.Name}
		if *branch.Sha != "" {
			newItem.desc = fmt.Sprintf("SHA: %s", *branch.Sha)
		}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)

	title := "Choose a Branch"
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

func GetBranchFromPrompt(branches []apiclient.GitBranch, additionalProjectOrder int) *apiclient.GitBranch {
	choiceChan := make(chan string)

	go selectBranchPrompt(branches, additionalProjectOrder, choiceChan)

	branchName := <-choiceChan

	for _, b := range branches {
		if *b.Name == branchName {
			return &b
		}
	}

	return nil
}
