// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	dotStyle     = helpStyle.Copy().UnsetMargins()
	appStyle     = lipgloss.NewStyle().Margin(1, 2, 1, 2).BorderStyle(lipgloss.NormalBorder()).BorderForeground(views.Green).Width(50).Padding(1, 1, 0, 2)
)

type ResultMsg struct {
	Duration time.Duration
	Line     string
}

func (r ResultMsg) String() string {
	if r.Duration == 0 {
		return dotStyle.Render(strings.Repeat(".", 36))
	}
	return fmt.Sprintf(r.Line)
}

type model struct {
	spinner  spinner.Model
	results  []ResultMsg
	quitting bool
}

func NewModel() model {
	const numLastResults = 4
	s := spinner.New()
	s.Style = spinnerStyle
	results := make([]ResultMsg, numLastResults)
	results[0] = ResultMsg{Line: "Workspace creation is pending..."}
	return model{
		spinner: s,
		results: results,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ResultMsg:
		if msg.Line == "END_SIGNAL" {
			m.quitting = true
			return m, tea.Quit
		}
		m.results = append(m.results[1:], msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View() string {
	var s string

	for _, res := range m.results {
		s += res.String() + "\n"
	}

	return appStyle.Render(s)
}
