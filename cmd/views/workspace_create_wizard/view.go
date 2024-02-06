// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace_create_wizard

import (
	"dagent/cmd/views"
	"dagent/internal/util"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const maxWidth = 80

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
	formStep                        FormStep
	workspaceCreationPromptResponse WorkspaceCreationPromptResponse
}

type FormStep int

const (
	InitialForm FormStep = iota
	SecodaryProjectsForm
	WorkspaceNameForm
)

func InitialModel(workspaceCreationPromptResponseChan chan<- WorkspaceCreationPromptResponse) {
	m := Model{width: maxWidth, formStep: InitialForm}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	primaryRepo := ""

	secondaryProjectsCheck := false
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
		Value(&secondaryProjectsCheck)

	secondaryProjectCountPrompt := huh.NewInput().
		Title("How many secondary projects?").
		Value(&secondaryProjectsCountString).
		Validate(func(str string) error {
			_, err := strconv.Atoi(str) // Try to convert the input string to an integer
			if err != nil {
				return errors.New("Enter a valid number")
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
			return !secondaryProjectsCheck
		}),
	).WithTheme(dTheme).
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(true)

	p, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	secondaryProjectsCount, err := strconv.Atoi(secondaryProjectsCountString)
	if err != nil {
		secondaryProjectsCount = 0
	}

	if _, ok := p.(Model); ok {
		workspaceCreationPromptResponseChan <- WorkspaceCreationPromptResponse{
			PrimaryRepository:     primaryRepo,
			SecondaryProjectCount: secondaryProjectsCount,
		}
	} else {
		workspaceCreationPromptResponseChan <- WorkspaceCreationPromptResponse{
			PrimaryRepository:     "",
			SecondaryProjectCount: 0,
		}
	}
}

func SecondaryProjectsModel(workspaceCreationPromptResponse WorkspaceCreationPromptResponse, workspaceCreationPromptResponseChan chan<- WorkspaceCreationPromptResponse) {
	m := Model{width: maxWidth, workspaceCreationPromptResponse: workspaceCreationPromptResponse, formStep: SecodaryProjectsForm}
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
				return fmt.Errorf("Double-check and hit 'Yes'")
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

	p, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	for i := 0; i < count; i++ {
		secondaryRepoList[i], err = util.GetValidatedUrl(secondaryRepoList[i])
		if err != nil {
			log.Fatal("Invalid repository URL.", secondaryRepoList[i])
		}
	}

	result := workspaceCreationPromptResponse
	result.SecondaryRepositories = secondaryRepoList

	if _, ok := p.(Model); ok {
		workspaceCreationPromptResponseChan <- result
	} else {
		log.Fatal("Error running program:", err)
	}
}

func WorkspaceNameModel(workspaceCreationPromptResponse WorkspaceCreationPromptResponse, suggestedName string, workspaceNames []string, workspaceCreationPromptResponseChan chan<- WorkspaceCreationPromptResponse) {
	m := Model{width: maxWidth, workspaceCreationPromptResponse: workspaceCreationPromptResponse, formStep: WorkspaceNameForm}
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
						return errors.New("Workspace name already exists")
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

	p, err := tea.NewProgram(m).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	result := workspaceCreationPromptResponse
	result.WorkspaceName = workspaceName

	if _, ok := p.(Model); ok {
		workspaceCreationPromptResponseChan <- result
	} else {
		log.Fatal("Error running program:", err)
	}
}

