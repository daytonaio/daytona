// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

func generateEnvVarsList(envVars []apiclient.EnvironmentVariable, isMultipleSelect bool, action string) []list.Item {
	// Initialize an empty list of items.
	items := []list.Item{}

	// Populate items with titles and descriptions from envVars.
	for _, envVar := range envVars {
		newItem := item[apiclient.EnvironmentVariable]{
			key:            envVar.Key,
			desc:           envVar.Value,
			envVar:         &envVar,
			choiceProperty: envVar,
		}

		if isMultipleSelect {
			newItem.isMultipleSelect = true
			newItem.action = action
		}

		items = append(items, newItem)
	}

	return items
}

func getEnvVarsProgramEssentials(modelTitle string, actionVerb string, envVars []apiclient.EnvironmentVariable, footerText string, isMultipleSelect bool) tea.Model {

	items := generateEnvVarsList(envVars, isMultipleSelect, actionVerb)

	d := ItemDelegate[apiclient.EnvironmentVariable]{}

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	m := model[apiclient.EnvironmentVariable]{list: l}

	m.list.Title = views.GetStyledMainTitle(modelTitle + actionVerb)
	m.list.Styles.Title = lipgloss.NewStyle().Foreground(views.Green).Bold(true)
	m.footer = footerText

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()

	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	return p
}

func selectEnvironmentVariablePrompt(envVars []apiclient.EnvironmentVariable, actionVerb string, choiceChan chan<- *apiclient.EnvironmentVariable) {
	p := getEnvVarsProgramEssentials("Select an Environment Variable To ", actionVerb, envVars, "", false)
	if m, ok := p.(model[apiclient.EnvironmentVariable]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}

func GetEnvironmentVariableFromPrompt(envVars []apiclient.EnvironmentVariable, actionVerb string) *apiclient.EnvironmentVariable {
	choiceChan := make(chan *apiclient.EnvironmentVariable)

	go selectEnvironmentVariablePrompt(envVars, actionVerb, choiceChan)

	return <-choiceChan
}

func selectEnvironmentVariablesFromPrompt(envVars []apiclient.EnvironmentVariable, actionVerb string, choiceChan chan<- []*apiclient.EnvironmentVariable) {
	footerText := lipgloss.NewStyle().Bold(true).PaddingLeft(2).Render(fmt.Sprintf("\n\nPress 'x' to mark a server environment variable.\nPress 'enter' to %s the current/marked server environment variables.", actionVerb))
	p := getEnvVarsProgramEssentials("Select Server Environment Variables To ", actionVerb, envVars, footerText, true)

	m, ok := p.(model[apiclient.EnvironmentVariable])
	if ok && m.choices != nil {
		choiceChan <- m.choices
	} else if ok && m.choice != nil {
		choiceChan <- []*apiclient.EnvironmentVariable{m.choice}
	} else {
		choiceChan <- nil
	}
}

func GetEnvironmentVariablesFromPrompt(envVars []apiclient.EnvironmentVariable, actionVerb string) []*apiclient.EnvironmentVariable {
	choiceChan := make(chan []*apiclient.EnvironmentVariable)

	go selectEnvironmentVariablesFromPrompt(envVars, actionVerb, choiceChan)

	return <-choiceChan
}
