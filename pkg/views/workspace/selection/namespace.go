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

func selectNamespacePrompt(namespaces []apiclient.GitNamespace, projectOrder int, selectionListOptions views.SelectionListOptions, choiceChan chan<- string, navChan chan<- string) {
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

	if !selectionListOptions.IsPaginationDisabled {
		items = AddLoadMoreOptionToList(items)
	}

	l := views.GetStyledSelectList(items, selectionListOptions)

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
		if !selectionListOptions.IsPaginationDisabled && choice == views.ListNavigationText {
			navChan <- choice
		} else {
			choiceChan <- choice
		}
	} else {
		choiceChan <- ""
	}
}

func GetNamespaceIdFromPrompt(namespaces []apiclient.GitNamespace, projectOrder int, selectionListOptions views.SelectionListOptions) (string, string) {
	choiceChan := make(chan string)
	navChan := make(chan string)

	go selectNamespacePrompt(namespaces, projectOrder, selectionListOptions, choiceChan, navChan)

	select {
	case choice := <-choiceChan:
		return choice, ""
	case navigate := <-navChan:
		return "", navigate
	}
}
