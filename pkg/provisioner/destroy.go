// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/workspace"
)

func (p *Provisioner) DestroyTarget(t *target.Target) error {
	targetProvider, err := p.providerManager.GetProvider(t.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).DestroyTarget(&provider.TargetRequest{
		Target: t,
	})

	return err
}

func (p *Provisioner) DestroyWorkspace(ws *workspace.Workspace, t *target.Target) error {
	targetProvider, err := p.providerManager.GetProvider(t.ProviderInfo.Name)
	if err != nil {
		return err
	}

	_, err = (*targetProvider).DestroyWorkspace(&provider.WorkspaceRequest{
		Target:    t,
		Workspace: ws,
	})

	return err
}
