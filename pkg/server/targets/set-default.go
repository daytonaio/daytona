// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targets

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/target"
)

func (s *TargetService) SetDefault(ctx context.Context, id string) error {
	currentTarget, err := s.GetTarget(ctx, &target.TargetFilter{
		IdOrName: &id,
	}, false)
	if err != nil || currentTarget == nil {
		return err
	}

	defaultConfig, err := s.GetTarget(ctx, &target.TargetFilter{
		Default: util.Pointer(true),
	}, false)
	if err != nil && !target.IsTargetNotFound(err) {
		return err
	}

	if defaultConfig != nil {
		defaultConfig.IsDefault = false
		err := s.targetStore.Save(TargetDtoToTarget(*defaultConfig))
		if err != nil {
			return err
		}
	}

	currentTarget.IsDefault = true
	return s.targetStore.Save(TargetDtoToTarget(*currentTarget))
}

func TargetDtoToTarget(targetDto dto.TargetDTO) *target.Target {
	return &target.Target{
		Id:           targetDto.Id,
		Name:         targetDto.Name,
		ProviderInfo: targetDto.ProviderInfo,
		Options:      targetDto.Options,
		IsDefault:    targetDto.IsDefault,
	}
}
