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
	currentTarget, err := s.GetTarget(ctx, &stores.TargetFilter{
		IdOrName: &id,
	}, services.TargetRetrievalParams{})
	if err != nil || currentTarget == nil {
		return err
	}

	defaultTarget, err := s.GetTarget(ctx, &stores.TargetFilter{
		Default: util.Pointer(true),
	}, services.TargetRetrievalParams{})
	if err != nil && !stores.IsTargetNotFound(err) {
		return err
	}

	if defaultTarget != nil {
		defaultTarget.IsDefault = false
		err := s.targetStore.Save(&defaultTarget.Target)
		if err != nil {
			return err
		}
	}

	currentTarget.IsDefault = true
	return s.targetStore.Save(&currentTarget.Target)
}
