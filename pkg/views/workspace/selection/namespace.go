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

func selectNamespacePrompt(namespaces []apiclient.GitNamespace, additionalProjectOrder int, choiceChan chan<- string, providerId string, disabledGitProviders map[string]bool, disabledNamespaces map[string]bool) {
	items := []list.Item{}
	var desc string
	disabledNamespacesCount := 0

	// Populate items with titles and descriptions from workspaces.
	for _, namespace := range namespaces {
		isDisabled := false
		id := *namespace.Id
		title := *namespace.Name

		// additionalProjectOrder > 1 indicates use of 'multi-project' command
		if additionalProjectOrder > 1 && len(disabledNamespaces) > 0 && disabledNamespaces[id] {
			title += statusMessageDangerStyle(" (All repositories under this are already selected)")
			// isDisabled property helps in skipping over this specific namespace option, refer to
			// handling of up/down key press under update method in ./view.go file
			isDisabled = true
			disabledNamespacesCount++
		}

		if id == "<PERSONAL>" {
			desc = "personal"
		} else {
			desc = "organization"
		}
		newItem := item[string]{id: id, title: title, desc: desc, choiceProperty: id, isDisabled: isDisabled}
		items = append(items, newItem)
	}

	if disabledNamespacesCount == len(namespaces) {
		disabledGitProviders[providerId] = true
	}

	l := views.GetStyledSelectList(items)

	title := "Choose a Namespace"
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

func GetNamespaceIdFromPrompt(namespaces []apiclient.GitNamespace, additionalProjectOrder int, providerId string, disabledGitProviders map[string]bool, disabledNamespaces map[string]bool) string {
	choiceChan := make(chan string)

	go selectNamespacePrompt(namespaces, additionalProjectOrder, choiceChan, providerId, disabledGitProviders, disabledNamespaces)

	return <-choiceChan
}
