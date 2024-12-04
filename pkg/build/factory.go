// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"errors"
	"fmt"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/ports"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/docker/docker/pkg/stringid"
)

type IBuilderFactory interface {
	Create(build models.Build, workspaceDir string) (IBuilder, error)
	CheckExistingBuild(build models.Build) (*models.Build, error)
}

type BuilderFactory struct {
	containerRegistry           *models.ContainerRegistry
	buildImageContainerRegistry *models.ContainerRegistry
	buildImageNamespace         string
	buildService                services.IBuildService
	loggerFactory               logs.LoggerFactory
	image                       string
	defaultWorkspaceImage       string
	defaultWorkspaceUser        string
}

type BuilderFactoryConfig struct {
	Image                       string
	ContainerRegistry           *models.ContainerRegistry
	BuildImageContainerRegistry *models.ContainerRegistry
	BuildService                services.IBuildService
	BuildImageNamespace         string // Namespace to be used when tagging and pushing the build image
	LoggerFactory               logs.LoggerFactory
	DefaultWorkspaceImage       string
	DefaultWorkspaceUser        string
}

func NewBuilderFactory(config BuilderFactoryConfig) IBuilderFactory {
	return &BuilderFactory{
		image:                       config.Image,
		containerRegistry:           config.ContainerRegistry,
		buildImageNamespace:         config.BuildImageNamespace,
		buildImageContainerRegistry: config.BuildImageContainerRegistry,
		buildService:                config.BuildService,
		loggerFactory:               config.LoggerFactory,
		defaultWorkspaceImage:       config.DefaultWorkspaceImage,
		defaultWorkspaceUser:        config.DefaultWorkspaceUser,
	}
}

func (f *BuilderFactory) Create(build models.Build, workspaceDir string) (IBuilder, error) {
	// TODO: Implement factory logic after adding prebuilds and other builder types
	return f.newDevcontainerBuilder(workspaceDir)
}

func (f *BuilderFactory) CheckExistingBuild(b models.Build) (*models.Build, error) {
	if b.Repository == nil {
		return nil, errors.New("repository must be set")
	}

	build, err := f.buildService.Find(&services.BuildFilter{
		StoreFilter: stores.BuildFilter{
			Branch:        &b.Repository.Branch,
			RepositoryUrl: &b.Repository.Url,
			BuildConfig:   b.BuildConfig,
			EnvVars:       &b.EnvVars,
		},
	}, services.BuildRetrievalParams{})
	if err != nil {
		return nil, err
	}

	return &build.Build, nil
}

func (f *BuilderFactory) newDevcontainerBuilder(workspaceDir string) (*DevcontainerBuilder, error) {
	builderDockerPort, err := ports.GetAvailableEphemeralPort()
	if err != nil {
		return nil, err
	}

	id := stringid.GenerateRandomID()
	id = stringid.TruncateID(id)
	id = fmt.Sprintf("%s-%s", "devcontainer-builder", id)

	return &DevcontainerBuilder{
		Builder: &Builder{
			id:                          id,
			workspaceDir:                workspaceDir,
			image:                       f.image,
			containerRegistry:           f.containerRegistry,
			buildImageContainerRegistry: f.buildImageContainerRegistry,
			buildImageNamespace:         f.buildImageNamespace,
			buildService:                f.buildService,
			loggerFactory:               f.loggerFactory,
			defaultWorkspaceImage:       f.defaultWorkspaceImage,
			defaultWorkspaceUser:        f.defaultWorkspaceUser,
		},
		builderDockerPort: builderDockerPort,
	}, nil
}
