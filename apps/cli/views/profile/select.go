// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package profile

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/views/common"
)

type editState struct {
	form   *huh.Form
	name   string
	apiUrl string
	// TODO: tokens?
}

type profileListModel struct {
	config   *config.Config
	profiles []config.Profile
	activeId string
	selected int
	quitting bool

	termWidth  int
	termHeight int

	edit *editState

	errMsg string

	// Callbacks
	onSetActive func(*config.Config, config.Profile) error
	onEdit      func(*config.Config, config.Profile) error
	onDelete    func(*config.Config, config.Profile) error
}

func (m profileListModel) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m profileListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if localMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.termWidth = localMsg.Width
		m.termHeight = localMsg.Height
	}

	if m.edit != nil {
		return m.updateEdit(msg)
	}
	return m.updateList(msg)
}

func (m profileListModel) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear error message if set
		m.errMsg = ""
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.profiles)-1 {
				m.selected++
			}
		case "enter", "e":
			return m.startEdit()
		case " ": // space
			return m.setActive()
		case "d":
			return m.delete()
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m profileListModel) startEdit() (tea.Model, tea.Cmd) {
	p := m.profiles[m.selected]

	es := &editState{
		name:   p.Name,
		apiUrl: p.Api.Url,
	}

	editForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Profile Name").
				Value(&es.name).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("name cannot be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("API URL").
				Value(&es.apiUrl).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("URL cannot be empty")
					}
					return nil
				}),
		).WithTheme(common.GetCustomTheme()),
	)

	// Prevent default behavior so the embedded form cannot quit the parent form
	editForm.SubmitCmd = nil
	editForm.CancelCmd = nil

	// Set form size
	editForm.WithWidth(m.termWidth)
	editForm.WithHeight(m.termHeight)

	es.form = editForm
	m.edit = es

	// First clear the screen and then init the edit form
	return m, tea.Batch(tea.ClearScreen, m.edit.form.Init())
}

func (m profileListModel) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	formModel, cmd := m.edit.form.Update(msg)
	m.edit.form = formModel.(*huh.Form)

	switch m.edit.form.State {
	case huh.StateCompleted:
		p := m.profiles[m.selected]
		p.Name = m.edit.name
		p.Api.Url = m.edit.apiUrl

		if m.onEdit != nil {
			err := m.onEdit(m.config, p)
			if err != nil {
				panic(err)
			}
			// TODO: Maybe handle error
		}

		m.profiles[m.selected] = p
		m.edit = nil
		return m, tea.ClearScreen

	case huh.StateAborted:
		m.edit = nil
		return m, tea.ClearScreen

	default:
		return m, cmd
	}
}

func (m profileListModel) setActive() (tea.Model, tea.Cmd) {
	p := m.profiles[m.selected]

	if m.onSetActive != nil {
		err := m.onSetActive(m.config, p)
		if err != nil {
			panic(err)
		}
	}

	m.activeId = p.Id
	return m, tea.ClearScreen
}

func (m profileListModel) delete() (tea.Model, tea.Cmd) {
	p := m.profiles[m.selected]

	if m.onDelete != nil {
		err := m.onDelete(m.config, p)
		if err != nil {
			m.errMsg = err.Error()
		}
		m.profiles = m.config.Profiles
	}

	return m, tea.ClearScreen
}

func (m profileListModel) View() string {
	if m.quitting {
		return ""
	}

	if m.edit != nil {
		return lipgloss.Place(
			m.termWidth,
			m.termHeight,
			lipgloss.Left,
			lipgloss.Top,
			m.edit.form.View(),
		)
	}

	// TODO: Since this style is being repeated from views/common/select.go
	//  Think of extracting this to the common styles.
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.Green).
		MarginLeft(2).
		MarginTop(1)

	s := titleStyle.Render("Manage Profiles") + "\n\n"

	descStyle := lipgloss.NewStyle().
		MarginLeft(6).
		Foreground(lipgloss.Color("241"))

	for i, p := range m.profiles {
		nameStyle := lipgloss.NewStyle().
			MarginLeft(4).
			Foreground(lipgloss.Color("151"))

		activeStyle := lipgloss.NewStyle().
			Foreground(common.Cyan).
			Bold(true)

		cursor := "  "

		if i == m.selected {
			cursor = "› "
			nameStyle = nameStyle.Foreground(lipgloss.Color("42")).Bold(true)
		}

		// Append [active] to the currently active profile
		profileName := p.Name
		if p.Id == m.activeId {
			profileName += activeStyle.Render("[active]")
		}

		s += nameStyle.Render(cursor+profileName) + "\n"
		s += descStyle.Render(p.Api.Url) + "\n\n"
	}

	// Footer
	helpStyle := lipgloss.NewStyle().
		Foreground(common.LightGray)
	s += helpStyle.Render("space: set active · enter/e: edit · d: delete · q: quit") + "\n\n"
	s += helpStyle.Render("to add new profiles use 'daytona login'") + "\n"

	// Error message
	if m.errMsg != "" {
		errStyle := lipgloss.NewStyle().
			Foreground(common.Red).
			MarginLeft(2)
		s += errStyle.Render(m.errMsg) + "\n"
	}

	return lipgloss.Place(
		m.termWidth,
		m.termHeight,
		lipgloss.Left,
		lipgloss.Top,
		s,
	)
}

func SelectProfile(
	cfg *config.Config,
	profiles []config.Profile,
	activeId string,
	onSetActive func(*config.Config, config.Profile) error,
	onEdit func(*config.Config, config.Profile) error,
	onDelete func(*config.Config, config.Profile) error,
) error {
	p := tea.NewProgram(
		profileListModel{
			config:      cfg,
			profiles:    profiles,
			activeId:    activeId,
			onSetActive: onSetActive,
			onEdit:      onEdit,
			onDelete:    onDelete,
		},
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	return err
}
