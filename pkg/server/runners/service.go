// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runners

import (
	"context"
	"io"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

type RunnerServiceConfig struct {
	RunnerStore         stores.RunnerStore
	RunnerMetadataStore stores.RunnerMetadataStore
	LoggerFactory       logs.ILoggerFactory

	CreateJob          func(ctx context.Context, runnerId string, action models.JobAction, metadata string) error
	ListJobsForRunner  func(ctx context.Context, runnerId string) ([]*models.Job, error)
	UpdateJobState     func(ctx context.Context, jobId string, updateJobStateDto services.UpdateJobStateDTO) error
	GenerateApiKey     func(ctx context.Context, name string) (string, error)
	RevokeApiKey       func(ctx context.Context, name string) error
	UnsetDefaultTarget func(ctx context.Context, runnerId string) error

	TrackTelemetryEvent func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error
}

func NewRunnerService(config RunnerServiceConfig) services.IRunnerService {
	return &RunnerService{
		runnerStore:         config.RunnerStore,
		runnerMetadataStore: config.RunnerMetadataStore,
		loggerFactory:       config.LoggerFactory,

		createJob:          config.CreateJob,
		listJobsForRunner:  config.ListJobsForRunner,
		updateJobState:     config.UpdateJobState,
		generateApiKey:     config.GenerateApiKey,
		revokeApiKey:       config.RevokeApiKey,
		unsetDefaultTarget: config.UnsetDefaultTarget,

		trackTelemetryEvent: config.TrackTelemetryEvent,
	}
}

type RunnerService struct {
	runnerStore         stores.RunnerStore
	runnerMetadataStore stores.RunnerMetadataStore
	loggerFactory       logs.ILoggerFactory

	createJob          func(ctx context.Context, runnerId string, action models.JobAction, metadata string) error
	listJobsForRunner  func(ctx context.Context, runnerId string) ([]*models.Job, error)
	updateJobState     func(ctx context.Context, jobId string, updateJobStateDto services.UpdateJobStateDTO) error
	generateApiKey     func(ctx context.Context, name string) (string, error)
	revokeApiKey       func(ctx context.Context, name string) error
	unsetDefaultTarget func(ctx context.Context, runnerId string) error

	trackTelemetryEvent func(event telemetry.ServerEvent, clientId string, props map[string]interface{}) error
}

func (s *RunnerService) GetRunner(ctx context.Context, runnerId string) (*services.RunnerDTO, error) {
	runner, err := s.runnerStore.Find(ctx, runnerId)
	if err != nil {
		return nil, stores.ErrRunnerNotFound
	}

	return &services.RunnerDTO{
		Runner: *runner,
		State:  runner.GetState(),
	}, nil
}

func (s *RunnerService) ListRunners(ctx context.Context) ([]*services.RunnerDTO, error) {
	runners, err := s.runnerStore.List(ctx)
	if err != nil {
		return nil, err
	}

	return util.ArrayMap(runners, func(runner *models.Runner) *services.RunnerDTO {
		return &services.RunnerDTO{
			Runner: *runner,
			State:  runner.GetState(),
		}
	}), nil
}

func (s *RunnerService) SetRunnerMetadata(ctx context.Context, runnerId string, metadata *models.RunnerMetadata) error {
	m, err := s.runnerMetadataStore.Find(ctx, runnerId)
	if err != nil {
		return stores.ErrRunnerMetadataNotFound
	}

	m.Uptime = metadata.Uptime
	m.RunningJobs = metadata.RunningJobs
	m.Providers = metadata.Providers
	m.UpdatedAt = metadata.UpdatedAt
	return s.runnerMetadataStore.Save(ctx, m)
}

func (s *RunnerService) GetRunnerLogReader(ctx context.Context, runnerId string) (io.Reader, error) {
	return s.loggerFactory.CreateLogReader(runnerId)
}

func (s *RunnerService) GetRunnerLogWriter(ctx context.Context, runnerId string) (io.WriteCloser, error) {
	return s.loggerFactory.CreateLogWriter(runnerId)
}
