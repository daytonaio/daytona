// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	list_view "github.com/daytonaio/daytona/pkg/views/target/list"
)

func generateTargetList(targets []apiclient.TargetDTO, isMultipleSelect bool, action string) []list.Item {

	// Initialize an empty list of items.
	items := []list.Item{}

	// Populate items with titles and descriptions from targets.
	for _, target := range targets {
		var workspacesInfo []string

		if len(target.Workspaces) == 0 {
			continue
		}

		if len(target.Workspaces) == 1 {
			workspacesInfo = append(workspacesInfo, util.GetRepositorySlugFromUrl(target.Workspaces[0].Repository.Url, true))
		} else {
			for _, workspace := range target.Workspaces {
				workspacesInfo = append(workspacesInfo, workspace.Name)
			}
		}

		// Get the time if available
		uptime := ""
		createdTime := ""
		if target.Info != nil && target.Info.Workspaces != nil && len(target.Info.Workspaces) > 0 {
			createdTime = util.FormatTimestamp(target.Info.Workspaces[0].Created)
		}
		if len(target.Workspaces) > 0 && target.Workspaces[0].State != nil {
			if target.Workspaces[0].State.Uptime == 0 {
				uptime = "STOPPED"
			} else {
				uptime = fmt.Sprintf("up %s", util.FormatUptime(target.Workspaces[0].State.Uptime))
			}
		}

		newItem := item[apiclient.TargetDTO]{
			title:          target.Name,
			id:             target.Id,
			desc:           strings.Join(workspacesInfo, ", "),
			createdTime:    createdTime,
			uptime:         uptime,
			targetConfig:   target.TargetConfig,
			choiceProperty: target,
		}

		if isMultipleSelect {
			newItem.isMultipleSelect = true
			newItem.action = action
		}

		items = append(items, newItem)
	}

	return items
}

func getTargetProgramEssentials(modelTitle string, actionVerb string, targets []apiclient.TargetDTO, footerText string, isMultipleSelect bool) tea.Model {

	items := generateTargetList(targets, isMultipleSelect, actionVerb)

	d := ItemDelegate[apiclient.TargetDTO]{}

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	m := model[apiclient.TargetDTO]{list: l}

	m.list.Title = views.GetStyledMainTitle(modelTitle + actionVerb)
	m.list.Styles.Title = lipgloss.NewStyle().Foreground(views.Green).Bold(true)
	m.footer = footerText

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()

	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	return p
}

func selectTargetPrompt(targets []apiclient.TargetDTO, actionVerb string, choiceChan chan<- *apiclient.TargetDTO) {
	list_view.SortTargets(&targets, true)

	p := getTargetProgramEssentials("Select a Target To ", actionVerb, targets, "", false)
	if m, ok := p.(model[apiclient.TargetDTO]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}

func GetTargetFromPrompt(targets []apiclient.TargetDTO, actionVerb string) *apiclient.TargetDTO {
	choiceChan := make(chan *apiclient.TargetDTO)

	go selectTargetPrompt(targets, actionVerb, choiceChan)

	return <-choiceChan
}

func selectTargetsFromPrompt(targets []apiclient.TargetDTO, actionVerb string, choiceChan chan<- []*apiclient.TargetDTO) {
	list_view.SortTargets(&targets, true)

	footerText := lipgloss.NewStyle().Bold(true).PaddingLeft(2).Render(fmt.Sprintf("\n\nPress 'x' to mark target.\nPress 'enter' to %s the current/marked targets.", actionVerb))
	p := getTargetProgramEssentials("Select Targets To ", actionVerb, targets, footerText, true)

	m, ok := p.(model[apiclient.TargetDTO])
	if ok && m.choices != nil {
		choiceChan <- m.choices
	} else if ok && m.choice != nil {
		choiceChan <- []*apiclient.TargetDTO{m.choice}
	} else {
		choiceChan <- nil
	}
}

func GetTargetsFromPrompt(targets []apiclient.TargetDTO, actionVerb string) []*apiclient.TargetDTO {
	choiceChan := make(chan []*apiclient.TargetDTO)

	go selectTargetsFromPrompt(targets, actionVerb, choiceChan)

	return <-choiceChan
}
