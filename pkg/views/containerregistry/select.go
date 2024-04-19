// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

const NewRegistryServerIdentifier = "+ New Container Registry"

func GetRegistryFromPrompt(registries []serverapiclient.ContainerRegistry, activeProfileName string, withNewRegistry bool) (*serverapiclient.ContainerRegistry, error) {
	items := util.ArrayMap(registries, func(r serverapiclient.ContainerRegistry) list.Item {
		return item{
			registry: r,
		}
	})

	if withNewRegistry {
		name := NewRegistryServerIdentifier
		emptyString := ""
		items = append(items, item{
			registry: serverapiclient.ContainerRegistry{
				Password: &emptyString,
				Username: &emptyString,
				Server:   &name,
			},
		})
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = "Choose a container registry"
	m.footer = views.GetListFooter(activeProfileName)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model); ok && m.choice != nil {
		return m.choice, nil
	}

	return nil, errors.New("no container registry selected")
}
