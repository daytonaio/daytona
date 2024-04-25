// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	configure "github.com/daytonaio/daytona/pkg/views/server"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
)

var doneCheck bool

type WorkspaceDataModel struct {
	lg                      *lipgloss.Renderer
	styles                  *Styles
	form                    *huh.Form
	basicViewActive         bool
	width                   int
	quitting                bool
	showConfigurationOption bool
}

func GetWorkspaceDataFromPrompt(apiServerConfig *serverapiclient.ServerConfig, suggestedName string, workspaceNames []string, showConfigurationOption bool) (string, string, string, []string, error) {
	if apiServerConfig.DefaultProjectImage == nil || apiServerConfig.DefaultProjectUser == nil || apiServerConfig.DefaultProjectPostStartCommands == nil {
		return "", "", "", nil, errors.New("default project entries are not set")
	}

	var postStartCommands []string

	postStartCommandString := view_util.GetJoinedCommands(apiServerConfig.DefaultProjectPostStartCommands)

	workspaceName, containerImage, containerUser := suggestedName, *apiServerConfig.DefaultProjectImage, *apiServerConfig.DefaultProjectUser

	m := NewWorkspaceDataModel(workspaceNames, &workspaceName, &containerImage, &containerUser, &postStartCommandString, showConfigurationOption)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if !doneCheck {
		return "", "", "", nil, errors.New("workspace creation cancelled")
	}

	if postStartCommandString != "" {
		postStartCommands = view_util.GetSplitCommands(postStartCommandString)
	}

	return workspaceName, containerImage, containerUser, postStartCommands, nil
}

func NewWorkspaceDataModel(workspaceNames []string, workspaceName *string, containerImage *string, containerUser *string, postStartCommands *string, showConfigurationOption bool) WorkspaceDataModel {
	m := WorkspaceDataModel{width: maxWidth, basicViewActive: true, showConfigurationOption: showConfigurationOption}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	workspaceNamePrompt :=
		huh.NewInput().
			Title("Workspace name").
			Value(workspaceName).
			Key("workspaceName").
			Validate(func(str string) error {
				result, err := util.GetValidatedWorkspaceName(str)
				if err != nil {
					return err
				}
				for _, name := range workspaceNames {
					if name == result {
						return errors.New("workspace name already exists")
					}
				}
				*workspaceName = result
				return nil
			})

	dTheme := views.GetCustomTheme()

	m.form = huh.NewForm(
		huh.NewGroup(
			workspaceNamePrompt,
		),
		GetProjectConfigurationGroup(containerImage, containerUser, postStartCommands),
	).WithTheme(dTheme).
		WithWidth(maxWidth).
		WithShowErrors(true).WithShowHelp(false)

	return m
}

func GetProjectConfigurationGroup(image *string, user *string, postStartCommands *string) *huh.Group {
	group := huh.NewGroup(
		huh.NewInput().
			Title("Custom container image").
			Value(image),
		huh.NewInput().
			Title("Container user").
			Value(user),
		huh.NewInput().
			Title("Post start commands").
			Description(configure.CommandsInputHelp).
			Value(postStartCommands),
	)

	return group
}

func (m WorkspaceDataModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m WorkspaceDataModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.basicViewActive {
				m.form.State = huh.StateCompleted
			}
			doneCheck = true
		case "f10":
			if !m.showConfigurationOption {
				return m, nil
			}
			if m.basicViewActive {
				m.form.NextGroup()
			} else {
				m.form.PrevGroup()
			}
			m.basicViewActive = !m.basicViewActive
		}
	}

	var cmds []tea.Cmd

	// Process the form
	activeForm, cmd := m.form.Update(msg)
	if f, ok := activeForm.(*huh.Form); ok {
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

func (m WorkspaceDataModel) View() string {
	if m.quitting {
		return ""
	}

	view := m.form.View()

	if m.showConfigurationOption {
		view += configurationHelpLine
	}

	return view
}
