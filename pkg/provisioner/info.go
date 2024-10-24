// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"context"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type InfoResult struct {
	Info *workspace.WorkspaceInfo
	Err  error
}

// Gets the workspace info from the provider - the context is used to cancel the request if it takes too long
func (p *Provisioner) GetWorkspaceInfo(ctx context.Context, ws *workspace.Workspace, targetConfig *provider.TargetConfig) (*workspace.WorkspaceInfo, error) {
	ch := make(chan InfoResult, 1)

	go func() {
		targetProvider, err := p.providerManager.GetProvider(targetConfig.ProviderInfo.Name)
		if err != nil {
			ch <- InfoResult{nil, err}
			return
		}

		info, err := (*targetProvider).GetWorkspaceInfo(&provider.WorkspaceRequest{
			TargetConfigOptions: targetConfig.Options,
			Workspace:           ws,
		})

		ch <- InfoResult{info, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case data := <-ch:
		return data.Info, data.Err
	}
}
