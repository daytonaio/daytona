// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runners

import (
	"context"

	"github.com/daytonaio/daytona/pkg/stores"
)

func (s *RunnerService) RemoveRunner(ctx context.Context, runnerId string) error {
	var err error
	ctx, err = s.runnerStore.BeginTransaction(ctx)
	if err != nil {
		return s.runnerStore.RollbackTransaction(ctx, err)
	}

	defer stores.RecoverAndRollback(ctx, s.runnerStore)

	runner, err := s.runnerStore.Find(ctx, runnerId)
	if err != nil {
		return s.runnerStore.RollbackTransaction(ctx, err)
	}

	err = s.runnerStore.Delete(ctx, runner)
	if err != nil {
		return s.runnerStore.RollbackTransaction(ctx, err)
	}

	metadata, err := s.runnerMetadataStore.Find(ctx, runnerId)
	if err != nil && !stores.IsRunnerMetadataNotFound(err) {
		return s.runnerStore.RollbackTransaction(ctx, err)
	}
	if metadata != nil {
		err = s.runnerMetadataStore.Delete(ctx, metadata)
		if err != nil {
			return s.runnerStore.RollbackTransaction(ctx, err)
		}
	}

	err = s.revokeApiKey(ctx, runner.Name)
	if err != nil {
		return s.runnerStore.RollbackTransaction(ctx, err)
	}

	err = s.unsetDefaultTarget(ctx, runner.Id)
	if err != nil {
		return s.runnerStore.RollbackTransaction(ctx, err)
	}

	return s.runnerStore.CommitTransaction(ctx)
}
