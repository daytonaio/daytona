// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type JobStore interface {
	IStore
	List(ctx context.Context, filter *JobFilter) ([]*models.Job, error)
	Find(ctx context.Context, filter *JobFilter) (*models.Job, error)
	Save(ctx context.Context, job *models.Job) error
	Delete(ctx context.Context, job *models.Job) error
}

type JobFilter struct {
	Id              *string
	ResourceId      *string
	RunnerIdOrIsNil *string
	ResourceType    *models.ResourceType
	States          *[]models.JobState
	Actions         *[]models.JobAction
}

func (f *JobFilter) StatesToInterface() []interface{} {
	args := make([]interface{}, len(*f.States))
	for i, v := range *f.States {
		args[i] = v
	}
	return args
}

func (f *JobFilter) ActionsToInterface() []interface{} {
	args := make([]interface{}, len(*f.Actions))
	for i, v := range *f.Actions {
		args[i] = v
	}
	return args
}

var (
	ErrJobNotFound   = errors.New("job not found")
	ErrJobInProgress = errors.New("another job is in progress")
)

func IsJobNotFound(err error) bool {
	return err.Error() == ErrJobNotFound.Error()
}

func IsJobInProgress(err error) bool {
	return err.Error() == ErrJobInProgress.Error()
}
