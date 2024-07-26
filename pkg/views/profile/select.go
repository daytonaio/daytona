// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"errors"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"
)

var NewProfileId = "+ New Profile"

func GetProfileFromPrompt(profiles []config.Profile, activeProfileName string, withNewProfile bool) (*config.Profile, error) {
	var items []list.Item

	for _, p := range profiles {
		items = append(items, item{
			profile: p,
		})
	}

	if withNewProfile {
		name := NewProfileId
		items = append(items, item{
			profile: config.Profile{
				Id:   NewProfileId,
				Name: name,
			},
		})
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = views.GetStyledMainTitle("Choose a Profile")

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}

	if m, ok := p.(model); ok && m.choice != nil {
		return m.choice, nil
	}

	return nil, errors.New("no profile selected")
}
