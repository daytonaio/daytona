// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	util "github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
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
var userCancelled bool

func RunSubmissionForm(workspaceName *string, suggestedName string, workspaceNames []string, projectList *[]serverapiclient.CreateWorkspaceRequestProject, apiServerConfig *serverapiclient.ServerConfig) error {
	configureCheck = false

	m := NewSummaryModel(workspaceName, suggestedName, workspaceNames, *projectList)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return err
	}

	if userCancelled {
		return errors.New("user cancelled")
	}

	if !configureCheck {
		return nil
	}

	if apiServerConfig.DefaultProjectImage == nil || apiServerConfig.DefaultProjectUser == nil || apiServerConfig.DefaultProjectPostStartCommands == nil {
		return fmt.Errorf("default project entries are not set")
	}

	for i := range *projectList {
		if (*projectList)[i].Image == nil {
			(*projectList)[i].Image = apiServerConfig.DefaultProjectImage
		}
		if (*projectList)[i].User == nil {
			(*projectList)[i].User = apiServerConfig.DefaultProjectUser
		}
		if (*projectList)[i].PostStartCommands == nil {
			(*projectList)[i].PostStartCommands = apiServerConfig.DefaultProjectPostStartCommands
		}
		if (*projectList)[i].EnvVars == nil {
			(*projectList)[i].EnvVars = &map[string]string{}
		}
		if (*projectList)[i].Build == nil {
			(*projectList)[i].Build = &serverapiclient.ProjectBuild{}
		}
	}

	configuredProjects, err := ConfigureProjects(*projectList, *apiServerConfig)
	if err != nil {
		return err
	}

	*projectList = configuredProjects

	return RunSubmissionForm(workspaceName, suggestedName, workspaceNames, projectList, apiServerConfig)
}

func RenderSummary(workspaceName string, projectList []serverapiclient.CreateWorkspaceRequestProject) (string, error) {

	output := views.GetStyledMainTitle(fmt.Sprintf("SUMMARY - Workspace %s", workspaceName))

	for _, project := range projectList {
		if project.Source == nil || project.Source.Repository == nil || project.Source.Repository.Url == nil {
			return "", fmt.Errorf("repository is required")
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

func NewSummaryModel(workspaceName *string, suggestedName string, workspaceNames []string, projectList []serverapiclient.CreateWorkspaceRequestProject) SummaryModel {
	m := SummaryModel{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	m.workspaceName = *workspaceName
	m.projectList = projectList

	if *workspaceName == "" {
		*workspaceName = suggestedName
	}

	m.form = huh.NewForm(
		huh.NewGroup(
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
				}),
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
			userCancelled = true
			m.quitting = true
			return m, tea.Quit
		case "f10":
			m.quitting = true
			m.form.State = huh.StateCompleted
			configureCheck = true
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

	view := m.form.View() + configurationHelpLine

	if len(m.projectList) > 1 {
		summary, err := RenderSummary(m.workspaceName, m.projectList)
		if err != nil {
			log.Fatal(err)
		}
		view = views.GetBorderedMessage(summary) + "\n" + view
	}

	return view
}
