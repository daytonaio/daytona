// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views_provisioner

import (
	"github.com/daytonaio/daytona/cli/cmd/views"
	"github.com/daytonaio/daytona/common/api_client"

	"github.com/charmbracelet/huh"
)

func GetProvisionerFromPrompt(provisionerPlugins []api_client.ProvisionerPlugin, description string, selectedProvisioner *api_client.ProvisionerPlugin) (*api_client.ProvisionerPlugin, error) {
	var provisioner = selectedProvisioner

	provisionerOptions := []huh.Option[*api_client.ProvisionerPlugin]{}
	for _, provisioner := range provisionerPlugins {
		provisionerOptions = append(provisionerOptions, huh.NewOption(*provisioner.Name, &provisioner))
	}

	provisionerSelect := huh.NewSelect[*api_client.ProvisionerPlugin]().
		Title("Default provisioner").
		Options(provisionerOptions...).
		Value(&provisioner)

	if description != "" {
		provisionerSelect.Description(description)
	}

	form := huh.NewForm(
		huh.NewGroup(
			provisionerSelect,
		)).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		return nil, err
	}

	return provisioner, nil
}
