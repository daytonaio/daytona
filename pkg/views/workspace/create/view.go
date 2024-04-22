// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const maxWidth = 160

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style
}

type WorkspaceCreationPromptResponse struct {
	WorkspaceName      string
	PrimaryProject     serverapiclient.CreateWorkspaceRequestProject
	SecondaryProjects  []serverapiclient.CreateWorkspaceRequestProject
	AddingMoreProjects bool
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(0, 4, 1, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(views.Green).
		Bold(true).
		Padding(1, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(views.Green).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(views.Green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.Copy().
		Foreground(views.Green)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

type Model struct {
	lg                              *lipgloss.Renderer
	styles                          *Styles
	form                            *huh.Form
	width                           int
	workspaceCreationPromptResponse WorkspaceCreationPromptResponse
}

func RunInitialForm(primaryRepoUrl string, multiProject bool) (WorkspaceCreationPromptResponse, error) {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	primaryRepoPrompt := huh.NewInput().
		Title("Primary project repository").
		Value(&primaryRepoUrl).
		Key("primaryProjectRepo").
		Validate(func(str string) error {
			result, err := util.GetValidatedUrl(str)
			if err != nil {
				return err
			}
			primaryRepoUrl = result
			return nil
		})

	dTheme := views.GetCustomTheme()

	m.form = huh.NewForm(
		huh.NewGroup(
			primaryRepoPrompt,
		).WithHide(primaryRepoUrl != ""),
	).WithTheme(dTheme).
		WithWidth(maxWidth).
		WithShowHelp(false).
		WithShowErrors(true)

	err := m.form.Run()
	if err != nil {
		return WorkspaceCreationPromptResponse{}, err
	}

	primaryProject := serverapiclient.CreateWorkspaceRequestProject{
		Source: &serverapiclient.CreateWorkspaceRequestProjectSource{
			Repository: &serverapiclient.GitRepository{Url: &primaryRepoUrl},
		},
	}

	return WorkspaceCreationPromptResponse{
		WorkspaceName:      "",
		PrimaryProject:     primaryProject,
		SecondaryProjects:  []serverapiclient.CreateWorkspaceRequestProject{},
		AddingMoreProjects: multiProject,
	}, nil
}

func RunProjectForm(workspaceCreationPromptResponse WorkspaceCreationPromptResponse, providerRepoUrl string) (WorkspaceCreationPromptResponse, error) {
	m := Model{width: maxWidth, workspaceCreationPromptResponse: workspaceCreationPromptResponse}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	project := serverapiclient.CreateWorkspaceRequestProject{
		Source: &serverapiclient.CreateWorkspaceRequestProjectSource{
			Repository: &serverapiclient.GitRepository{Url: &providerRepoUrl},
		},
	}

	var moreCheck bool

	repositoryUrlInput :=
		huh.NewInput().
			Title(getOrderNumberString(len(workspaceCreationPromptResponse.SecondaryProjects)+1) + " secondary project repository").
			Value(project.Source.Repository.Url).
			Key(fmt.Sprintf("secondaryRepo%d", len(workspaceCreationPromptResponse.SecondaryProjects)+1)).
			Validate(func(str string) error {
				_, err := util.GetValidatedUrl(str)
				if err != nil {
					return err
				}
				return nil
			})

	confirmInput :=
		huh.NewConfirm().
			Title("Add another project?").
			Value(&moreCheck)

	var formGroup *huh.Group

	if project.Source.Repository.Url == nil || *project.Source.Repository.Url == "" {
		formGroup = huh.NewGroup(
			repositoryUrlInput,
			confirmInput,
		)
	} else {
		formGroup = huh.NewGroup(
			confirmInput,
		)
	}

	m.form = huh.NewForm(
		formGroup,
	).
		WithWidth(maxWidth).
		WithShowHelp(false).
		WithShowErrors(true).
		WithTheme(views.GetCustomTheme())

	err := m.form.Run()
	if err != nil {
		return WorkspaceCreationPromptResponse{}, err
	}

	validatedURL, err := util.GetValidatedUrl(*project.Source.Repository.Url)
	if err != nil {
		return WorkspaceCreationPromptResponse{}, err
	}

	*project.Source.Repository.Url = validatedURL
	result := workspaceCreationPromptResponse
	result.SecondaryProjects = append(result.SecondaryProjects, project)
	result.AddingMoreProjects = moreCheck

	return result, nil
}

func RunWorkspaceNameForm(workspaceCreationPromptResponse WorkspaceCreationPromptResponse, suggestedName string, workspaceNames []string) (WorkspaceCreationPromptResponse, error) {
	m := Model{width: maxWidth, workspaceCreationPromptResponse: workspaceCreationPromptResponse}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	workspaceName := suggestedName

	workspaceNamePrompt :=
		huh.NewInput().
			Title("Workspace name").
			Value(&workspaceName).
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
				workspaceName = result
				return nil
			})

	dTheme := views.GetCustomTheme()

	m.form = huh.NewForm(
		huh.NewGroup(
			workspaceNamePrompt,
		),
	).WithTheme(dTheme).
		WithWidth(maxWidth).
		WithShowHelp(false).
		WithShowErrors(true)

	err := m.form.Run()
	if err != nil {
		return WorkspaceCreationPromptResponse{}, err
	}

	result := workspaceCreationPromptResponse
	result.WorkspaceName = workspaceName

	return result, nil
}

func GetSuggestedWorkspaceName(repo string) string {
	var result strings.Builder
	input := repo
	input = strings.TrimSuffix(input, "/")

	// Find the last index of '/' in the repo string
	lastIndex := strings.LastIndex(input, "/")

	// If '/' is found, extract the content after it
	if lastIndex != -1 && lastIndex < len(repo)-1 {
		input = repo[lastIndex+1:]
	}

	input = strings.TrimSuffix(input, ".git")

	for _, char := range input {
		if unicode.IsLetter(char) || unicode.IsNumber(char) || char == '-' {
			result.WriteRune(char)
		} else if char == ' ' {
			result.WriteRune('-')
		}
	}

	return strings.ToLower(result.String())
}

func getOrderNumberString(number int) string {
	if number >= 1 && number <= 10 {
		// Handle numbers 1 to 10
		switch number {
		case 1:
			return "First"
		case 2:
			return "Second"
		case 3:
			return "Third"
		case 4:
			return "Fourth"
		case 5:
			return "Fifth"
		case 6:
			return "Sixth"
		case 7:
			return "Seventh"
		case 8:
			return "Eighth"
		case 9:
			return "Ninth"
		case 10:
			return "Tenth"
		}
	} else if number >= 11 {
		// Handle numbers 11 and beyond
		return fmt.Sprintf("%d.", number)
	}
	// Handle invalid numbers or negative numbers
	return "Invalid"
}
