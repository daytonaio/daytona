// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cli/internal"
	"github.com/daytonaio/daytona/cli/internal/clierr"
)

// SelectItem represents an item in the selection list
type SelectItem struct {
	Title string
	Desc  string
}

// SelectModel represents the selection UI model
type SelectModel struct {
	Title    string
	Items    []SelectItem
	Selected int
	Choice   string
	Quitting bool
}

// NewSelectModel creates a new select model with the given title and items
func NewSelectModel(title string, items []SelectItem) SelectModel {
	return SelectModel{
		Title:    title,
		Items:    items,
		Selected: 0,
	}
}

func (m SelectModel) Init() tea.Cmd {
	return nil
}

func (m SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.Selected > 0 {
				m.Selected--
			}
		case "down", "j":
			if m.Selected < len(m.Items)-1 {
				m.Selected++
			}
		case "enter":
			m.Choice = m.Items[m.Selected].Title
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SelectModel) View() string {
	if m.Quitting {
		return ""
	}

	s := lipgloss.NewStyle().
		Bold(true).
		MarginLeft(2).
		MarginTop(1).
		Render(m.Title) + "\n\n"

	for i, item := range m.Items {
		cursor := "  "
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("151")).
			PaddingLeft(2)

		if i == m.Selected {
			cursor = "› "
			style = style.Foreground(lipgloss.Color("42")).Bold(true)
		}

		s += style.Render(cursor+item.Title) + "\n"
		s += lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(lipgloss.Color("241")).
			Render(item.Desc) + "\n\n"
	}

	return s
}

// Select displays a selection prompt with the given title and items
// Returns the selected item's title and any error that occurred
func Select(title string, items []SelectItem) (string, error) {
	if !internal.Interactive() {
		return "", clierr.New(clierr.CategoryUsage, "cannot prompt for input in non-interactive mode").WithHint("re-run interactively or provide the value via flags")
	}

	p := tea.NewProgram(NewSelectModel(title, items), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	finalModel, ok := m.(SelectModel)
	if !ok {
		return "", nil
	}

	if finalModel.Quitting {
		return "", nil
	}

	return finalModel.Choice, nil
}
