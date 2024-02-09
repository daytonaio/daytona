// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package creation_wizard

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/daytonaio/daytona/cli/cmd/views"
	"github.com/daytonaio/daytona/internal/util"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const maxWidth = 80
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
	PrimaryRepository     string
	SecondaryRepositories []string
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

func runInitialForm() (WorkspaceCreationPromptResponse, error) {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	primaryRepo := ""
	hasSecondaryProjectsCheck := false
	secondaryProjectsCountString := ""

	primaryRepoPrompt := huh.NewInput().
		Title("Primary project repository").
		Value(&primaryRepo).
		Key("primaryProjectRepo").
		Validate(func(str string) error {
			result, err := util.GetValidatedUrl(str)
			if err != nil {
				return err
			}
			primaryRepo = result
			return nil
		})

	secondaryProjectsPrompt := huh.NewConfirm().
		Key("secondaryProjectsPrompt").
		Title("Add secondary projects?").
		Value(&hasSecondaryProjectsCheck)

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
		}).Value(&secondaryProjectsCountString)

	dTheme := views.GetCustomTheme()

	m.form = huh.NewForm(
		huh.NewGroup(
			primaryRepoPrompt,
			secondaryProjectsPrompt,
		),
		huh.NewGroup(
			secondaryProjectCountPrompt,
		).WithHideFunc(func() bool {
			return !hasSecondaryProjectsCheck
		}),
	).WithTheme(dTheme).
		WithWidth(45).
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

	return WorkspaceCreationPromptResponse{
		PrimaryRepository:     primaryRepo,
		SecondaryProjectCount: secondaryProjectsCount,
	}, nil
}

func runSecondaryProjectsForm(workspaceCreationPromptResponse WorkspaceCreationPromptResponse) (WorkspaceCreationPromptResponse, error) {
	m := Model{width: maxWidth, workspaceCreationPromptResponse: workspaceCreationPromptResponse}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	var secondaryRepoList []string
	count := workspaceCreationPromptResponse.SecondaryProjectCount

	// Add empty strings to the slice
	for i := 0; i < count; i++ {
		secondaryRepoList = append(secondaryRepoList, "")
	}

	formFields := make([]huh.Field, count+1)
	for i := 0; i < count; i++ {
		formFields[i] = huh.NewInput().
			Title(getOrderNumberString(i+1) + " secondary project repository").
			Value(&secondaryRepoList[i]).
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
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(true).
		WithTheme(views.GetCustomTheme())

	err := m.form.Run()
	if err != nil {
		return WorkspaceCreationPromptResponse{}, err
	}

	for i := 0; i < count; i++ {
		secondaryRepoList[i], err = util.GetValidatedUrl(secondaryRepoList[i])
		if err != nil {
			return WorkspaceCreationPromptResponse{}, err
		}
	}

	result := workspaceCreationPromptResponse
	result.SecondaryRepositories = secondaryRepoList

	return result, nil
}

func runWorkspaceNameForm(workspaceCreationPromptResponse WorkspaceCreationPromptResponse, suggestedName string, workspaceNames []string) (WorkspaceCreationPromptResponse, error) {
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
		WithWidth(45).
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

func GetCreationDataFromPrompt(workspaceNames []string) (workspaceName string, projectRepositoryList []string, err error) {
	var projectRepoList []string

	workspaceCreationPromptResponse, err := runInitialForm()
	if err != nil {
		return "", nil, err
	}

	if workspaceCreationPromptResponse.PrimaryRepository == "" {
		return "", nil, errors.New("primary repository is required")
	}

	projectRepoList = []string{workspaceCreationPromptResponse.PrimaryRepository}

	if workspaceCreationPromptResponse.SecondaryProjectCount > 0 {

		workspaceCreationPromptResponse, err = runSecondaryProjectsForm(workspaceCreationPromptResponse)
		if err != nil {
			return "", nil, err
		}

		projectRepoList = append(projectRepoList, workspaceCreationPromptResponse.SecondaryRepositories...)
	}

	suggestedName := getSuggestedWorkspaceName(workspaceCreationPromptResponse.PrimaryRepository)

	workspaceCreationPromptResponse, err = runWorkspaceNameForm(workspaceCreationPromptResponse, suggestedName, workspaceNames)
	if err != nil {
		return "", nil, err
	}

	if workspaceCreationPromptResponse.WorkspaceName == "" {
		return "", nil, errors.New("workspace name is required")
	}

	return workspaceCreationPromptResponse.WorkspaceName, projectRepoList, nil
}

func getSuggestedWorkspaceName(repo string) string {
	var result strings.Builder
	input := repo

	// Find the last index of '/' in the repo string
	lastIndex := strings.LastIndex(repo, "/")

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
