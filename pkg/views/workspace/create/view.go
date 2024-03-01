// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

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

func runInitialForm(providerRepoUrl string, multiProject bool) (WorkspaceCreationPromptResponse, error) {
	m := Model{width: maxWidth}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	primaryRepo := providerRepoUrl
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
		),
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

	secondaryRepoList = workspaceCreationPromptResponse.SecondaryRepositories

	// Add empty strings to the slice
	for i := 0; i < (count - len(workspaceCreationPromptResponse.SecondaryRepositories)); i++ {
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
		WithWidth(maxWidth).
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

func GetCreationDataFromPrompt(workspaceNames []string, userGitProviders []types.GitProvider, manual bool, multiProject bool) (workspaceName string, projectRepositoryList []string, err error) {
	var projectRepoList []string
	var providerRepoUrl string

	if !manual && userGitProviders != nil && len(userGitProviders) > 0 {
		providerRepoUrl, err = GetRepositoryUrlFromWizard(userGitProviders, 0)
		if err != nil {
			return "", nil, err
		}
		if providerRepoUrl == "" {
			return "", nil, nil
		}
	}

	workspaceCreationPromptResponse, err := runInitialForm(providerRepoUrl, multiProject)
	if err != nil {
		return "", nil, err
	}

	if workspaceCreationPromptResponse.PrimaryRepository == "" {
		return "", nil, errors.New("primary repository is required")
	}

	projectRepoList = []string{workspaceCreationPromptResponse.PrimaryRepository}

	if workspaceCreationPromptResponse.SecondaryProjectCount > 0 {

		if !manual && userGitProviders != nil && len(userGitProviders) > 0 {
			for i := 0; i < workspaceCreationPromptResponse.SecondaryProjectCount; i++ {
				providerRepoUrl, err = GetRepositoryUrlFromWizard(userGitProviders, i+1)
				if err != nil {
					return "", nil, err
				}
				if providerRepoUrl == "" {
					return "", nil, nil
				}
				workspaceCreationPromptResponse.SecondaryRepositories = append(workspaceCreationPromptResponse.SecondaryRepositories, providerRepoUrl)
			}
		}

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

func GetRepositoryUrlFromWizard(userGitProviders []types.GitProvider, secondaryProjectOrder int) (string, error) {
	var providerId string
	var namespaceId string
	var branchName string
	var gitProvider gitprovider.GitProvider

	if len(userGitProviders) == 1 {
		providerId = userGitProviders[0].Id
	} else {

		availableGitProviderViews := config.GetGitProviderList()
		var gitProviderViewList []config.GitProvider

		for _, gitProvider := range userGitProviders {
			for _, availableGitProviderView := range availableGitProviderViews {
				if gitProvider.Id == availableGitProviderView.Id {
					gitProviderViewList = append(gitProviderViewList, availableGitProviderView)
				}
			}
		}
		providerId = selection.GetProviderIdFromPrompt(gitProviderViewList, secondaryProjectOrder)
		if providerId == "" {
			return "", nil
		}
	}

	gitProvider = gitprovider.GetGitProvider(providerId, userGitProviders)
	if gitProvider == nil {
		return "", errors.New("provider not found")
	}

	namespaceList, err := gitProvider.GetNamespaces()
	if err != nil {
		return "", err
	}

	if len(namespaceList) == 1 {
		namespaceId = namespaceList[0].Id
	} else {
		var namespaceViewList []gitprovider.GitNamespace
		namespaceViewList = append(namespaceViewList, namespaceList...)
		namespaceId = selection.GetNamespaceIdFromPrompt(namespaceViewList, secondaryProjectOrder)
		if namespaceId == "" {
			return "", nil
		}
	}

	repos, err := gitProvider.GetRepositories(namespaceId)
	if err != nil {
		return "", err
	}

	chosenRepo := selection.GetRepositoryFromPrompt(repos, secondaryProjectOrder)
	if chosenRepo == (gitprovider.GitRepository{}) {
		return "", nil
	}

	branchList, err := gitProvider.GetRepoBranches(chosenRepo, namespaceId)
	if err != nil {
		return "", err
	}

	if len(branchList) == 0 {
		return "", errors.New("no branches found")
	}

	if len(branchList) == 1 {
		branchName = branchList[0].Name
	} else {
		branchName = selection.GetBranchNameFromPrompt(branchList, secondaryProjectOrder)
		if branchName == "" {
			return "", nil
		}
	}

	return chosenRepo.Url, nil
}

func getSuggestedWorkspaceName(repo string) string {
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
