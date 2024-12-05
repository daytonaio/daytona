// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/stores"
)

type IJobService interface {
	Create(ctx context.Context, job *models.Job) error
	SetState(ctx context.Context, jobId string, state models.JobState, error *string) error
	Find(ctx context.Context, filter *stores.JobFilter) (*models.Job, error)
	List(ctx context.Context, filter *stores.JobFilter) ([]*models.Job, error)
	Delete(ctx context.Context, job *models.Job) error
}

var (
	ErrInvalidResourceJobAction = errors.New("invalid job action for resource")
)

func IsInvalidResourceJobAction(err error) bool {
	return err.Error() == ErrInvalidResourceJobAction.Error()
}
