// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"fmt"
	"log"
	"math"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	util "github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
)

type ProjectDetail string

const (
	Build              ProjectDetail = "Build"
	DevcontainerConfig ProjectDetail = "Devcontainer Config"
	Image              ProjectDetail = "Image"
	User               ProjectDetail = "User"
	EnvVars            ProjectDetail = "Env Vars"
	EMPTY_STRING                     = ""
	DEFAULT_PADDING                  = 21
)

type SummaryModel struct {
	lg          *lipgloss.Renderer
	styles      *Styles
	form        *huh.Form
	width       int
	height      int
	quitting    bool
	name        string
	projectList []apiclient.CreateProjectDTO
	defaults    *views_util.ProjectConfigDefaults
	nameLabel   string
	viewport    viewport.Model
}

type SubmissionFormConfig struct {
	ChosenName    *string
	SuggestedName string
	ExistingNames []string
	ProjectList   *[]apiclient.CreateProjectDTO
	NameLabel     string
	Defaults      *views_util.ProjectConfigDefaults
}

var configureCheck bool
var userCancelled bool
var ProjectsConfigurationChanged bool

func RunSubmissionForm(config SubmissionFormConfig) error {
	configureCheck = false

	m := NewSummaryModel(config)

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
		return errors.New("default project entries are not set")
	}

	var err error
	ProjectsConfigurationChanged, err = RunProjectConfiguration(config.ProjectList, *config.Defaults)
	if err != nil {
		return err
	}

	return RunSubmissionForm(config)
}

func RenderSummary(name string, projectList []apiclient.CreateProjectDTO, defaults *views_util.ProjectConfigDefaults, nameLabel string) (string, error) {
	var output string
	if name == "" {
		output = views.GetStyledMainTitle("SUMMARY")
	} else {
		output = views.GetStyledMainTitle(fmt.Sprintf("SUMMARY - %s %s", nameLabel, name))
	}

	output += "\n\n"

	for i := range projectList {
		if len(projectList) == 1 {
			output += fmt.Sprintf("%s - %s\n", lipgloss.NewStyle().Foreground(views.Green).Render("Project"), (projectList[i].Source.Repository.Url))
		} else {
			output += fmt.Sprintf("%s - %s\n", lipgloss.NewStyle().Foreground(views.Green).Render(fmt.Sprintf("%s #%d", "Project", i+1)), (projectList[i].Source.Repository.Url))
		}

		projectBuildChoice, choiceName := views_util.GetProjectBuildChoice(projectList[i], defaults)
		output += renderProjectDetails(projectList[i], projectBuildChoice, choiceName)
		if i < len(projectList)-1 {
			output += "\n\n"
		}
	}

	return output, nil
}

func renderProjectDetails(project apiclient.CreateProjectDTO, buildChoice views_util.BuildChoice, choiceName string) string {
	output := projectDetailOutput(Build, choiceName)

	if buildChoice == views_util.DEVCONTAINER {
		if project.BuildConfig != nil {
			if project.BuildConfig.Devcontainer != nil {
				output += "\n"
				output += projectDetailOutput(DevcontainerConfig, project.BuildConfig.Devcontainer.FilePath)
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
	}

	if len(project.EnvVars) > 0 {
		if output != "" {
			output += "\n"
		}

		var envVars string
		for key, val := range project.EnvVars {
			envVars += fmt.Sprintf("%s=%s; ", key, val)
		}
		output += projectDetailOutput(EnvVars, strings.TrimSuffix(envVars, "; "))
	}

	return output
}

func projectDetailOutput(projectDetailKey ProjectDetail, projectDetailValue string) string {
	return fmt.Sprintf("\t%s%-*s%s", lipgloss.NewStyle().Foreground(views.Green).Render(string(projectDetailKey)), DEFAULT_PADDING-len(string(projectDetailKey)), EMPTY_STRING, projectDetailValue)
}

func calculateViewportSize(content string, terminalHeight int) (width, height int) {
	lines := strings.Split(content, "\n")
	longestLine := slices.MaxFunc(lines, func(a, b string) int {
		return len(a) - len(b)
	})
	width = len(longestLine)

	maxHeight := terminalHeight

	height = int(math.Min(float64(len(lines)), float64(maxHeight)))

	return width, height
}

func NewSummaryModel(config SubmissionFormConfig) SummaryModel {
	m := SummaryModel{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	m.name = *config.ChosenName
	m.projectList = *config.ProjectList
	m.defaults = config.Defaults
	m.nameLabel = config.NameLabel

	if *config.ChosenName == "" {
		*config.ChosenName = config.SuggestedName
	}

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

	content, _ := RenderSummary(m.name, m.projectList, m.defaults, m.nameLabel)

	// Dynamically calculate viewport size
	m.width, m.height = calculateViewportSize(content, views_util.GetTerminalHeight())
	m.viewport = viewport.New(m.width, m.height)
	m.viewport.SetContent(content)

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
		case "up":
			m.viewport.LineUp(1) // Scroll up
		case "down":
			m.viewport.LineDown(1) // Scroll down
		}
	case tea.WindowSizeMsg:
		m.viewport.Height = max(1, min(m.height, msg.Height-15))
		m.viewport.Width = max(20, min(maxWidth, min(m.width, msg.Width-15)))

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

	helpLine := helpStyle.Render("enter: next • f10: advanced configuration")
	var content string

	if len(m.projectList) > 1 || ProjectsConfigurationChanged {
		content = renderSummaryView(m)
	} else {
		content = m.form.WithHeight(5).View()
	}

	return content + "\n" + helpLine
}

func renderSummaryView(m SummaryModel) string {
	summary, err := RenderSummary(m.name, m.projectList, m.defaults, m.nameLabel)
	if err != nil {
		log.Fatal(err)
	}
	m.viewport.SetContent(summary)

	return lipgloss.JoinVertical(lipgloss.Top, renderBody(m), renderFooter(m)) + m.form.WithHeight(5).View()
}

func renderBody(m SummaryModel) string {
	return m.viewport.Style.
		Margin(1, 0, 0).
		Padding(1, 2).
		BorderForeground(views.LightGray).
		Border(lipgloss.RoundedBorder()).
		Render(m.viewport.View())
}

func renderFooter(m SummaryModel) string {
	return helpStyle.Align(lipgloss.Right).Width(m.viewport.Width + 4).Render("↑ up • ↓ down")
}
