// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

type item struct {
	targetConfig TargetConfigView
}

func (i item) Title() string { return i.targetConfig.Name }

func (i item) Description() string {
	desc := i.targetConfig.ProviderInfo.Name

	if i.targetConfig.ProviderInfo.Label != nil {
		desc = *i.targetConfig.ProviderInfo.Label
	}

	if i.targetConfig.ProviderInfo.RunnerName != "" {
		desc = fmt.Sprintf("%s (Runner %s)", desc, i.targetConfig.ProviderInfo.RunnerName)
	}

	if i.targetConfig.ProviderInfo.Installed != nil {
		if !*i.targetConfig.ProviderInfo.Installed {
			desc += " (needs installing)"
		}
	}

	return desc
}
func (i item) FilterValue() string { return i.targetConfig.Name }

type model struct {
	list   list.Model
	choice *TargetConfigView
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
				m.choice = &i.targetConfig
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

	return views.DocStyle.Width(terminalWidth - 4).Height(terminalHeight - 4).Render(m.list.View() + m.footer)
}
