// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package jobs

import (
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/docker/docker/pkg/stringid"
)

type JobServiceConfig struct {
	JobStore stores.JobStore
}

type JobService struct {
	jobStore stores.JobStore
}

func NewJobService(config JobServiceConfig) services.IJobService {
	return &JobService{
		jobStore: config.JobStore,
	}
}

func (s *JobService) List(filter *stores.JobFilter) ([]*models.Job, error) {
	return s.jobStore.List(filter)
}

func (s *JobService) Find(filter *stores.JobFilter) (*models.Job, error) {
	return s.jobStore.Find(filter)
}

func (s *JobService) Save(j *models.Job) error {
	if j.Id == "" {
		id := stringid.GenerateRandomID()
		id = stringid.TruncateID(id)
		j.Id = id
	}
	return s.jobStore.Save(j)
}

func (s *JobService) Delete(j *models.Job) error {
	return s.jobStore.Delete(j)
}
