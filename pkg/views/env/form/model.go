// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package form

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/views"
)

const maxWidth = 160

var userCancelled bool

type styles struct {
	HeaderText, FooterText lipgloss.Style
}

func newStyles(lg *lipgloss.Renderer) *styles {
	s := styles{}
	s.HeaderText = lg.NewStyle().
		Foreground(views.Green).
		Bold(true).
		Padding(1, 1, 0, 0)
	s.FooterText = lg.NewStyle().
		Foreground(views.Gray).
		Bold(true).
		Padding(1, 1, 0, 0)
	return &s
}

type formModel struct {
	lg       *lipgloss.Renderer
	styles   *styles
	form     *huh.Form
	width    int
	quitting bool
	tip      string
}

func NewFormModel(key, value *string) formModel {
	m := formModel{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = newStyles(m.lg)

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Key").
				Value(key).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("key cannot be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("Value").
				Value(value).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("key cannot be empty")
					}
					return nil
				}),
		),
	).WithTheme(views.GetCustomTheme()).WithHeight(12)

	m.tip = "\nTip: To set container registry credentials, add the following environment variables:\n   <*>_CONTAINER_REGISTRY_SERVER\n   <*>_CONTAINER_REGISTRY_USERNAME\n   <*>_CONTAINER_REGISTRY_PASSWORD"

	return m
}

func (m formModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			userCancelled = true
			return m, tea.Quit
		case "f10":
			m.quitting = true
			m.form.State = huh.StateCompleted
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		// Quit when the form is done.
		m.quitting = true
		cmds = append(cmds, tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func (m formModel) View() string {
	if m.quitting {
		return ""
	}

	view := m.styles.HeaderText.Render("Set server environment variable\n") + m.form.WithHeight(5).View() + m.styles.FooterText.Render(m.tip)

	return view
}
func IsUserCancelled() bool {
	return userCancelled
}
