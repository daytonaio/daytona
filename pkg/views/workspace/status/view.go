// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	dotStyle     = helpStyle.Copy().UnsetMargins()
	appStyle     = lipgloss.NewStyle().Margin(0, 2, 1, 2).BorderStyle(lipgloss.NormalBorder()).BorderForeground(views.Green).Width(50).Padding(1, 1, 0, 2)
)

type ResultMsg struct {
	Line string
	Dots bool
}

type ClearScreenMsg struct{}

func (r ResultMsg) String() string {
	if r.Dots {
		return dotStyle.Render(strings.Repeat(".", 30))
	}
	return fmt.Sprintf(r.Line)
}

type model struct {
	spinner       spinner.Model
	results       []ResultMsg
	quitting      bool
	width, height int
}

func NewModel() model {
	const numLastResults = 6
	s := spinner.New()
	s.Style = spinnerStyle
	results := make([]ResultMsg, numLastResults)

	for i := range results {
		results[i] = ResultMsg{Dots: true}
	}

	results[len(results)-1] = ResultMsg{Line: "Workspace creation request submitted", Dots: false}

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
	case ClearScreenMsg:
		m.quitting = true
		return m, nil
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m model) View() string {
	var s string

	for _, res := range m.results {
		s += res.String() + "\n"
	}

	return appStyle.Width(m.width - 20).Render(s)
}
