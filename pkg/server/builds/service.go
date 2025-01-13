// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builds

import (
	"context"
	"io"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/daytonaio/daytona/pkg/telemetry"

	log "github.com/sirupsen/logrus"
)

type BuildServiceConfig struct {
	BuildStore            stores.BuildStore
	FindWorkspaceTemplate func(ctx context.Context, name string) (*models.WorkspaceTemplate, error)
	GetRepositoryContext  func(ctx context.Context, url, branch string) (*gitprovider.GitRepository, error)
	CreateJob             func(ctx context.Context, workspaceId string, action models.JobAction) error
	TrackTelemetryEvent   func(event telemetry.Event, clientId string) error
	LoggerFactory         logs.ILoggerFactory
}

type BuildService struct {
	buildStore            stores.BuildStore
	findWorkspaceTemplate func(ctx context.Context, name string) (*models.WorkspaceTemplate, error)
	getRepositoryContext  func(ctx context.Context, url, branch string) (*gitprovider.GitRepository, error)
	createJob             func(ctx context.Context, workspaceId string, action models.JobAction) error
	trackTelemetryEvent   func(event telemetry.Event, clientId string) error
	loggerFactory         logs.ILoggerFactory
}

func NewBuildService(config BuildServiceConfig) services.IBuildService {
	return &BuildService{
		buildStore:            config.BuildStore,
		findWorkspaceTemplate: config.FindWorkspaceTemplate,
		getRepositoryContext:  config.GetRepositoryContext,
		loggerFactory:         config.LoggerFactory,
		createJob:             config.CreateJob,
		trackTelemetryEvent:   config.TrackTelemetryEvent,
	}
}

func (s *BuildService) Find(ctx context.Context, filter *services.BuildFilter) (*services.BuildDTO, error) {
	var storeFilter *stores.BuildFilter

	if filter != nil {
		storeFilter = &filter.StoreFilter
	}

	build, err := s.buildStore.Find(ctx, storeFilter)
	if err != nil {
		return nil, err
	}

	state := build.GetState()

	if state.Name == models.ResourceStateNameDeleted && (filter == nil || !filter.ShowDeleted) {
		return nil, services.ErrBuildDeleted
	}

	return &services.BuildDTO{
		Build: *build,
		State: state,
	}, nil
}

func (s *BuildService) List(ctx context.Context, filter *services.BuildFilter) ([]*services.BuildDTO, error) {
	var storeFilter *stores.BuildFilter

	if filter != nil {
		storeFilter = &filter.StoreFilter
	}

	builds, err := s.buildStore.List(ctx, storeFilter)
	if err != nil {
		return nil, err
	}

	var result []*services.BuildDTO

	for _, b := range builds {
		state := b.GetState()

		if state.Name == models.ResourceStateNameDeleted && (filter == nil || !filter.ShowDeleted) {
			continue
		}

		result = append(result, &services.BuildDTO{
			Build: *b,
			State: state,
		})
	}

	return result, nil
}

func (s *BuildService) HandleSuccessfulRemoval(ctx context.Context, id string) error {
	return s.buildStore.Delete(ctx, id)
}

func (s *BuildService) Delete(ctx context.Context, filter *services.BuildFilter, force bool) []error {
	var errors []error

	builds, err := s.List(ctx, filter)
	if err != nil {
		return []error{s.handleDeleteError(ctx, nil, err, force)}
	}

	for _, b := range builds {
		if force {
			err = s.createJob(ctx, b.Id, models.JobActionForceDelete)
			err = s.handleDeleteError(ctx, &b.Build, err, force)
			if err != nil {
				errors = append(errors, err)
			}
		} else {
			err = s.createJob(ctx, b.Id, models.JobActionDelete)
			err = s.handleDeleteError(ctx, &b.Build, err, force)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

func (s *BuildService) handleDeleteError(ctx context.Context, b *models.Build, err error, force bool) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.BuildEventLifecycleDeleted
	if force {
		eventName = telemetry.BuildEventLifecycleForceDeleted
	}
	if err != nil {
		eventName = telemetry.BuildEventLifecycleDeletionFailed
		if force {
			eventName = telemetry.BuildEventLifecycleForceDeletionFailed
		}
	}

	event := telemetry.NewBuildEvent(eventName, b, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}

func (s *BuildService) GetBuildLogReader(ctx context.Context, buildId string) (io.Reader, error) {
	return s.loggerFactory.CreateLogReader(buildId)
}

func (s *BuildService) GetBuildLogWriter(ctx context.Context, buildId string) (io.WriteCloser, error) {
	return s.loggerFactory.CreateLogWriter(buildId)
}
