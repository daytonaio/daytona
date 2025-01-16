// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider_install

import (
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
)

type ProviderInstallView struct {
	Name    string
	Label   *string
	Version string
}

var NewProviderId = "+ New Provider"

func GetProviderFromInstallPrompt(providers []ProviderInstallView, title string, withNewProvider bool) (*ProviderInstallView, error) {
	sortProvidersForInstall(providers)

	var items []list.Item

	for _, p := range providers {
		items = append(items, item{
			provider: p,
		})
	}

	if withNewProvider {
		name := NewProviderId
		items = append(items, item{
			provider: ProviderInstallView{
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

func ProviderInstallListToView(providers []apiclient.ProviderDTO) []ProviderInstallView {
	var providerViews []ProviderInstallView

	for _, p := range providers {
		providerViews = append(providerViews, ProviderInstallView{
			Name:    p.Name,
			Label:   p.Label,
			Version: p.Version,
		})
	}

	return providerViews
}

func sortProvidersForInstall(providers []ProviderInstallView) {
	sort.Slice(providers, func(i, j int) bool {
		return (providers)[i].Name > (providers)[j].Name
	})
}
