// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package select_prompt

import (
	"fmt"
	"os"
	"strings"

	"github.com/daytonaio/daytona/cli/cmd/views"
	"github.com/daytonaio/daytona/common/grpc/proto/types"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func selectWorkspacePrompt(workspaces []*types.WorkspaceInfo, actionVerb string, choiceChan chan<- string) {
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, workspace := range workspaces {
		var projectNames []string
		for _, project := range workspace.Projects {
			projectNames = append(projectNames, project.Name)
		}
		newItem := item{id: workspace.Name, title: workspace.Name, choiceProperty: workspace.Name, desc: "Projects: " + strings.Join(projectNames, ", ")}
		items = append(items, newItem)
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = "SELECT A WORKSPACE TO " + strings.ToUpper(actionVerb)

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

func GetWorkspaceNameFromPrompt(workspaces []*types.WorkspaceInfo, actionVerb string) string {
	choiceChan := make(chan string)

	go selectWorkspacePrompt(workspaces, actionVerb, choiceChan)

	return <-choiceChan
}
