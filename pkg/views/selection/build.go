// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	list_view "github.com/daytonaio/daytona/pkg/views/build/list"
)

func GetBuildFromPrompt(builds []apiclient.BuildDTO, actionVerb string) *apiclient.BuildDTO {
	choiceChan := make(chan *apiclient.BuildDTO)
	go selectBuildPrompt(builds, actionVerb, choiceChan)
	return <-choiceChan
}

func selectBuildPrompt(builds []apiclient.BuildDTO, actionVerb string, choiceChan chan<- *apiclient.BuildDTO) {
	list_view.SortBuilds(&builds)

	items := []list.Item{}

	for _, b := range builds {
		newItem := item[apiclient.BuildDTO]{title: fmt.Sprintf("ID: %s - %s", b.Id, b.Repository.Url), desc: fmt.Sprintf("State: %s (created %s)", b.State.Name, util.FormatTimestamp(b.CreatedAt)), choiceProperty: b}
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

	title := "Select a Build To " + actionVerb
	l.Title = views.GetStyledMainTitle(title)
	l.Styles.Title = titleStyle

	m := model[apiclient.BuildDTO]{list: l}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[apiclient.BuildDTO]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}
