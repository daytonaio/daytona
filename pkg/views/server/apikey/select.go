// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikey

import (
	"errors"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

var NewApiKeyName = "+ New API Key"

func GetApiKeyFromPrompt(apiKeys []serverapiclient.ApiKey, title string, withNewApiKey bool) (*serverapiclient.ApiKey, error) {
	var items []list.Item

	for _, p := range apiKeys {
		items = append(items, item{
			apiKey: p,
		})
	}

	if withNewApiKey {
		name := NewApiKeyName
		items = append(items, item{
			apiKey: serverapiclient.ApiKey{
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

	return nil, errors.New("no API key selected")
}
