// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/target/project"
)

func (p *Provisioner) DestroyTarget(target *target.Target, targetConfig *provider.TargetConfig) error {
	targetProvider, err := p.providerManager.GetProvider(targetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).DestroyTarget(&provider.TargetRequest{
		TargetConfigOptions: targetConfig.Options,
		Target:              target,
	})

	return err
}

func (p *Provisioner) DestroyProject(proj *project.Project, targetConfig *provider.TargetConfig) error {
	targetProvider, err := p.providerManager.GetProvider(targetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).DestroyProject(&provider.ProjectRequest{
		TargetConfigOptions: targetConfig.Options,
		Project:             proj,
	})

	return err
}
