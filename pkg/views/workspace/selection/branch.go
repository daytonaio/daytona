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

func selectBranchPrompt(branches []apiclient.GitBranch, projectOrder int, choiceChan chan<- string, navChan chan<- string, isPaginationDisabled bool, curPage, perPage int32) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, branch := range branches {
		newItem := item[string]{id: branch.Name, title: branch.Name, choiceProperty: branch.Name}
		if branch.Sha != "" {
			newItem.desc = fmt.Sprintf("SHA: %s", branch.Sha)
		}
		items = append(items, newItem)
	}

	if !isPaginationDisabled {
		items = AddNavigationOptionsToList(items, len(branches), curPage, perPage)
	}

	l := views.GetStyledSelectList(items, curPage)

	title := "Choose a Branch"
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
		if choice == "next" || choice == "prev" {
			navChan <- choice
		} else {
			choiceChan <- choice
		}
	} else {
		choiceChan <- ""
	}
}

func GetBranchFromPrompt(branches []apiclient.GitBranch, projectOrder int, isPaginationDisabled bool, curPage, perPage int32) (*apiclient.GitBranch, string) {
	choiceChan := make(chan string)
	navChan := make(chan string)

	go selectBranchPrompt(branches, projectOrder, choiceChan, navChan, isPaginationDisabled, curPage, perPage)

	select {
	case branchName := <-choiceChan:
		for _, b := range branches {
			if b.Name == branchName {
				return &b, ""
			}
		}
		return nil, ""
	case navigate := <-navChan:
		return nil, navigate
	}
}
