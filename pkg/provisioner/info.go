// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/provider"
)

type TargetInfoResult struct {
	Info *models.TargetInfo
	Err  error
}

// Gets the target info from the provider - the context is used to cancel the request if it takes too long
func (p *Provisioner) GetTargetInfo(ctx context.Context, t *models.Target) (*models.TargetInfo, error) {
	ch := make(chan TargetInfoResult, 1)

	go func() {
		targetProvider, err := p.providerManager.GetProvider(t.TargetConfig.ProviderInfo.Name)
		if err != nil {
			ch <- TargetInfoResult{nil, err}
			return
		}

		info, err := (*targetProvider).GetTargetInfo(&provider.TargetRequest{
			Target: t,
		})

		ch <- TargetInfoResult{info, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case data := <-ch:
		return data.Info, data.Err
	}
}

type WorkspaceInfoResult struct {
	Info *models.WorkspaceInfo
	Err  error
}

// Gets the workspace info from the provider - the context is used to cancel the request if it takes too long
func (p *Provisioner) GetWorkspaceInfo(ctx context.Context, workspace *models.Workspace) (*models.WorkspaceInfo, error) {
	ch := make(chan WorkspaceInfoResult, 1)

	go func() {
		targetProvider, err := p.providerManager.GetProvider(workspace.Target.TargetConfig.ProviderInfo.Name)
		if err != nil {
			ch <- WorkspaceInfoResult{nil, err}
			return
		}

		info, err := (*targetProvider).GetWorkspaceInfo(&provider.WorkspaceRequest{
			Workspace: workspace,
		})

		ch <- WorkspaceInfoResult{info, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case data := <-ch:
		return data.Info, data.Err
	}
}
