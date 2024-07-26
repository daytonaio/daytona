// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"errors"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

const NewRegistryServerIdentifier = "+ New Container Registry"

func GetRegistryFromPrompt(registries []apiclient.ContainerRegistry, activeProfileName string, withNewRegistry bool) (*apiclient.ContainerRegistry, error) {
	items := util.ArrayMap(registries, func(r apiclient.ContainerRegistry) list.Item {
		return item{
			registry: r,
		}
	})

	if withNewRegistry {
		name := NewRegistryServerIdentifier
		emptyString := ""
		items = append(items, item{
			registry: apiclient.ContainerRegistry{
				Password: &emptyString,
				Username: &emptyString,
				Server:   &name,
			},
		})
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = "Choose a container registry"
	m.footer = views.GetListFooter(activeProfileName, views.DefaultListFooterPadding)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}

	if m, ok := p.(model); ok && m.choice != nil {
		return m.choice, nil
	}

	return nil, errors.New("ctrl-c exit")
}
