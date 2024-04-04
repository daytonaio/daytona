// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/types"
)

func (p *Provisioner) GetWorkspaceInfo(workspace *types.Workspace, target *provider.ProviderTarget) (*types.WorkspaceInfo, error) {
	targetProvider, err := manager.GetProvider(target.ProviderInfo.Name)
	if err != nil {
		return nil, err
	}

	return (*targetProvider).GetWorkspaceInfo(&provider.WorkspaceRequest{
		TargetOptions: target.Options,
		Workspace:     workspace,
	})
}
