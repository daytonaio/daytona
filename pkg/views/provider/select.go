// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
)

type ProviderView struct {
	Name                 string
	Label                *string
	Version              string
	Installed            *bool
	RunnerName           string
	RunnerId             string
	TargetConfigManifest map[string]apiclient.TargetConfigProperty
}

var NewProviderId = "+ New Provider"

func GetProviderFromPrompt(providers []ProviderView, title string, withNewProvider bool) (*ProviderView, error) {
	sortProviders(&providers)

	var items []list.Item

	for _, p := range providers {
		items = append(items, item{
			provider: p,
		})
	}

	if withNewProvider {
		name := NewProviderId
		items = append(items, item{
			provider: ProviderView{
				Name: name,
			},
		})
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = views.GetStyledMainTitle(title)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}

	if m, ok := p.(model); ok && m.choice != nil {
		return m.choice, nil
	}

	return nil, common.ErrCtrlCAbort
}

func ProviderListToView(providers []apiclient.ProviderInfo) []ProviderView {
	var providerViews []ProviderView

	for _, p := range providers {
		providerViews = append(providerViews, ProviderView{
			Name:                 p.Name,
			Label:                p.Label,
			Version:              p.Version,
			Installed:            nil,
			RunnerId:             p.RunnerId,
			RunnerName:           p.RunnerName,
			TargetConfigManifest: p.TargetConfigManifest,
		})
	}

	return providerViews
}

func sortProviders(providers *[]ProviderView) {
	sort.Slice(*providers, func(i, j int) bool {
		if (*providers)[i].Installed == nil {
			return false
		}
		if (*providers)[j].Installed == nil {
			return true
		}
		return *(*providers)[i].Installed
	})
}
