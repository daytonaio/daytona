// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runners

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

func (s *RunnerService) RegisterRunner(ctx context.Context, req services.RegisterRunnerDTO) (*services.RunnerDTO, error) {
	var err error
	ctx, err = s.runnerStore.BeginTransaction(ctx)
	if err != nil {
		return nil, s.runnerStore.RollbackTransaction(ctx, err)
	}

	defer stores.RecoverAndRollback(ctx, s.runnerStore)

	_, err = s.runnerStore.Find(ctx, req.Name)
	if err == nil {
		return nil, s.runnerStore.RollbackTransaction(ctx, services.ErrRunnerAlreadyExists)
	}

	apiKey, err := s.generateApiKey(ctx, req.Id)
	if err != nil {
		return nil, s.runnerStore.RollbackTransaction(ctx, err)
	}

	runner := &models.Runner{
		Id:     req.Id,
		Name:   req.Name,
		ApiKey: apiKey,
		Metadata: &models.RunnerMetadata{
			RunnerId: req.Id,
			Uptime:   0,
		},
	}

	err = s.runnerStore.Save(ctx, runner)
	if err != nil {
		return nil, s.runnerStore.RollbackTransaction(ctx, err)
	}

	err = s.runnerMetadataStore.Save(ctx, runner.Metadata)
	if err != nil {
		return nil, s.runnerStore.RollbackTransaction(ctx, err)
	}

	err = s.runnerStore.CommitTransaction(ctx)
	if err != nil {
		return nil, s.runnerStore.RollbackTransaction(ctx, err)
	}

	return &services.RunnerDTO{
		Runner: *runner,
		State:  runner.GetState(),
	}, nil
}
