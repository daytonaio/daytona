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
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type WorkspaceDetail string

const (
	Build              WorkspaceDetail = "Build"
	DevcontainerConfig WorkspaceDetail = "Devcontainer Config"
	Image              WorkspaceDetail = "Image"
	User               WorkspaceDetail = "User"
	EnvVars            WorkspaceDetail = "Env Vars"
	EMPTY_STRING                       = ""
	DEFAULT_PADDING                    = 21
)

type SummaryModel struct {
	lg            *lipgloss.Renderer
	styles        *Styles
	form          *huh.Form
	width         int
	quitting      bool
	name          string
	workspaceList []apiclient.CreateWorkspaceDTO
	defaults      *views_util.WorkspaceConfigDefaults
	nameLabel     string
}

type SubmissionFormConfig struct {
	ChosenName    *string
	SuggestedName string
	ExistingNames []string
	WorkspaceList *[]apiclient.CreateWorkspaceDTO
	NameLabel     string
	Defaults      *views_util.WorkspaceConfigDefaults
}

var configureCheck bool
var userCancelled bool
var WorkspacesConfigurationChanged bool

func RunSubmissionForm(config SubmissionFormConfig, wtImport *bool) error {
	configureCheck = false

	m := NewSummaryModel(config, wtImport)

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		return err
	}

	if userCancelled {
		return errors.New("user cancelled")
	}

	if !configureCheck {
		return nil
	}

	if config.Defaults.Image == nil || config.Defaults.ImageUser == nil {
		return errors.New("default workspace entries are not set")
	}

	var err error
	WorkspacesConfigurationChanged, err = RunWorkspaceConfiguration(config.WorkspaceList, *config.Defaults, *wtImport)
	if err != nil {
		return err
	}

	return RunSubmissionForm(config, wtImport)
}

func RenderSummary(name string, workspaceList []apiclient.CreateWorkspaceDTO, defaults *views_util.WorkspaceConfigDefaults, nameLabel string) (string, error) {
	var output string
	if name == "" {
		output = views.GetStyledMainTitle("SUMMARY")
	} else {
		output = views.GetStyledMainTitle(fmt.Sprintf("SUMMARY - %s %s", nameLabel, name))
	}

	output += "\n\n"

	for i := range workspaceList {
		if len(workspaceList) == 1 {
			output += fmt.Sprintf("%s - %s\n", lipgloss.NewStyle().Foreground(views.Green).Render("Workspace"), (workspaceList[i].Source.Repository.Url))
		} else {
			output += fmt.Sprintf("%s - %s\n", lipgloss.NewStyle().Foreground(views.Green).Render(fmt.Sprintf("%s #%d", "Workspace", i+1)), (workspaceList[i].Source.Repository.Url))
		}

		workspaceBuildChoice, choiceName := views_util.GetWorkspaceBuildChoice(workspaceList[i], defaults)
		output += renderWorkspaceDetails(workspaceList[i], workspaceBuildChoice, choiceName)
		if i < len(workspaceList)-1 {
			output += "\n\n"
		}
	}

	return output, nil
}

func renderWorkspaceDetails(workspace apiclient.CreateWorkspaceDTO, buildChoice views_util.BuildChoice, choiceName string) string {
	output := workspaceDetailOutput(Build, choiceName)

	if buildChoice == views_util.DEVCONTAINER {
		if workspace.BuildConfig != nil {
			if workspace.BuildConfig.Devcontainer != nil {
				output += "\n"
				output += workspaceDetailOutput(DevcontainerConfig, workspace.BuildConfig.Devcontainer.FilePath)
			}
		}
	} else {
		if workspace.Image != nil {
			if output != "" {
				output += "\n"
			}
			output += workspaceDetailOutput(Image, *workspace.Image)
		}

		if workspace.User != nil {
			if output != "" {
				output += "\n"
			}
			output += workspaceDetailOutput(User, *workspace.User)
		}
	}

	if len(workspace.EnvVars) > 0 {
		if output != "" {
			output += "\n"
		}

		var envVars string
		for key, val := range workspace.EnvVars {
			envVars += fmt.Sprintf("%s=%s; ", key, val)
		}
		output += workspaceDetailOutput(EnvVars, strings.TrimSuffix(envVars, "; "))
	}

	return output
}

func workspaceDetailOutput(workspaceDetailKey WorkspaceDetail, workspaceDetailValue string) string {
	return fmt.Sprintf("\t%s%-*s%s", lipgloss.NewStyle().Foreground(views.Green).Render(string(workspaceDetailKey)), DEFAULT_PADDING-len(string(workspaceDetailKey)), EMPTY_STRING, workspaceDetailValue)
}

func NewSummaryModel(config SubmissionFormConfig, wtImport *bool) SummaryModel {
	m := SummaryModel{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	m.name = *config.ChosenName
	m.workspaceList = *config.WorkspaceList
	m.defaults = config.Defaults
	m.nameLabel = config.NameLabel

	if *config.ChosenName == "" {
		*config.ChosenName = config.SuggestedName
	}

	if !*wtImport {
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("%s name", config.NameLabel)).
					Value(config.ChosenName).
					Key("name").
					Validate(func(str string) error {
						result, err := util.GetValidatedName(str)
						if err != nil {
							return err
						}
						for _, name := range config.ExistingNames {
							if name == result {
								return errors.New("name already exists")
							}
						}
						*config.ChosenName = result
						return nil
					}),
			),
		).WithShowHelp(false).WithTheme(views.GetCustomTheme())
	} else {
		m.form = huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Is the above information correct?").
					Value(wtImport),
			),
		).WithShowHelp(false).WithTheme(views.GetCustomTheme())
	}

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

	view := m.form.WithHeight(5).View() + "\n" + configurationHelpLine

	if len(m.workspaceList) > 1 || len(m.workspaceList) == 1 && WorkspacesConfigurationChanged {
		summary, err := RenderSummary(m.name, m.workspaceList, m.defaults, m.nameLabel)
		if err != nil {
			log.Fatal(err)
		}
		view = views.GetBorderedMessage(summary) + "\n" + view
	}

	return view
}
