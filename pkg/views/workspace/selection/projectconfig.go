// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var BlankProjectIdentifier = "<BLANK_PROJECT>"

func GetProjectConfigFromPrompt(projectConfigs []apiclient.ProjectConfig, projectOrder int, showBlankOption bool, actionVerb string) *apiclient.ProjectConfig {
	choiceChan := make(chan *apiclient.ProjectConfig)
	go selectProjectConfigPrompt(projectConfigs, projectOrder, showBlankOption, actionVerb, choiceChan)
	return <-choiceChan
}

func selectProjectConfigPrompt(projectConfigs []apiclient.ProjectConfig, projectOrder int, showBlankOption bool, actionVerb string, choiceChan chan<- *apiclient.ProjectConfig) {
	items := []list.Item{}

	if showBlankOption {
		newItem := item[apiclient.ProjectConfig]{title: "Make a blank project", desc: "(default project configuration)", choiceProperty: apiclient.ProjectConfig{
			Name: BlankProjectIdentifier,
		}}
		items = append(items, newItem)
	}

	for _, pc := range projectConfigs {
		projectConfigName := pc.Name
		if pc.Name == "" {
			projectConfigName = "Unnamed Project Config"
		}

		newItem := item[apiclient.ProjectConfig]{title: projectConfigName, desc: pc.Repository.Url, choiceProperty: pc}
		items = append(items, newItem)
	}

	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(views.Green).
		Foreground(views.Green).
		Bold(true).
		Padding(0, 0, 0, 1)

	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Copy().Foreground(views.DimmedGreen)

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	title := "Select a Project Config To " + actionVerb
	if projectOrder > 1 {
		title += fmt.Sprintf(" (Project #%d)", projectOrder)
	}
	l.Title = views.GetStyledMainTitle(title)
	l.Styles.Title = titleStyle

	m := model[apiclient.ProjectConfig]{list: l}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[apiclient.ProjectConfig]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}
