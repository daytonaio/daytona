// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"context"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type TargetInfoResult struct {
	Info *target.TargetInfo
	Err  error
}

// Gets the target info from the provider - the context is used to cancel the request if it takes too long
func (p *Provisioner) GetTargetInfo(ctx context.Context, t *target.Target) (*target.TargetInfo, error) {
	ch := make(chan TargetInfoResult, 1)

	go func() {
		targetProvider, err := p.providerManager.GetProvider(t.ProviderInfo.Name)
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
	Info *workspace.WorkspaceInfo
	Err  error
}

// Gets the workspace info from the provider - the context is used to cancel the request if it takes too long
func (p *Provisioner) GetWorkspaceInfo(ctx context.Context, workspace *workspace.Workspace, t *target.Target) (*workspace.WorkspaceInfo, error) {
	ch := make(chan WorkspaceInfoResult, 1)

	go func() {
		targetProvider, err := p.providerManager.GetProvider(t.ProviderInfo.Name)
		if err != nil {
			ch <- WorkspaceInfoResult{nil, err}
			return
		}

		info, err := (*targetProvider).GetWorkspaceInfo(&provider.WorkspaceRequest{
			Target:    t,
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
