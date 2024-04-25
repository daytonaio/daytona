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

func selectProjectRequestPrompt(projects []serverapiclient.CreateWorkspaceRequestProject, choiceChan chan<- *serverapiclient.CreateWorkspaceRequestProject) {
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
		if user == "" {
			user = "user not defined"
		}
		newItem := item[serverapiclient.CreateWorkspaceRequestProject]{id: name, title: name, choiceProperty: project}
		newItem.desc = fmt.Sprintf("%s (%s)", image, user)

		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)
	m := model[serverapiclient.CreateWorkspaceRequestProject]{list: l}
	m.list.Title = "CHOOSE A PROJECT TO CONFIGURE"

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[serverapiclient.CreateWorkspaceRequestProject]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}

func GetProjectRequestFromPrompt(projects []serverapiclient.CreateWorkspaceRequestProject) *serverapiclient.CreateWorkspaceRequestProject {
	choiceChan := make(chan *serverapiclient.CreateWorkspaceRequestProject)

	go selectProjectRequestPrompt(projects, choiceChan)

	return <-choiceChan
}
