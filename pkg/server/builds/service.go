// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builds

import (
	"context"
	"errors"
	"io"
	"time"

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
	LoggerFactory         logs.LoggerFactory
}

type BuildService struct {
	buildStore            stores.BuildStore
	findWorkspaceTemplate func(ctx context.Context, name string) (*models.WorkspaceTemplate, error)
	getRepositoryContext  func(ctx context.Context, url, branch string) (*gitprovider.GitRepository, error)
	loggerFactory         logs.LoggerFactory
}

func NewBuildService(config BuildServiceConfig) services.IBuildService {
	return &BuildService{
		buildStore:            config.BuildStore,
		findWorkspaceTemplate: config.FindWorkspaceTemplate,
		getRepositoryContext:  config.GetRepositoryContext,
		loggerFactory:         config.LoggerFactory,
	}
}

func (s *BuildService) Create(b services.CreateBuildDTO) (string, error) {
	id := stringid.GenerateRandomID()
	id = stringid.TruncateID(id)
	ctx := context.Background()

	workspaceTemplate, err := s.findWorkspaceTemplate(ctx, b.WorkspaceTemplateName)
	if err != nil {
		return "", err
	}

	repo, err := s.getRepositoryContext(ctx, workspaceTemplate.RepositoryUrl, b.Branch)
	if err != nil {
		return "", err
	}

	newBuild := models.Build{
		Id:    id,
		State: models.BuildStatePendingRun,
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

	err = s.buildStore.Save(&newBuild)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *BuildService) Find(filter *stores.BuildFilter) (*models.Build, error) {
	return s.buildStore.Find(filter)
}

func (s *BuildService) List(filter *stores.BuildFilter) ([]*models.Build, error) {
	return s.buildStore.List(filter)
}

func (s *BuildService) MarkForDeletion(filter *stores.BuildFilter, force bool) []error {
	var errors []error

	builds, err := s.List(filter)
	if err != nil {
		return []error{err}
	}

	for _, b := range builds {
		if force {
			b.State = models.BuildStatePendingForcedDelete
		} else {
			b.State = models.BuildStatePendingDelete
		}

		err = s.buildStore.Save(b)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (s *BuildService) Delete(id string) error {
	return s.buildStore.Delete(id)
}

func (s *BuildService) AwaitEmptyList(waitTime time.Duration) error {
	timeout := time.NewTimer(waitTime)
	defer timeout.Stop()

	for {
		select {
		case <-timeout.C:
			return errors.New("awaiting empty build list timed out")
		default:
			builds, err := s.List(nil)
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

func (s *BuildService) GetBuildLogReader(buildId string) (io.Reader, error) {
	return s.loggerFactory.CreateBuildLogReader(buildId)
}
