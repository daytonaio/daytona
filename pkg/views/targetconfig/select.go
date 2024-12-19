// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/provider"
)

const NewTargetConfigName = "+ New Target Config"

type TargetConfigView struct {
	Id           string
	Name         string
	RunnerName   string
	Options      string
	ProviderInfo ProviderInfo
}

type ProviderInfo struct {
	Name       string
	RunnerId   string
	RunnerName string
	Version    string
	Label      *string
	Installed  *bool
}

func GetTargetConfigFromPrompt(targetConfigs []apiclient.TargetConfig, activeProfileName string, providerViewList *[]provider.ProviderView, withNewTargetConfig bool, actionVerb string) (*TargetConfigView, error) {
	items := util.ArrayMap(targetConfigs, func(t apiclient.TargetConfig) list.Item {
		return item{
			targetConfig: ToTargetConfigView(t),
		}
	})

	if withNewTargetConfig {
		name := NewTargetConfigName
		options := "{}"
		items = append(items, item{
			targetConfig: TargetConfigView{
				Name:    name,
				Options: options,
			},
		})
	}

	// Display options for providers that are not installed
	if providerViewList != nil {
		for _, providerView := range *providerViewList {
			if providerView.Installed != nil && *providerView.Installed {
				continue
			}
			var label string
			if providerView.Label != nil {
				label = *providerView.Label
			} else {
				label = providerView.Name
			}

			items = append(items, item{
				targetConfig: TargetConfigView{
					Name:    fmt.Sprintf("Add a %s Target Config", label),
					Options: "{}",
					ProviderInfo: ProviderInfo{
						Name:       providerView.Name,
						RunnerId:   providerView.RunnerId,
						RunnerName: providerView.RunnerName,
						Version:    providerView.Version,
						Label:      providerView.Label,
						Installed:  providerView.Installed,
					},
				},
			})
		}
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = views.GetStyledMainTitle("Choose a Target Config To " + actionVerb)
	m.footer = views.GetListFooter(activeProfileName, views.DefaultListFooterPadding)

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return nil, err
	}

	if m, ok := p.(model); ok && m.choice != nil {
		return m.choice, nil
	}

	return nil, common.ErrCtrlCAbort
}

func ToTargetConfigView(targetConfig apiclient.TargetConfig) TargetConfigView {
	return TargetConfigView{
		Id:         targetConfig.Id,
		Name:       targetConfig.Name,
		RunnerName: targetConfig.ProviderInfo.RunnerName,
		Options:    targetConfig.Options,
		ProviderInfo: ProviderInfo{
			Name:       targetConfig.ProviderInfo.Name,
			RunnerId:   targetConfig.ProviderInfo.RunnerId,
			RunnerName: targetConfig.ProviderInfo.RunnerName,
			Version:    targetConfig.ProviderInfo.Version,
			Label:      targetConfig.ProviderInfo.Label,
		},
	}
}
