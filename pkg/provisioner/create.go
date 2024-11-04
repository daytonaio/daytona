// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/workspace"
)

func (p *Provisioner) CreateTarget(t *target.Target) error {
	targetProvider, err := p.providerManager.GetProvider(t.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).CreateTarget(&provider.TargetRequest{
		Target: t,
	})

	return err
}

func (p *Provisioner) CreateWorkspace(ws *workspace.Workspace, t *target.Target, cr *containerregistry.ContainerRegistry, gc *gitprovider.GitProviderConfig) error {
	targetProvider, err := p.providerManager.GetProvider(t.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).CreateWorkspace(&provider.WorkspaceRequest{
		Target:            t,
		Workspace:         ws,
		ContainerRegistry: cr,
		GitProviderConfig: gc,
	})

	return err
}
