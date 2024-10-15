// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var NewGitProviderConfigIdentifier = "<NEW_GIT_PROVIDER_CONFIG>"

func GetGitProviderConfigFromPrompt(gitProviderConfigs []apiclient.GitProvider, withNewGitProviderConfig bool, actionVerb string) *apiclient.GitProvider {
	choiceChan := make(chan *apiclient.GitProvider)
	go selectGitProviderConfigPrompt(gitProviderConfigs, withNewGitProviderConfig, actionVerb, choiceChan)
	return <-choiceChan
}

func selectGitProviderConfigPrompt(gitProviderConfigs []apiclient.GitProvider, withNewGitProviderConfig bool, actionVerb string, choiceChan chan<- *apiclient.GitProvider) {
	items := []list.Item{}

	supportedGitProviders := config.GetSupportedGitProviders()

	for _, gp := range gitProviderConfigs {
		title := fmt.Sprintf("%s (%s)", gp.ProviderId, gp.Alias)

		for _, provider := range supportedGitProviders {
			if provider.Id == gp.ProviderId {
				title = fmt.Sprintf("%s (%s)", provider.Name, gp.Alias)
			}
		}

		desc := gp.Id

		if gp.BaseApiUrl != nil && *gp.BaseApiUrl != "" {
			desc = fmt.Sprintf("%s - %s", desc, *gp.BaseApiUrl)
		}

		newItem := item[apiclient.GitProvider]{title: title, desc: desc, choiceProperty: gp}
		items = append(items, newItem)
	}

	if withNewGitProviderConfig {
		newItem := item[apiclient.GitProvider]{title: "+ Add a new Git provider", desc: "", choiceProperty: apiclient.GitProvider{
			Id: NewGitProviderConfigIdentifier,
		}}
		items = append(items, newItem)
	}

	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(views.Green).
		Foreground(views.Green).
		Bold(true).
		Padding(0, 0, 0, 1)

	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Foreground(views.DimmedGreen)

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	title := "Select a Git Provider Config To " + actionVerb

	l.Title = views.GetStyledMainTitle(title)
	l.Styles.Title = titleStyle

	m := model[apiclient.GitProvider]{list: l}

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if m, ok := p.(model[apiclient.GitProvider]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}
