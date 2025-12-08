// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package provider

import (
	"context"

	"github.com/daytonaio/runner-manager/pkg/provider/types"
)

// IRunnerProvider defines the interface for managing runner instances
type IRunnerProvider interface {
	// AddRunners creates new runner instances
	AddRunners(ctx context.Context, instances int) (*types.AddRunnerResponse, error)

	// RemoveRunners removes runner instances
	RemoveRunners(ctx context.Context, instances int) error

	// ListRunners returns information about all runners
	ListRunners(ctx context.Context) ([]types.RunnerInfo, error)

	// GetRunner returns information about a specific runner
	GetRunner(ctx context.Context, runnerId string) (*types.RunnerInfo, error)

	// GetProviderName returns the name of the provider
	GetProviderName() string
}
