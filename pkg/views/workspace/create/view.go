// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const maxWidth = 160
const maximumSecondaryProjects = 8

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
	WorkspaceName         string
	PrimaryRepository     serverapiclient.GitRepository
	SecondaryRepositories []serverapiclient.GitRepository
	SecondaryProjectCount int
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

func RunInitialForm(providerRepo serverapiclient.GitRepository, multiProject bool) (WorkspaceCreationPromptResponse, error) {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	var primaryRepoUrl string

	if providerRepo.Url != nil {
		primaryRepoUrl = *providerRepo.Url
	}

	secondaryProjectsCountString := ""

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

	secondaryProjectCountPrompt := huh.NewInput().
		Title("How many secondary projects?").
		Value(&secondaryProjectsCountString).
		Validate(func(str string) error {
			count, err := strconv.Atoi(str) // Try to convert the input string to an integer
			if err != nil {
				return errors.New("enter a valid number")
			}
			if count > maximumSecondaryProjects {
				return errors.New("maximum 8 secondary projects allowed")
			}
			return nil
		})

	dTheme := views.GetCustomTheme()

	m.form = huh.NewForm(
		huh.NewGroup(
			primaryRepoPrompt,
		).WithHide(primaryRepoUrl != ""),
		huh.NewGroup(
			secondaryProjectCountPrompt,
		).WithHide(!multiProject),
	).WithTheme(dTheme).
		WithWidth(maxWidth).
		WithShowHelp(false).
		WithShowErrors(true)

	err := m.form.Run()
	if err != nil {
		return WorkspaceCreationPromptResponse{}, err
	}

	secondaryProjectsCount, err := strconv.Atoi(secondaryProjectsCountString)
	if err != nil {
		secondaryProjectsCount = 0
	}

	providerRepo.Url = &primaryRepoUrl

	return WorkspaceCreationPromptResponse{
		PrimaryRepository:     providerRepo,
		SecondaryProjectCount: secondaryProjectsCount,
	}, nil
}

func RunSecondaryProjectsForm(workspaceCreationPromptResponse WorkspaceCreationPromptResponse) (WorkspaceCreationPromptResponse, error) {
	m := Model{width: maxWidth, workspaceCreationPromptResponse: workspaceCreationPromptResponse}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	var secondaryRepoList []serverapiclient.GitRepository
	count := workspaceCreationPromptResponse.SecondaryProjectCount

	secondaryRepoList = workspaceCreationPromptResponse.SecondaryRepositories

	// Add empty strings to the slice
	for i := 0; i < (count - len(workspaceCreationPromptResponse.SecondaryRepositories)); i++ {
		emptyString := ""
		secondaryRepoList = append(secondaryRepoList, serverapiclient.GitRepository{
			Url: &emptyString,
		})
	}

	formFields := make([]huh.Field, count+1)
	for i := 0; i < count; i++ {
		formFields[i] = huh.NewInput().
			Title(getOrderNumberString(i+1) + " secondary project repository").
			Value(secondaryRepoList[i].Url).
			Key(fmt.Sprintf("secondaryRepo%d", i)).
			Validate(func(str string) error {
				_, err := util.GetValidatedUrl(str)
				if err != nil {
					return err
				}
				return nil
			})
	}

	formFields[count] = huh.NewConfirm().
		Title("Good to go?").
		Validate(func(v bool) error {
			if !v {
				return fmt.Errorf("double-check and hit 'Yes'")
			}
			return nil
		})

	secondaryRepoGroup := huh.NewGroup(
		formFields...,
	)

	m.form = huh.NewForm(
		secondaryRepoGroup,
	).
		WithWidth(maxWidth).
		WithShowHelp(false).
		WithShowErrors(true).
		WithTheme(views.GetCustomTheme())

	err := m.form.Run()
	if err != nil {
		return WorkspaceCreationPromptResponse{}, err
	}

	for i := 0; i < count; i++ {
		validatedURL, err := util.GetValidatedUrl(*secondaryRepoList[i].Url)
		if err != nil {
			return WorkspaceCreationPromptResponse{}, err
		}
		*secondaryRepoList[i].Url = validatedURL
	}

	result := workspaceCreationPromptResponse
	result.SecondaryRepositories = secondaryRepoList

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
