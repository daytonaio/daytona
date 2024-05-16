// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/workspace"
)

func (p *Provisioner) GetWorkspaceInfo(workspace *workspace.Workspace, target *provider.ProviderTarget) (*workspace.WorkspaceInfo, error) {
	targetProvider, err := p.providerManager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return nil, err
	}

	return (*targetProvider).GetWorkspaceInfo(&provider.WorkspaceRequest{
		TargetOptions: target.Options,
		Workspace:     workspace,
	})
}
