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

func selectProjectRequestPrompt(projects []serverapiclient.CreateWorkspaceRequestProject, defaultContainerUser string, choiceChan chan<- *serverapiclient.CreateWorkspaceRequestProject) {
	items := []list.Item{}

	for _, project := range projects {
		var name string
		var image string
		var user string

		if project.Name != "" {
			name = project.Name
		}
		if project.Image != nil {
			image = *project.Image
		}
		if project.User != nil {
			user = *project.User
		}
		if user == "" {
			user = "user not defined"
		}

		newItem := item[serverapiclient.CreateWorkspaceRequestProject]{id: name, desc: image, title: name, choiceProperty: project}
		if len(project.PostStartCommands) > 0 {
			newItem.desc = fmt.Sprintf("%s + %d post start command%s", newItem.desc, len(project.PostStartCommands), func() string {
				if len(project.PostStartCommands) == 1 {
					return ""
				} else {
					return "s"
				}
			}())
		}

		if user != defaultContainerUser {
			newItem.title += fmt.Sprintf(" (%s)", user)
		}

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

func GetProjectRequestFromPrompt(projects []serverapiclient.CreateWorkspaceRequestProject, defaultContainerUser string) *serverapiclient.CreateWorkspaceRequestProject {
	choiceChan := make(chan *serverapiclient.CreateWorkspaceRequestProject)

	go selectProjectRequestPrompt(projects, defaultContainerUser, choiceChan)

	return <-choiceChan
}
