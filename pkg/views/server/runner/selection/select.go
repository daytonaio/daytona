// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
)

const NewTargetConfigName = "+ New Target Config"

type RunnerView struct {
	Id    string
	Alias string
}

func GetRunnerFromPrompt(runners []apiclient.RunnerDTO, activeProfileName string, actionVerb string) (*RunnerView, error) {
	items := []list.Item{}

	for _, r := range runners {
		items = append(items, item{
			runner: RunnerView{
				Id:    r.Id,
				Alias: r.Alias,
			},
		})
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = views.GetStyledMainTitle("Choose a Runner To " + actionVerb)
	m.footer = views.GetListFooter(activeProfileName, views.DefaultListFooterPadding)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}

	if m, ok := p.(model); ok && m.choice != nil {
		return m.choice, nil
	}

	return nil, common.ErrCtrlCAbort
}
