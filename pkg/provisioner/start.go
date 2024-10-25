// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/workspace"
)

func (p *Provisioner) StartWorkspace(workspace *workspace.Workspace, targetConfig *provider.TargetConfig) error {
	targetProvider, err := p.providerManager.GetProvider(targetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).StartWorkspace(&provider.WorkspaceRequest{
		TargetConfigOptions: targetConfig.Options,
		Workspace:           workspace,
	})

	return err
}

func (p *Provisioner) StartProject(params ProjectParams) error {
	targetProvider, err := p.providerManager.GetProvider(params.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).StartProject(&provider.ProjectRequest{
		TargetConfigOptions:      params.TargetConfig.Options,
		Project:                  params.Project,
		ContainerRegistry:        params.ContainerRegistry,
		GitProviderConfig:        params.GitProviderConfig,
		BuilderImage:             params.BuilderImage,
		BuilderContainerRegistry: params.BuilderImageContainerRegistry,
	})

	return err
}
