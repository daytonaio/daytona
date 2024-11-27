// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package stores

import (
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

type JobStore interface {
	List(filter *JobFilter) ([]*models.Job, error)
	Find(filter *JobFilter) (*models.Job, error)
	Save(job *models.Job) error
	Delete(job *models.Job) error
}

type JobFilter struct {
	Id           *string
	ResourceType *models.ResourceType
	States       *[]models.JobState
	Actions      *[]models.JobAction
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
	ErrJobNotFound = errors.New("job not found")
)

func IsJobNotFound(err error) bool {
	return err.Error() == ErrJobNotFound.Error()
}
