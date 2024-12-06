// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
)

func (s *TargetService) SetDefault(ctx context.Context, id string) error {
	var err error
	ctx, err = s.targetStore.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	defer stores.RecoverAndRollback(ctx, s.targetStore)

	currentTarget, err := s.GetTarget(ctx, &stores.TargetFilter{
		IdOrName: &id,
	}, services.TargetRetrievalParams{})
	if err != nil || currentTarget == nil {
		return s.targetStore.RollbackTransaction(ctx, err)
	}

	defaultTarget, err := s.GetTarget(ctx, &stores.TargetFilter{
		Default: util.Pointer(true),
	}, services.TargetRetrievalParams{})
	if err != nil && !stores.IsTargetNotFound(err) {
		return s.targetStore.RollbackTransaction(ctx, err)
	}

	if defaultTarget != nil {
		defaultTarget.IsDefault = false
		err := s.targetStore.Save(ctx, &defaultTarget.Target)
		if err != nil {
			return s.targetStore.RollbackTransaction(ctx, err)
		}
	}

	currentTarget.IsDefault = true
	err = s.targetStore.Save(ctx, &currentTarget.Target)
	if err != nil {
		return s.targetStore.RollbackTransaction(ctx, err)
	}

	return s.targetStore.CommitTransaction(ctx)
}
