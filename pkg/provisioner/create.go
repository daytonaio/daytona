// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/project"
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

func (p *Provisioner) CreateProject(proj *project.Project, targetConfig *provider.TargetConfig, cr *containerregistry.ContainerRegistry, gc *gitprovider.GitProviderConfig) error {
	targetProvider, err := p.providerManager.GetProvider(targetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).CreateProject(&provider.ProjectRequest{
		TargetConfigOptions: targetConfig.Options,
		Project:             proj,
		ContainerRegistry:   cr,
		GitProviderConfig:   gc,
	})

	return err
}
