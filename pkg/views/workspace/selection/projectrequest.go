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

func selectProjectRequestPrompt(projects []serverapiclient.CreateWorkspaceRequestProject, choiceChan chan<- string) {
	items := []list.Item{}

	for _, project := range projects {
		var name string
		if project.Name != "" {
			name = project.Name
		}
		var image string
		if project.Image != nil {
			image = *project.Image
		}
		var user string
		if project.User != nil {
			user = *project.User
		}
		newItem := item[string]{id: name, title: name, choiceProperty: name}
		if image != "" {
			if user != "" {
				newItem.desc = fmt.Sprintf("%s (%s)", image, user)
			} else {
				newItem.desc = fmt.Sprintf("Image: %s", image)
			}
		} else if user != "" {
			newItem.desc = fmt.Sprintf("User: %s", user)
		} else {
			newItem.desc = "Default configuration"
		}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)
	m := model[string]{list: l}
	m.list.Title = "CHOOSE A PROJECT TO CONFIGURE"

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

func GetProjectRequestFromPrompt(projects []serverapiclient.CreateWorkspaceRequestProject) string {
	choiceChan := make(chan string)

	go selectProjectRequestPrompt(projects, choiceChan)

	return <-choiceChan
}
