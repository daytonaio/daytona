// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

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

const NewTargetName = "+ New Target"

type TargetView struct {
	Name         string
	Options      string
	IsDefault    bool
	ProviderInfo ProviderInfo
}

type ProviderInfo struct {
	Name      string
	Version   string
	Installed *bool
}

func GetTargetFromPrompt(targets []apiclient.ProviderTarget, activeProfileName string, providerViewList *[]provider.ProviderView, withNewTarget bool, actionVerb string) (*TargetView, error) {
	items := util.ArrayMap(targets, func(t apiclient.ProviderTarget) list.Item {
		return item{
			target: GetTargetViewFromTarget(t),
		}
	})

	if withNewTarget {
		name := NewTargetName
		options := "{}"
		items = append(items, item{
			target: TargetView{
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
				target: TargetView{
					Name:    fmt.Sprintf("Add a %s Provider Target", label),
					Options: "{}",
					ProviderInfo: ProviderInfo{
						Name:      providerView.Name,
						Version:   providerView.Version,
						Installed: providerView.Installed,
					},
				},
			})
		}
	}

	l := views.GetStyledSelectList(items)
	m := model{list: l}
	m.list.Title = views.GetStyledMainTitle("Choose a Target To " + actionVerb)
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

func GetTargetViewFromTarget(target apiclient.ProviderTarget) TargetView {
	return TargetView{
		Name:      target.Name,
		Options:   target.Options,
		IsDefault: target.IsDefault,
		ProviderInfo: ProviderInfo{
			Name:    target.ProviderInfo.Name,
			Version: target.ProviderInfo.Version,
		},
	}
}
