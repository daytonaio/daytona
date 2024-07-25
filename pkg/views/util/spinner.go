// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
	log "github.com/sirupsen/logrus"
)

type model struct {
	spinner  spinner.Model
	quitting bool
	message  string
}

type Msg string

func initialModel(message string) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(views.Green)
	return model{spinner: s, message: message}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case Msg:
		m.quitting = true
		return m, tea.Quit

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func WithSpinner(message string, fn func() error) error {
	p := start(message)
	defer stop(p)
	return fn()
}

func start(message string) *tea.Program {
	p := tea.NewProgram(initialModel(message), tea.WithAltScreen())
	go func() {
		if _, err := p.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()
	return p

}

func stop(p *tea.Program) {
	p.Send(Msg("quit"))
	err := p.ReleaseTerminal()
	if err != nil {
		log.Fatal(err)
	}
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	str := views.DocStyle.Render(fmt.Sprintf("\n\n   %s %s...\n\n", m.spinner.View(), m.message))

	return str
}
