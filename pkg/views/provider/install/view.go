// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider_install

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

type item struct {
	provider ProviderInstallView
}

func (i item) Title() string {
	title := i.provider.Name
	if i.provider.Label != nil {
		title = *i.provider.Label
	}

	return title
}

func (i item) Description() string {
	return i.provider.Version
}
func (i item) FilterValue() string {
	if i.provider.Label != nil {
		return *i.provider.Label
	} else {
		return i.provider.Name
	}
}

type model struct {
	list   list.Model
	choice *ProviderInstallView
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
				m.choice = &i.provider
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
	terminalWidth, terminalHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return ""
	}

	return views.DocStyle.Width(terminalWidth - 4).Height(terminalHeight - 4).Render(m.list.View())
}