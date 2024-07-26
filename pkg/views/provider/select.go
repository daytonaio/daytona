// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"errors"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

var NewProviderId = "+ New Provider"

func GetProviderFromPrompt(providers []apiclient.Provider, title string, withNewProvider bool) (*apiclient.Provider, error) {
	var items []list.Item

	for _, p := range providers {
		items = append(items, item{
			provider: p,
		})
	}

	if withNewProvider {
		name := NewProviderId
		items = append(items, item{
			provider: apiclient.Provider{
				Name: &name,
			},
		})
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = views.GetStyledMainTitle(title)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}

	if m, ok := p.(model); ok && m.choice != nil {
		return m.choice, nil
	}

	return nil, errors.New("ctrl-c exit")
}
