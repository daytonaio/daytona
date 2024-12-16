// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builds

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/docker/docker/pkg/stringid"
)

type BuildServiceConfig struct {
	BuildStore            stores.BuildStore
	FindWorkspaceTemplate func(ctx context.Context, name string) (*models.WorkspaceTemplate, error)
	GetRepositoryContext  func(ctx context.Context, url, branch string) (*gitprovider.GitRepository, error)
	CreateJob             func(ctx context.Context, workspaceId string, action models.JobAction) error
	LoggerFactory         logs.ILoggerFactory
}

type BuildService struct {
	buildStore            stores.BuildStore
	findWorkspaceTemplate func(ctx context.Context, name string) (*models.WorkspaceTemplate, error)
	getRepositoryContext  func(ctx context.Context, url, branch string) (*gitprovider.GitRepository, error)
	createJob             func(ctx context.Context, workspaceId string, action models.JobAction) error
	loggerFactory         logs.ILoggerFactory
}

func NewBuildService(config BuildServiceConfig) services.IBuildService {
	return &BuildService{
		buildStore:            config.BuildStore,
		findWorkspaceTemplate: config.FindWorkspaceTemplate,
		getRepositoryContext:  config.GetRepositoryContext,
		loggerFactory:         config.LoggerFactory,
		createJob:             config.CreateJob,
	}
}

func (s *BuildService) Create(ctx context.Context, b services.CreateBuildDTO) (string, error) {
	id := stringid.GenerateRandomID()
	id = stringid.TruncateID(id)

	workspaceTemplate, err := s.findWorkspaceTemplate(ctx, b.WorkspaceTemplateName)
	if err != nil {
		return "", err
	}

	repo, err := s.getRepositoryContext(ctx, workspaceTemplate.RepositoryUrl, b.Branch)
	if err != nil {
		return "", err
	}

	newBuild := models.Build{
		Id: id,
		ContainerConfig: models.ContainerConfig{
			Image: workspaceTemplate.Image,
			User:  workspaceTemplate.User,
		},
		BuildConfig: workspaceTemplate.BuildConfig,
		Repository:  repo,
		EnvVars:     b.EnvVars,
	}

	if b.PrebuildId != nil {
		newBuild.PrebuildId = *b.PrebuildId
	}

	err = s.buildStore.Save(ctx, &newBuild)
	if err != nil {
		return "", err
	}

	err = s.createJob(ctx, id, models.JobActionRun)
	if err != nil {
		return "", err
	}

	return id, nil
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
		return []error{err}
	}

	for _, b := range builds {
		if force {
			err = s.createJob(ctx, b.Id, models.JobActionForceDelete)
			if err != nil {
				errors = append(errors, err)
			}
		} else {
			err = s.createJob(ctx, b.Id, models.JobActionDelete)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

func (s *BuildService) AwaitEmptyList(ctx context.Context, waitTime time.Duration) error {
	timeout := time.NewTimer(waitTime)
	defer timeout.Stop()

	for {
		select {
		case <-timeout.C:
			return errors.New("awaiting empty build list timed out")
		default:
			builds, err := s.List(ctx, &services.BuildFilter{
				ShowDeleted: true,
			})
			if err != nil {
				return err
			}

			if len(builds) == 0 {
				return nil
			}

			time.Sleep(time.Second)
		}
	}
}

func (s *BuildService) GetBuildLogReader(ctx context.Context, buildId string) (io.Reader, error) {
	return s.loggerFactory.CreateBuildLogReader(buildId)
}

func (s *BuildService) GetBuildLogWriter(ctx context.Context, buildId string) (io.WriteCloser, error) {
	targetLogsDir, err := build.GetBuildLogsDir()
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(targetLogsDir, 0755)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(targetLogsDir, buildId, "log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}
