// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"io"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/stores"
)

type ITargetService interface {
	CreateTarget(ctx context.Context, req dto.CreateTargetDTO) (*models.Target, error)
	GetTarget(ctx context.Context, filter *stores.TargetFilter, verbose bool) (*dto.TargetDTO, error)
	GetTargetLogReader(targetId string) (io.Reader, error)
	ListTargets(ctx context.Context, filter *stores.TargetFilter, verbose bool) ([]dto.TargetDTO, error)
	StartTarget(ctx context.Context, targetId string) error
	StopTarget(ctx context.Context, targetId string) error
	SetDefault(ctx context.Context, targetId string) error
	RemoveTarget(ctx context.Context, targetId string) error
	ForceRemoveTarget(ctx context.Context, targetId string) error
}
