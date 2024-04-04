// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/types"
)

func (p *Provisioner) DestroyWorkspace(workspace *types.Workspace, target *provider.ProviderTarget) error {
	targetProvider, err := manager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).DestroyWorkspace(&provider.WorkspaceRequest{
		TargetOptions: target.Options,
		Workspace:     workspace,
	})

	return err
}

func (p *Provisioner) DestroyProject(project *types.Project, target *provider.ProviderTarget) error {
	targetProvider, err := manager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).DestroyProject(&provider.ProjectRequest{
		TargetOptions: target.Options,
		Project:       project,
	})

	return err
}
