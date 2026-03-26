// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type promptModel struct {
	textInput textinput.Model
	err       error
	done      bool
	title     string
	desc      string
}

func (m promptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC:
			m.done = true
			m.err = fmt.Errorf("user cancelled")
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m promptModel) View() string {
	if m.done {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		MarginLeft(2).
		MarginTop(1)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginLeft(2)

	return fmt.Sprintf("\n%s\n%s\n\n %s\n\n",
		titleStyle.Render(m.title),
		descStyle.Render(m.desc),
		m.textInput.View())
}

func PromptForInput(prompt, title, desc string) (string, error) {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 80
	ti.Prompt = "› "

	m := promptModel{
		textInput: ti,
		title:     title,
		desc:      desc,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	finalModel, ok := model.(promptModel)
	if !ok {
		return "", fmt.Errorf("could not read model state")
	}

	if finalModel.err != nil {
		return "", finalModel.err
	}

	return strings.TrimSpace(finalModel.textInput.Value()), nil
}
