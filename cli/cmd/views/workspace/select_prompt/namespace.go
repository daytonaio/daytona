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

func (i item) getChoice() string { return i.id }

func selectNamespacePrompt(namespaces []git_provider.GitNamespace, choiceChan chan<- string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, namespace := range namespaces {
		newItem := item{title: namespace.Name}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = "CHOOSE A NAMESPACE"

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

func GetNamespaceIdFromPrompt(namespaces []git_provider.GitNamespace) string {
	choiceChan := make(chan string)

	go selectNamespacePrompt(namespaces, choiceChan)

	return <-choiceChan
}
