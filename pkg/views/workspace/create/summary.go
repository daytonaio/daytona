// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	util "github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

type ProjectDetail string

const (
	Build             ProjectDetail = "Build"
	FilePath          ProjectDetail = "File Path"
	Image             ProjectDetail = "Image"
	User              ProjectDetail = "User"
	PostStartCommands ProjectDetail = "Post Start Commands"
	EnvVars           ProjectDetail = "Env Vars"
	EMPTY_STRING                    = ""
	DEFAULT_PADDING                 = 21
)

type SummaryModel struct {
	lg              *lipgloss.Renderer
	styles          *Styles
	form            *huh.Form
	width           int
	quitting        bool
	workspaceName   string
	projectList     []apiclient.CreateWorkspaceRequestProject
	apiServerConfig *apiclient.ServerConfig
}

var configureCheck bool
var userCancelled bool
var projectsConfigurationChanged bool

func RunSubmissionForm(workspaceName *string, suggestedName string, workspaceNames []string, projectList *[]apiclient.CreateWorkspaceRequestProject, apiServerConfig *apiclient.ServerConfig) error {
	configureCheck = false

	m := NewSummaryModel(workspaceName, suggestedName, workspaceNames, *projectList, apiServerConfig)

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

	var err error
	projectsConfigurationChanged, err = ConfigureProjects(projectList, *apiServerConfig)
	if err != nil {
		return err
	}

	return RunSubmissionForm(workspaceName, suggestedName, workspaceNames, projectList, apiServerConfig)
}

func RenderSummary(workspaceName string, projectList []apiclient.CreateWorkspaceRequestProject, apiServerConfig *apiclient.ServerConfig) (string, error) {

	output := views.GetStyledMainTitle(fmt.Sprintf("SUMMARY - Workspace %s", workspaceName))

	for _, project := range projectList {
		if project.Source == nil || project.Source.Repository == nil || project.Source.Repository.Url == nil {
			return "", fmt.Errorf("repository is required")
		}
	}

	output += "\n\n"

	for i := range projectList {
		output += fmt.Sprintf("%s - %s\n", lipgloss.NewStyle().Foreground(views.Green).Render(fmt.Sprintf("%s #%d", "Project", i+1)), (*projectList[i].Source.Repository.Url))
		projectBuildChoice, choiceName := getProjectBuildChoice(projectList[i], apiServerConfig)
		output += renderProjectDetails(projectList[i], projectBuildChoice, choiceName)
		if i < len(projectList)-1 {
			output += "\n"
		}
	}

	return output, nil
}

func renderProjectDetails(project apiclient.CreateWorkspaceRequestProject, buildChoice BuildChoice, choiceName string) string {
	output := projectDetailOutput(Build, choiceName)

	if buildChoice == DEVCONTAINER {
		if project.Build != nil {
			if project.Build.Devcontainer != nil {
				if project.Build.Devcontainer.DevContainerFilePath != nil {
					output += "\n"
					output += projectDetailOutput(FilePath, *project.Build.Devcontainer.DevContainerFilePath)
				}
			}
		}
	} else {
		if project.Image != nil {
			if output != "" {
				output += "\n"
			}
			output += projectDetailOutput(Image, *project.Image)
		}

		if project.User != nil {
			if output != "" {
				output += "\n"
			}
			output += projectDetailOutput(User, *project.User)
		}

		if project.PostStartCommands != nil && len(project.PostStartCommands) > 0 {
			if output != "" {
				output += "\n"
			}

			var commands string
			for _, command := range project.PostStartCommands {
				commands += command
				commands += "; "
			}
			output += projectDetailOutput(PostStartCommands, strings.TrimSuffix(commands, "; "))
		}
	}

	if project.EnvVars != nil && len(*project.EnvVars) > 0 {
		if output != "" {
			output += "\n"
		}

		var envVars string
		for key, val := range *project.EnvVars {
			envVars += fmt.Sprintf("%s=%s; ", key, val)
		}
		output += projectDetailOutput(EnvVars, strings.TrimSuffix(envVars, "; "))
	}

	return output
}

func projectDetailOutput(projectDetailKey ProjectDetail, projectDetailValue string) string {
	return fmt.Sprintf("\t%s%-*s%s", lipgloss.NewStyle().Foreground(views.Green).Render(string(projectDetailKey)), DEFAULT_PADDING-len(string(projectDetailKey)), EMPTY_STRING, projectDetailValue)
}

func getProjectBuildChoice(project apiclient.CreateWorkspaceRequestProject, apiServerConfig *apiclient.ServerConfig) (BuildChoice, string) {
	if project.Build == nil {
		if *project.Image == *apiServerConfig.DefaultProjectImage && *project.User == *apiServerConfig.DefaultProjectUser {
			return NONE, "None"
		} else {
			return CUSTOMIMAGE, "Custom Image"
		}
	} else {
		if project.Build.Devcontainer != nil {
			return DEVCONTAINER, "Devcontainer"
		} else {
			return AUTOMATIC, "Automatic"
		}
	}
}

func NewSummaryModel(workspaceName *string, suggestedName string, workspaceNames []string, projectList []apiclient.CreateWorkspaceRequestProject, apiServerConfig *apiclient.ServerConfig) SummaryModel {
	m := SummaryModel{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	m.workspaceName = *workspaceName
	m.projectList = projectList
	m.apiServerConfig = apiServerConfig

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

	if len(m.projectList) > 1 || len(m.projectList) == 1 && projectsConfigurationChanged {
		summary, err := RenderSummary(m.workspaceName, m.projectList, m.apiServerConfig)
		if err != nil {
			log.Fatal(err)
		}
		view = views.GetBorderedMessage(summary) + "\n" + view
	}

	return view
}
