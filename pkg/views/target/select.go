// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"errors"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

const NewTargetName = "+ New Target"

func GetTargetFromPrompt(targets []serverapiclient.ProviderTarget, activeProfileName string, withNewTarget bool) (*serverapiclient.ProviderTarget, error) {
	items := util.ArrayMap(targets, func(t serverapiclient.ProviderTarget) list.Item {
		return item{
			target: t,
		}
	})

	if withNewTarget {
		name := NewTargetName
		options := "{}"
		items = append(items, item{
			target: serverapiclient.ProviderTarget{
				Name:    &name,
				Options: &options,
			},
		})
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = views.GetStyledMainTitle("Choose a Target")
	m.footer = views.GetListFooter(activeProfileName)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}

	if m, ok := p.(model); ok && m.choice != nil {
		return m.choice, nil
	}

	return nil, errors.New("no target selected")
}
