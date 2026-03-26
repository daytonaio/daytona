// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cli/views/common"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
)

var isAborted bool

type model struct {
	spinner  spinner.Model
	quitting bool
	message  string
	inline   bool
}

type Msg string

func initialModel(message string, inline bool) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(common.Green)
	return model{spinner: s, message: message, inline: inline}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case Msg:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			isAborted = true
			m.quitting = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd

}

func WithSpinner(message string, fn func() error) error {
	if isTTY() {
		p := start(message, false)
		defer stop(p)
	}
	return fn()
}

func WithInlineSpinner(message string, fn func() error) error {
	if isTTY() {
		p := start(message, true)
		defer stop(p)
	}
	return fn()
}

func start(message string, inline bool) *tea.Program {
	var p *tea.Program
	if inline {
		p = tea.NewProgram(initialModel(message, true))
	} else {
		p = tea.NewProgram(initialModel(message, false), tea.WithAltScreen())
	}
	go func() {
		if _, err := p.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if isAborted {
			fmt.Println("Operation cancelled")
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

	str := ""
	if m.inline {
		str = common.GetInfoMessage(fmt.Sprintf("%s %s ...", m.spinner.View(), m.message))
	} else {
		str = common.NameStyle.Render(fmt.Sprintf("\n\n   %s %s ...\n\n", m.spinner.View(), m.message))
	}

	return str
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