func GetCreationDataFromPrompt(workspaceNames []string) (workspaceName string, projectRepositoryList []string) {
	var projectRepoList []string

	responseChan := make(chan WorkspaceCreationPromptResponse)

	go InitialModel(responseChan)

	workspaceCreationPromptResponse := <-responseChan

	if workspaceCreationPromptResponse.PrimaryRepository == "" {
		return "", nil
	}

	projectRepoList = []string{workspaceCreationPromptResponse.PrimaryRepository}

	if workspaceCreationPromptResponse.SecondaryProjectCount > 0 {

		responseChan = make(chan WorkspaceCreationPromptResponse)
		go SecondaryProjectsModel(workspaceCreationPromptResponse, responseChan)

		workspaceCreationPromptResponse = <-responseChan

		projectRepoList = append(projectRepoList, workspaceCreationPromptResponse.SecondaryRepositories...)
	}

	responseChan = make(chan WorkspaceCreationPromptResponse)
	suggestedName := getSuggestedWorkspaceName(workspaceCreationPromptResponse.PrimaryRepository)

	go WorkspaceNameModel(workspaceCreationPromptResponse, suggestedName, workspaceNames, responseChan)

	workspaceCreationPromptResponse = <-responseChan

	if workspaceCreationPromptResponse.WorkspaceName == "" {
		return "", nil
	}

	return workspaceCreationPromptResponse.WorkspaceName, projectRepoList
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

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.styles.Base.GetHorizontalFrameSize()
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "q":
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
		cmds = append(cmds, tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	s := m.styles

	// Main form
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)

	var status string

	if m.workspaceCreationPromptResponse.SecondaryProjectCount == 0 {
		const statusWidth = 80
		statusMarginLeft := m.width - statusWidth - lipgloss.Width(form) - s.Status.GetMarginRight()
		status = s.Status.Copy().
			Height(lipgloss.Height(form)).
			Width(statusWidth).
			MarginLeft(statusMarginLeft).
			Render(s.StatusHeader.Render("WORKSPACE NAME: ") + m.form.GetString("workspaceName") + "\n\n" +

				lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("fff")).Render("Primary project: ") +
				m.workspaceCreationPromptResponse.PrimaryRepository)
	} else {
		var secondaryRepoList []string

		if m.formStep == SecodaryProjectsForm {
			for i := 0; i < m.workspaceCreationPromptResponse.SecondaryProjectCount; i++ {
				key := fmt.Sprint("secondaryRepo", i)
				if m.form.GetString(key) != "" {
					validatedUrl, err := util.GetValidatedUrl(m.form.GetString(key))
					if err != nil {
						continue
					}
					secondaryRepoList = append(secondaryRepoList, "- "+validatedUrl)
				}
			}
		} else {
			for i := 0; i < len(m.workspaceCreationPromptResponse.SecondaryRepositories); i++ {
				secondaryRepoList = append(secondaryRepoList, "- "+m.workspaceCreationPromptResponse.SecondaryRepositories[i])
			}
		}

		const statusWidth = 80
		statusMarginLeft := m.width - statusWidth - lipgloss.Width(form) - s.Status.GetMarginRight()
		status = s.Status.Copy().
			Height(lipgloss.Height(form)).
			Width(statusWidth).
			MarginLeft(statusMarginLeft).
			Render(s.StatusHeader.Render("WORKSPACE NAME: ") + m.workspaceCreationPromptResponse.WorkspaceName + "\n\n" +

				lipgloss.NewStyle().Bold(true).Foreground(views.Green).Render("Primary project: ") + "\n- " +
				m.workspaceCreationPromptResponse.PrimaryRepository + "\n" + "\n" +

				lipgloss.NewStyle().Bold(true).Foreground(views.Green).Render("Secondary projects:") + "\n" +
				strings.Join(secondaryRepoList, "\n"))
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, form, status)

	switch m.form.State {
	case huh.StateCompleted:
		if m.formStep == InitialForm || m.formStep == SecodaryProjectsForm {
			return ""
		}
		if m.formStep == WorkspaceNameForm {
			return s.Base.Render(body + "\n")
		}
		return s.Base.Render(body + "\n\n")
	default:
		if m.formStep == InitialForm || m.formStep == WorkspaceNameForm {
			body = lipgloss.JoinHorizontal(lipgloss.Top, form, "")
		}

		errors := m.form.Errors()
		header := m.appBoundaryView("WORKSPACE CREATION")
		if len(errors) > 0 {
			header = m.appErrorBoundaryView(m.errorView())
		}

		footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
		if len(errors) > 0 {
			footer = m.appErrorBoundaryView("")
		}

		return s.Base.Render(header + "\n" + body + "\n\n" + footer)
	}
}

func (m Model) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m Model) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceForeground(views.Green),
	)
}

func (m Model) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceForeground(views.Green),
	)
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
