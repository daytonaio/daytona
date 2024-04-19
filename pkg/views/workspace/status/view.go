// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	checkMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

type Message struct {
	Line string
}

type ClearScreenMsg struct{}

func (r Message) String() string {
	return r.Line
}

type model struct {
	spinner       spinner.Model
	messages      []Message
	quitting      bool
	width, height int
}

func NewModel() model {
	s := spinner.New()
	s.Style = spinnerStyle
	messages := make([]Message, 0)

	// messages = append(messages, Message{Line: "Workspace creation request submitted"})

	return model{
		spinner:  s,
		messages: messages,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case Message:
		if msg.Line == "END_SIGNAL" {
			m.quitting = true
			return m, tea.Quit
		}
		m.messages = append(m.messages, msg)
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

	for _, message := range m.messages {
		if message.Line != "" {
			// fmt.Printf("\r%s", msg)
			s += fmt.Sprintf("%s %s\n", checkMark, message.String())
		}
	}

	return s + m.spinner.View()
}
