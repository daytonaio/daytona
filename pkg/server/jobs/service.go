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
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/docker/docker/pkg/stringid"

	log "github.com/sirupsen/logrus"
)

type JobServiceConfig struct {
	JobStore            stores.JobStore
	TrackTelemetryEvent func(event telemetry.Event, clientId string) error

	UpdateWorkspaceLastJob func(ctx context.Context, workspaceId string, jobId string) error
	UpdateTargetLastJob    func(ctx context.Context, targetId string, jobId string) error
	UpdateBuildLastJob     func(ctx context.Context, buildId string, jobId string) error
}

type JobService struct {
	jobStore            stores.JobStore
	trackTelemetryEvent func(event telemetry.Event, clientId string) error

	updateWorkspaceLastJob func(ctx context.Context, workspaceId string, jobId string) error
	updateTargetLastJob    func(ctx context.Context, targetId string, jobId string) error
	updateBuildLastJob     func(ctx context.Context, buildId string, jobId string) error
}

func NewJobService(config JobServiceConfig) services.IJobService {
	return &JobService{
		jobStore:               config.JobStore,
		trackTelemetryEvent:    config.TrackTelemetryEvent,
		updateWorkspaceLastJob: config.UpdateWorkspaceLastJob,
		updateTargetLastJob:    config.UpdateTargetLastJob,
		updateBuildLastJob:     config.UpdateBuildLastJob,
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
		return s.handleCreateError(ctx, j, services.ErrInvalidResourceJobAction)
	}

	if !slices.Contains(validAction, j.Action) {
		return s.handleCreateError(ctx, j, services.ErrInvalidResourceJobAction)
	}

	pendingJobs, err := s.List(ctx, &stores.JobFilter{
		ResourceId:   &j.ResourceId,
		ResourceType: &j.ResourceType,
		States:       &[]models.JobState{models.JobStatePending, models.JobStateRunning},
	})
	if err != nil {
		return s.handleCreateError(ctx, j, err)
	}

	if len(pendingJobs) > 0 {
		return s.handleCreateError(ctx, j, stores.ErrJobInProgress)
	}

	if j.Id == "" {
		id := stringid.GenerateRandomID()
		id = stringid.TruncateID(id)
		j.Id = id
	}

	err = s.jobStore.Save(ctx, j)
	return s.handleCreateError(ctx, j, err)
}

func (s *JobService) UpdateState(ctx context.Context, jobId string, updateJobStateDto services.UpdateJobStateDTO) error {
	var err error
	ctx, err = s.jobStore.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	defer stores.RecoverAndRollback(ctx, s.jobStore)

	job, findErr := s.Find(ctx, &stores.JobFilter{
		Id: &jobId,
	})
	if findErr != nil {
		return s.jobStore.RollbackTransaction(ctx, findErr)
	}

	if job.State == updateJobStateDto.State {
		return s.jobStore.RollbackTransaction(ctx, errors.New("job is already in the specified state"))
	}

	job.State = updateJobStateDto.State
	job.Error = updateJobStateDto.ErrorMessage

	err = s.jobStore.Save(ctx, job)
	if err != nil {
		return s.jobStore.RollbackTransaction(ctx, err)
	}

	switch job.ResourceType {
	case models.ResourceTypeWorkspace:
		err = s.updateWorkspaceLastJob(ctx, job.ResourceId, job.Id)
	case models.ResourceTypeTarget:
		err = s.updateTargetLastJob(ctx, job.ResourceId, job.Id)
	case models.ResourceTypeBuild:
		err = s.updateBuildLastJob(ctx, job.ResourceId, job.Id)
	}

	if err != nil {
		return s.jobStore.RollbackTransaction(ctx, err)
	}

	return s.jobStore.CommitTransaction(ctx)
}

func (s *JobService) Delete(ctx context.Context, j *models.Job) error {
	return s.jobStore.Delete(ctx, j)
}

func (s *JobService) handleCreateError(ctx context.Context, j *models.Job, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.JobEventLifecycleCreated
	if err != nil {
		eventName = telemetry.JobEventLifecycleCreationFailed
	}
	event := telemetry.NewJobEvent(eventName, j, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
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
