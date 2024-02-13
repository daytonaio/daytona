// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views_provisioner

import (
	"github.com/daytonaio/daytona/common/grpc/proto"

	"github.com/charmbracelet/huh"
)

func GetProvisionerFromPrompt(provisionerPlugins []*proto.ProvisionerPlugin, description string, selectedProvisioner *proto.ProvisionerPlugin) (*proto.ProvisionerPlugin, error) {
	var provisioner = selectedProvisioner

	provisionerOptions := []huh.Option[*proto.ProvisionerPlugin]{}
	for _, provisioner := range provisionerPlugins {
		provisionerOptions = append(provisionerOptions, huh.NewOption(provisioner.Name, provisioner))
	}

	provisionerSelect := huh.NewSelect[*proto.ProvisionerPlugin]().
		Title("Provisioner").
		Options(provisionerOptions...).
		Value(&provisioner)

	if description != "" {
		provisionerSelect.Description(description)
	}

	form := huh.NewForm(
		huh.NewGroup(
			provisionerSelect,
		))

	err := form.Run()
	if err != nil {
		return nil, err
	}

	return provisioner, nil
}
