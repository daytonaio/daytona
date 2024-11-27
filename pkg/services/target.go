// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/server/targets/dto"
	"github.com/daytonaio/daytona/pkg/stores"
)

type ITargetService interface {
	CreateTarget(ctx context.Context, req dto.CreateTargetDTO) (*models.Target, error)
	GetTarget(ctx context.Context, filter *stores.TargetFilter, params TargetRetrievalParams) (*dto.TargetDTO, error)
	GetTargetLogReader(targetId string) (io.Reader, error)
	ListTargets(ctx context.Context, filter *stores.TargetFilter, params TargetRetrievalParams) ([]dto.TargetDTO, error)
	StartTarget(ctx context.Context, targetId string) error
	StopTarget(ctx context.Context, targetId string) error
	SetDefault(ctx context.Context, targetId string) error
	RemoveTarget(ctx context.Context, targetId string) error
	ForceRemoveTarget(ctx context.Context, targetId string) error
	HandleSuccessfulCreation(ctx context.Context, targetId string) error
	HandleSuccessfulRemoval(ctx context.Context, targetId string) error

	SetTargetMetadata(targetId string, metadata *models.TargetMetadata) (*models.TargetMetadata, error)
}

type TargetRetrievalParams struct {
	Verbose     bool
	ShowDeleted bool
}

var (
	ErrTargetAlreadyExists = errors.New("target already exists")
	ErrInvalidTargetName   = errors.New("name is not a valid alphanumeric string")
	ErrTargetDeleted       = errors.New("target is deleted")
)

func IsTargetDeleted(err error) bool {
	return errors.Is(err, ErrTargetDeleted)
}
