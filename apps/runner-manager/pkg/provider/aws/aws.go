// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package aws

import (
	"context"
	"errors"

	"github.com/daytonaio/runner-manager/pkg/provider/types"
)

type AwsProvider struct {
	// TODO: Add AWS client and configuration
}

// NewAwsProvider creates a new AWS provider instance
func NewAwsProvider() *AwsProvider {
	return &AwsProvider{}
}

func (p *AwsProvider) AddRunners(ctx context.Context, instances int) (*types.AddRunnerResponse, error) {
	return nil, errors.New("AddRunners not implemented for AWS provider")
}

func (p *AwsProvider) RemoveRunners(ctx context.Context, instances int) error {
	return errors.New("RemoveRunners not implemented for AWS provider")
}

func (p *AwsProvider) ListRunners(ctx context.Context) ([]types.RunnerInfo, error) {
	return nil, errors.New("ListRunners not implemented for AWS provider")
}

func (p *AwsProvider) GetRunner(ctx context.Context, runnerId string) (*types.RunnerInfo, error) {
	return nil, errors.New("GetRunner not implemented for AWS provider")
}

func (p *AwsProvider) GetProviderName() string {
	return "aws"
}
