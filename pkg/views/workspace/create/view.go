// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"fmt"

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
