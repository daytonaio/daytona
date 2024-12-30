// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/workspace"
)

func (p *Provisioner) CreateWorkspace(workspace *workspace.Workspace, target *provider.ProviderTarget) error {
	targetProvider, err := p.providerManager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).CreateWorkspace(&provider.WorkspaceRequest{
		TargetOptions: target.Options,
		Workspace:     workspace,
	})

	return err
}

func (p *Provisioner) CreateProject(params ProjectParams) error {
	targetProvider, err := p.providerManager.GetProvider(params.Target.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).CreateProject(&provider.ProjectRequest{
		TargetOptions:            params.Target.Options,
		Project:                  params.Project,
		ContainerRegistry:        params.ContainerRegistry,
		GitProviderConfig:        params.GitProviderConfig,
		BuilderImage:             params.BuilderImage,
		BuilderContainerRegistry: params.BuilderImageContainerRegistry,
	})

	return err
}
