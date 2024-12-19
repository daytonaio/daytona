// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"errors"
	"io"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type ITargetService interface {
	CreateTarget(ctx context.Context, req CreateTargetDTO) (*models.Target, error)
	GetTarget(ctx context.Context, filter *stores.TargetFilter, params TargetRetrievalParams) (*TargetDTO, error)
	SaveTarget(ctx context.Context, target *models.Target) error
	ListTargets(ctx context.Context, filter *stores.TargetFilter, params TargetRetrievalParams) ([]TargetDTO, error)
	StartTarget(ctx context.Context, targetId string) error
	StopTarget(ctx context.Context, targetId string) error
	SetDefault(ctx context.Context, targetId string) error
	UpdateTargetProviderMetadata(ctx context.Context, targetId, metadata string) error
	RemoveTarget(ctx context.Context, targetId string) error
	ForceRemoveTarget(ctx context.Context, targetId string) error
	HandleSuccessfulCreation(ctx context.Context, targetId string) error

	GetTargetLogReader(ctx context.Context, targetId string) (io.Reader, error)
	GetTargetLogWriter(ctx context.Context, targetId string) (io.WriteCloser, error)
	SetTargetMetadata(ctx context.Context, targetId string, metadata *models.TargetMetadata) (*models.TargetMetadata, error)
}

type TargetDTO struct {
	models.Target
	State models.ResourceState `json:"state" validate:"required"`
} //	@name	TargetDTO

type CreateTargetDTO struct {
	Id               string `json:"id" validate:"required"`
	Name             string `json:"name" validate:"required"`
	TargetConfigName string `json:"targetConfigName" validate:"required"`
} //	@name	CreateTargetDTO

type UpdateTargetProviderMetadataDTO struct {
	Metadata string `json:"metadata" validate:"required"`
} // @name UpdateTargetProviderMetadataDTO

type TargetRetrievalParams struct {
	ShowDeleted bool
}

var (
	ErrTargetAlreadyExists = errors.New("target already exists")
	ErrInvalidTargetName   = errors.New("name is not a valid alphanumeric string")
	ErrTargetDeleted       = errors.New("target is deleted")
	ErrAgentlessTarget     = errors.New("provider uses an agentless target")
)

func IsTargetDeleted(err error) bool {
	return errors.Is(err, ErrTargetDeleted)
}
