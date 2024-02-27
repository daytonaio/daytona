// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"

	"github.com/charmbracelet/huh"
)

func GetProvisionerFromPrompt(provisionerPlugins []serverapiclient.ProvisionerPlugin, description string, selectedProvisioner *serverapiclient.ProvisionerPlugin) (*serverapiclient.ProvisionerPlugin, error) {
	var provisioner = selectedProvisioner

	provisionerOptions := []huh.Option[*serverapiclient.ProvisionerPlugin]{}
	for _, provisioner := range provisionerPlugins {
		provisionerOptions = append(provisionerOptions, huh.NewOption(*provisioner.Name, &provisioner))
	}

	provisionerSelect := huh.NewSelect[*serverapiclient.ProvisionerPlugin]().
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
