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

func selectBranchPrompt(branches []gitprovider.GitBranch, secondaryProjectOrder int, choiceChan chan<- string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, branch := range branches {
		newItem := item[string]{id: branch.Name, title: branch.Name, choiceProperty: branch.Name}
		if branch.SHA != "" {
			newItem.desc = fmt.Sprintf("SHA: %s", branch.SHA)
		}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)
	m := model[string]{list: l}
	m.list.Title = "CHOOSE A BRANCH"
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

func GetBranchNameFromPrompt(branches []gitprovider.GitBranch, secondaryProjectOrder int) string {
	choiceChan := make(chan string)

	go selectBranchPrompt(branches, secondaryProjectOrder, choiceChan)

	return <-choiceChan
}
