// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
)

func (p *Provisioner) CreateTarget(target *target.Target, targetConfig *provider.TargetConfig) error {
	targetProvider, err := p.providerManager.GetProvider(targetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).CreateTarget(&provider.TargetRequest{
		TargetConfigOptions: targetConfig.Options,
		Target:              target,
	})

	return err
}

func (p *Provisioner) CreateProject(params ProjectParams) error {
	targetProvider, err := p.providerManager.GetProvider(params.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).CreateProject(&provider.ProjectRequest{
		TargetConfigOptions:      params.TargetConfig.Options,
		Project:                  params.Project,
		ContainerRegistry:        params.ContainerRegistry,
		GitProviderConfig:        params.GitProviderConfig,
		BuilderImage:             params.BuilderImage,
		BuilderContainerRegistry: params.BuilderImageContainerRegistry,
	})

	return err
}
