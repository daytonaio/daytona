// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package jobs

import (
	"context"
	"errors"
	"slices"

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

func (s *JobService) List(ctx context.Context, filter *stores.JobFilter) ([]*models.Job, error) {
	return s.jobStore.List(ctx, filter)
}

func (s *JobService) Find(ctx context.Context, filter *stores.JobFilter) (*models.Job, error) {
	return s.jobStore.Find(ctx, filter)
}

func (s *JobService) Create(ctx context.Context, j *models.Job) error {
	validAction, ok := validResourceActions[j.ResourceType]
	if !ok {
		return services.ErrInvalidResourceJobAction
	}

	if !slices.Contains(validAction, j.Action) {
		return services.ErrInvalidResourceJobAction
	}

	pendingJobs, err := s.List(ctx, &stores.JobFilter{
		ResourceId:   &j.ResourceId,
		ResourceType: &j.ResourceType,
		States:       &[]models.JobState{models.JobStatePending, models.JobStateRunning},
	})
	if err != nil {
		return err
	}

	if len(pendingJobs) > 0 {
		return stores.ErrJobInProgress
	}

	if j.Id == "" {
		id := stringid.GenerateRandomID()
		id = stringid.TruncateID(id)
		j.Id = id
	}
	return s.jobStore.Save(ctx, j)
}

func (s *JobService) SetState(ctx context.Context, jobId string, updateJobStateDto services.UpdateJobStateDTO) error {
	job, findErr := s.Find(ctx, &stores.JobFilter{
		Id: &jobId,
	})
	if findErr != nil {
		return findErr
	}

	if job.State == updateJobStateDto.State {
		return errors.New("job is already in the specified state")
	}

	job.State = updateJobStateDto.State
	job.Error = updateJobStateDto.ErrorMessage

	return s.jobStore.Save(ctx, job)
}

func (s *JobService) Delete(ctx context.Context, j *models.Job) error {
	return s.jobStore.Delete(ctx, j)
}

var validResourceActions = map[models.ResourceType][]models.JobAction{
	models.ResourceTypeWorkspace: {
		models.JobActionCreate,
		models.JobActionStart,
		models.JobActionStop,
		models.JobActionRestart,
		models.JobActionDelete,
		models.JobActionForceDelete,
	},
	models.ResourceTypeTarget: {
		models.JobActionCreate,
		models.JobActionStart,
		models.JobActionStop,
		models.JobActionRestart,
		models.JobActionDelete,
		models.JobActionForceDelete,
	},
	models.ResourceTypeBuild: {
		models.JobActionRun,
		models.JobActionDelete,
		models.JobActionForceDelete,
	},
	models.ResourceTypeRunner: {
		models.JobActionInstallProvider,
		models.JobActionUninstallProvider,
		models.JobActionUpdateProvider,
	},
}
