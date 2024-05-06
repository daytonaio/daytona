// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

type item struct {
	registry serverapiclient.ContainerRegistry
}

func (i item) Title() string { return *i.registry.Server }
func (i item) Description() string {
	if *i.registry.Server == NewRegistryServerIdentifier {
		return "Add a new container registry"
	}
	return *i.registry.Username
}
func (i item) FilterValue() string { return *i.registry.Server }

type model struct {
	list   list.Model
	choice *serverapiclient.ContainerRegistry
	footer string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = &i.registry
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := views.DocStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return views.DocStyle.Render(m.list.View() + m.footer)
}
