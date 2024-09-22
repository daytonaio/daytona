// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builds

import (
	"errors"
	"io"
	"time"

	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/server/builds/dto"
	"github.com/daytonaio/daytona/pkg/workspace/project/containerconfig"
	"github.com/docker/docker/pkg/stringid"
)

type IBuildService interface {
	Create(dto.BuildCreationData) (string, error)
	Find(filter *build.Filter) (*build.Build, error)
	List(filter *build.Filter) ([]*build.Build, error)
	MarkForDeletion(filter *build.Filter, force bool) []error
	Delete(id string) error
	AwaitEmptyList(time.Duration) error
	GetBuildLogReader(buildId string) (io.Reader, error)
}

type BuildServiceConfig struct {
	BuildStore    build.Store
	LoggerFactory logs.LoggerFactory
}

type BuildService struct {
	buildStore    build.Store
	loggerFactory logs.LoggerFactory
}

func NewBuildService(config BuildServiceConfig) IBuildService {
	return &BuildService{
		buildStore:    config.BuildStore,
		loggerFactory: config.LoggerFactory,
	}
}

func (s *BuildService) Create(b dto.BuildCreationData) (string, error) {
	var newBuild build.Build

	id := stringid.GenerateRandomID()
	id = stringid.TruncateID(id)

	newBuild.Id = id
	newBuild.State = build.BuildStatePendingRun
	newBuild.ContainerConfig = containerconfig.ContainerConfig{
		Image: b.Image,
		User:  b.User,
	}
	newBuild.BuildConfig = b.BuildConfig
	newBuild.Repository = b.Repository
	newBuild.EnvVars = b.EnvVars
	newBuild.PrebuildId = b.PrebuildId

	err := s.buildStore.Save(&newBuild)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *BuildService) Find(filter *build.Filter) (*build.Build, error) {
	return s.buildStore.Find(filter)
}

func (s *BuildService) List(filter *build.Filter) ([]*build.Build, error) {
	return s.buildStore.List(filter)
}

func (s *BuildService) MarkForDeletion(filter *build.Filter, force bool) []error {
	var errors []error

	builds, err := s.List(filter)
	if err != nil {
		return []error{err}
	}

	for _, b := range builds {
		if force {
			b.State = build.BuildStatePendingForcedDelete
		} else {
			b.State = build.BuildStatePendingDelete
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
