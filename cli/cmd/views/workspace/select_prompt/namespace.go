// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package select_prompt

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cli/cmd/views"
	"github.com/daytonaio/daytona/pkg/git_provider"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func selectNamespacePrompt(namespaces []git_provider.GitNamespace, secondaryProjectOrder int, choiceChan chan<- string) {
	items := []list.Item{}
	var desc string

	// Populate items with titles and descriptions from workspaces.
	for _, namespace := range namespaces {
		if namespace.Id == "<PERSONAL>" {
			desc = "personal"
		} else {
			desc = "organization"
		}
		newItem := item{id: namespace.Id, title: namespace.Name, desc: desc, choiceProperty: namespace.Id}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = "CHOOSE A NAMESPACE"
	if secondaryProjectOrder > 0 {
		m.list.Title += fmt.Sprintf(" (Secondary Project #%d)", secondaryProjectOrder)
	}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model); ok && m.choice != "" {
		choiceChan <- m.choice
	} else {
		choiceChan <- ""
	}
}

func GetNamespaceIdFromPrompt(namespaces []git_provider.GitNamespace, secondaryProjectOrder int) string {
	choiceChan := make(chan string)

	go selectNamespacePrompt(namespaces, secondaryProjectOrder, choiceChan)

	return <-choiceChan
}
