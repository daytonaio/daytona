// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider"
)

func (p *Provisioner) StartTarget(t *models.Target) error {
	targetProvider, err := p.providerManager.GetProvider(t.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).StartTarget(&provider.TargetRequest{
		Target: t,
	})

	return err
}

func (p *Provisioner) StartWorkspace(params WorkspaceParams) error {
	targetProvider, err := p.providerManager.GetProvider(params.Workspace.Target.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).StartWorkspace(&provider.WorkspaceRequest{
		Workspace:                params.Workspace,
		ContainerRegistry:        params.ContainerRegistry,
		GitProviderConfig:        params.GitProviderConfig,
		BuilderImage:             params.BuilderImage,
		BuilderContainerRegistry: params.BuilderImageContainerRegistry,
	})

	return err
}
