// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provisioner

import (
	"context"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/target"
)

type InfoResult struct {
	Info *target.TargetInfo
	Err  error
}

// Gets the target info from the provider - the context is used to cancel the request if it takes too long
func (p *Provisioner) GetTargetInfo(ctx context.Context, target *target.Target, targetConfig *provider.TargetConfig) (*target.TargetInfo, error) {
	ch := make(chan InfoResult, 1)

	go func() {
		targetProvider, err := p.providerManager.GetProvider(targetConfig.ProviderInfo.Name)
		if err != nil {
			ch <- InfoResult{nil, err}
			return
		}

		info, err := (*targetProvider).GetTargetInfo(&provider.TargetRequest{
			TargetConfigOptions: targetConfig.Options,
			Target:              target,
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
