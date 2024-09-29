// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

func selectNamespacePrompt(namespaces []apiclient.GitNamespace, projectOrder int, choiceChan chan<- string, navChan chan<- string, isPaginationDisabled bool, curPage, perPage int32) {
	items := []list.Item{}
	var desc string

	// Populate items with titles and descriptions from workspaces.
	for _, namespace := range namespaces {
		if namespace.Id == "<PERSONAL>" {
			desc = "personal"
		} else {
			desc = "organization"
		}
		newItem := item[string]{id: namespace.Id, title: namespace.Name, desc: desc, choiceProperty: namespace.Id}
		items = append(items, newItem)
	}

	if !isPaginationDisabled {
		items = AddNavigationOptionsToList(items, len(namespaces), curPage, perPage)
	}

	l := views.GetStyledSelectList(items, curPage)

	title := "Choose a Namespace"
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

func GetNamespaceIdFromPrompt(namespaces []apiclient.GitNamespace, projectOrder int, isPaginationDisabled bool, curPage, perPage int32) (string, string) {
	choiceChan := make(chan string)
	navChan := make(chan string)

	go selectNamespacePrompt(namespaces, projectOrder, choiceChan, navChan, isPaginationDisabled, curPage, perPage)

	select {
	case choice := <-choiceChan:
		return choice, ""
	case navigate := <-navChan:
		return "", navigate
	}
}
