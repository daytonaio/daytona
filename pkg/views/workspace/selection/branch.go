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

func selectBranchPrompt(branches []serverapiclient.GitBranch, secondaryProjectOrder int, choiceChan chan<- string) {
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

func GetBranchNameFromPrompt(branches []serverapiclient.GitBranch, secondaryProjectOrder int) string {
	choiceChan := make(chan string)

	go selectBranchPrompt(branches, secondaryProjectOrder, choiceChan)

	return <-choiceChan
}
