// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func selectWorkspacePrompt(workspaces []serverapiclient.WorkspaceDTO, actionVerb string, choiceChan chan<- *serverapiclient.WorkspaceDTO) {
	// Initialize an empty list of items.
	items := []list.Item{}

	// Populate items with titles and descriptions from workspaces.
	for _, workspace := range workspaces {
		var projectsInfo []string

		if workspace.Projects == nil || len(workspace.Projects) == 0 {
			continue
		}

		if len(workspace.Projects) == 1 {
			if workspace.Projects[0].Repository != nil && workspace.Projects[0].Repository.Url != nil {
				projectsInfo = append(projectsInfo, util.GetRepositorySlugFromUrl(*workspace.Projects[0].Repository.Url, true))
			}
		} else {
			for _, project := range workspace.Projects {
				projectsInfo = append(projectsInfo, *project.Name)
			}
		}

		// Get the time if available
		uptime := ""
		createdTime := ""
		if workspace.Info != nil && workspace.Info.Projects != nil && len(workspace.Info.Projects) > 0 && workspace.Info.Projects[0].Created != nil {
			createdTime = util.FormatCreatedTime(*workspace.Info.Projects[0].Created)
		}
		if len(workspace.Projects) > 0 && workspace.Projects[0].State != nil && workspace.Projects[0].State.Uptime != nil {
			uptime = fmt.Sprintf("up %s", util.FormatUptime(*workspace.Projects[0].State.Uptime))
		}

		newItem := item[serverapiclient.WorkspaceDTO]{
			title:          *workspace.Name,
			id:             *workspace.Id,
			desc:           strings.Join(projectsInfo, ", "),
			createdTime:    createdTime,
			uptime:         uptime,
			target:         *workspace.Target,
			choiceProperty: workspace,
		}

		items = append(items, newItem)
	}

	d := ItemDelegate[serverapiclient.WorkspaceDTO]{}

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	m := model[serverapiclient.WorkspaceDTO]{list: l}

	m.list.Title = views.GetStyledMainTitle("Select a Workspace To " + actionVerb)
	m.list.Styles.Title = lipgloss.NewStyle().Foreground(views.Green).Bold(true)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[serverapiclient.WorkspaceDTO]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}

func GetWorkspaceFromPrompt(workspaces []serverapiclient.WorkspaceDTO, actionVerb string) *serverapiclient.WorkspaceDTO {
	choiceChan := make(chan *serverapiclient.WorkspaceDTO)

	go selectWorkspacePrompt(workspaces, actionVerb, choiceChan)

	return <-choiceChan
}
