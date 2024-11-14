// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider"
)

func (p *Provisioner) CreateTarget(t *models.Target) error {
	targetProvider, err := p.providerManager.GetProvider(t.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).CreateTarget(&provider.TargetRequest{
		Target: t,
	})

	return err
}

func (p *Provisioner) CreateWorkspace(ws *models.Workspace, cr *models.ContainerRegistry, gc *models.GitProviderConfig) error {
	targetProvider, err := p.providerManager.GetProvider(ws.Target.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).CreateWorkspace(&provider.WorkspaceRequest{
		Workspace:         ws,
		ContainerRegistry: cr,
		GitProviderConfig: gc,
	})

	return err
}
