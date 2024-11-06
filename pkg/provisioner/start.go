// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
)

func (p *Provisioner) StartTarget(t *target.Target) error {
	targetProvider, err := p.providerManager.GetProvider(t.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).StartTarget(&provider.TargetRequest{
		Target: t,
	})

	return err
}

func (p *Provisioner) StartWorkspace(params WorkspaceParams) error {
	targetProvider, err := p.providerManager.GetProvider(params.Target.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).StartWorkspace(&provider.WorkspaceRequest{
		Target:                   params.Target,
		Workspace:                params.Workspace,
		ContainerRegistry:        params.ContainerRegistry,
		GitProviderConfig:        params.GitProviderConfig,
		BuilderImage:             params.BuilderImage,
		BuilderContainerRegistry: params.BuilderImageContainerRegistry,
	})

	return err
}
