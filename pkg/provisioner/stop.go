// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider"
)

func (p *Provisioner) StopTarget(t *models.Target) error {
	targetProvider, err := p.providerManager.GetProvider(t.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).StopTarget(&provider.TargetRequest{
		Target: t,
	})

	return err
}

func (p *Provisioner) StopWorkspace(ws *models.Workspace) error {
	targetProvider, err := p.providerManager.GetProvider(ws.Target.TargetConfig.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).StopWorkspace(&provider.WorkspaceRequest{
		Workspace: ws,
	})

	return err
}
