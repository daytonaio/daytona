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

func GetPrebuildFromPrompt(prebuilds []apiclient.PrebuildDTO, actionVerb string) *apiclient.PrebuildDTO {
	choiceChan := make(chan *apiclient.PrebuildDTO)
	go selectPrebuildPrompt(prebuilds, actionVerb, choiceChan)
	return <-choiceChan
}

func selectPrebuildPrompt(prebuilds []apiclient.PrebuildDTO, actionVerb string, choiceChan chan<- *apiclient.PrebuildDTO) {
	items := []list.Item{}

	for _, pb := range prebuilds {
		title := fmt.Sprintf("%s (%s)", pb.ProjectConfigName, views.GetBranchNameLabel(pb.Branch))

		desc := pb.Id
		if pb.CommitInterval != nil {
			desc = fmt.Sprintf("%s (every %d commits)", desc, *pb.CommitInterval)
		}

		newItem := item[apiclient.PrebuildDTO]{title: title, desc: desc, choiceProperty: pb}
		items = append(items, newItem)
	}

	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(views.Green).
		Foreground(views.Green).
		Bold(true).
		Padding(0, 0, 0, 1)

	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Foreground(views.DimmedGreen)

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	title := "Select a Prebuild To " + actionVerb

	l.Title = views.GetStyledMainTitle(title)
	l.Styles.Title = titleStyle

	m := model[apiclient.PrebuildDTO]{list: l}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[apiclient.PrebuildDTO]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}
