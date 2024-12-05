// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"

	"github.com/daytonaio/daytona/pkg/jobs"
	"github.com/daytonaio/daytona/pkg/jobs/build"
	"github.com/daytonaio/daytona/pkg/jobs/target"
	"github.com/daytonaio/daytona/pkg/jobs/workspace"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runners"
	"github.com/daytonaio/daytona/pkg/scheduler"
	log "github.com/sirupsen/logrus"
)

type JobRunnerConfig struct {
	ListPendingJobs func(ctx context.Context) ([]*models.Job, error)
	UpdateJobState  func(ctx context.Context, job *models.Job, state models.JobState, err *error) error

	WorkspaceJobFactory workspace.IWorkspaceJobFactory
	TargetJobFactory    target.ITargetJobFactory
	BuildJobFactory     build.IBuildJobFactory
}

func NewJobRunner(config JobRunnerConfig) runners.IJobRunner {
	return &JobRunner{
		listPendingJobs: config.ListPendingJobs,
		updateJobState:  config.UpdateJobState,

		workspaceJobFactory: config.WorkspaceJobFactory,
		targetJobFactory:    config.TargetJobFactory,
		buildJobFactory:     config.BuildJobFactory,
	}
}

type JobRunner struct {
	listPendingJobs func(ctx context.Context) ([]*models.Job, error)
	updateJobState  func(ctx context.Context, job *models.Job, state models.JobState, err *error) error

	workspaceJobFactory workspace.IWorkspaceJobFactory
	targetJobFactory    target.ITargetJobFactory
	buildJobFactory     build.IBuildJobFactory
}

func (s *JobRunner) StartRunner(ctx context.Context) error {
	scheduler := scheduler.NewCronScheduler()

	err := scheduler.AddFunc(runners.DEFAULT_JOB_POLL_INTERVAL, func() {
		err := s.CheckAndRunJobs(ctx)
		if err != nil {
			log.Error(err)
		}
	})
	if err != nil {
		return err
	}

	scheduler.Start()
	return nil
}

func (s *JobRunner) CheckAndRunJobs(ctx context.Context) error {
	jobs, err := s.listPendingJobs(ctx)
	if err != nil {
		return err
	}

	// goroutines, sync group
	for _, job := range jobs {
		err = s.runJob(ctx, job)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *JobRunner) runJob(ctx context.Context, j *models.Job) error {
	var job jobs.IJob

	err := s.updateJobState(ctx, j, models.JobStateRunning, nil)
	if err != nil {
		return err
	}

	switch j.ResourceType {
	case models.ResourceTypeWorkspace:
		job = s.workspaceJobFactory.Create(*j)
	case models.ResourceTypeTarget:
		job = s.targetJobFactory.Create(*j)
	case models.ResourceTypeBuild:
		job = s.buildJobFactory.Create(*j)
	}

	err = job.Execute(ctx)
	if err != nil {
		return s.updateJobState(ctx, j, models.JobStateError, &err)
	}

	return s.updateJobState(ctx, j, models.JobStateSuccess, nil)
}
