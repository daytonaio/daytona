// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/huh"
)

func GetProviderFromPrompt(providerPlugins []serverapiclient.ProviderPlugin, description string, selectedProvider *serverapiclient.ProviderPlugin) (*serverapiclient.ProviderPlugin, error) {
	var provider = selectedProvider

	providerOptions := []huh.Option[*serverapiclient.ProviderPlugin]{}
	for _, provider := range providerPlugins {
		providerOptions = append(providerOptions, huh.NewOption(*provider.Name, &provider))
	}

	providerSelect := huh.NewSelect[*serverapiclient.ProviderPlugin]().
		Title("Default provider").
		Options(providerOptions...).
		Value(&provider)

	if description != "" {
		providerSelect.Description(description)
	}

	form := huh.NewForm(
		huh.NewGroup(
			providerSelect,
		)).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		return nil, err
	}

	return provider, nil
}
