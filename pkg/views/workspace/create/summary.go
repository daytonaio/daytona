// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/util"
)

type SummaryModel struct {
	lg            *lipgloss.Renderer
	styles        *Styles
	form          *huh.Form
	width         int
	quitting      bool
	workspaceName string
	projectList   []serverapiclient.CreateWorkspaceRequestProject
}

var configureCheck bool
var confirmationResult bool

func DisplayMultiSubmitForm(workspaceName string, projectList *[]serverapiclient.CreateWorkspaceRequestProject, apiServerConfig *serverapiclient.ServerConfig, doneCheck *bool) error {
	configureCheck = false

	m := NewSummaryModel(workspaceName, *projectList)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return err
	}

	*doneCheck = confirmationResult

	if !configureCheck {
		return nil
	}

	if apiServerConfig.DefaultProjectImage == nil || apiServerConfig.DefaultProjectUser == nil || apiServerConfig.DefaultProjectPostStartCommands == nil {
		log.Fatal("Default project entries are not set")
	}

	for i := range *projectList {
		(*projectList)[i].Image = apiServerConfig.DefaultProjectImage
		(*projectList)[i].User = apiServerConfig.DefaultProjectUser
		(*projectList)[i].PostStartCommands = apiServerConfig.DefaultProjectPostStartCommands
	}

	defaultPostStartCommandString := util.GetJoinedCommands(apiServerConfig.DefaultProjectPostStartCommands)

	configuredProjects, err := ConfigureProjects(*projectList, *apiServerConfig.DefaultProjectImage, *apiServerConfig.DefaultProjectUser, defaultPostStartCommandString)
	if err != nil {
		log.Fatal(err)
	}

	*projectList = configuredProjects

	return DisplayMultiSubmitForm(workspaceName, projectList, apiServerConfig, doneCheck)
}

func RenderSummary(workspaceName string, projectList []serverapiclient.CreateWorkspaceRequestProject) (string, error) {

	output := views.GetStyledMainTitle(fmt.Sprintf("SUMMARY - Workspace %s", workspaceName))

	for _, project := range projectList {
		if project.Source == nil || project.Source.Repository == nil || project.Source.Repository.Url == nil {
			return "", fmt.Errorf("Repository is required")
		}
	}

	output += fmt.Sprintf("\n\n%s - %s\n", lipgloss.NewStyle().Foreground(views.Green).Render("Primary Project"), *projectList[0].Source.Repository.Url)

	// Remove the primary project from the list
	projectList = projectList[1:]

	if len(projectList) > 1 {
		output += "\n"
	}

	for i := range projectList {
		output += fmt.Sprintf("%s - %s", lipgloss.NewStyle().Foreground(views.Green).Render(fmt.Sprintf("#%d %s", i+1, "Secondary Project")), (*projectList[i].Source.Repository.Url))
		if i < len(projectList)-1 {
			output += "\n"
		}
	}

	return output, nil
}

func NewSummaryModel(workspaceName string, projectList []serverapiclient.CreateWorkspaceRequestProject) SummaryModel {
	m := SummaryModel{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	m.workspaceName = workspaceName
	m.projectList = projectList

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Good to go?").
				Negative("Abort").
				Value(&confirmationResult),
		),
	).WithShowHelp(false).WithTheme(views.GetCustomTheme())

	return m
}

func (m SummaryModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m SummaryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "f10":
			m.quitting = true
			m.form.State = huh.StateCompleted
			configureCheck = true
			confirmationResult = true
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

func (m SummaryModel) View() string {
	if m.quitting {
		return ""
	}

	summary, err := RenderSummary(m.workspaceName, m.projectList)
	if err != nil {
		log.Fatal(err)
	}

	return views.GetBorderedMessage(summary) + "\n" + m.form.View() + configurationHelpLine
}
