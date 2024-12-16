// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/ports"
	"github.com/docker/docker/pkg/stringid"
)

type IBuilderFactory interface {
	Create(build models.Build, workspaceDir string) (IBuilder, error)
}

type BuilderFactory struct {
	containerRegistries         common.ContainerRegistries
	buildImageContainerRegistry *models.ContainerRegistry
	buildImageNamespace         string
	loggerFactory               logs.ILoggerFactory
	image                       string
	defaultWorkspaceImage       string
	defaultWorkspaceUser        string
}

type BuilderFactoryConfig struct {
	Image                       string
	ContainerRegistries         common.ContainerRegistries
	BuildImageContainerRegistry *models.ContainerRegistry
	BuildImageNamespace         string // Namespace to be used when tagging and pushing the build image
	LoggerFactory               logs.ILoggerFactory
	DefaultWorkspaceImage       string
	DefaultWorkspaceUser        string
}

func NewBuilderFactory(config BuilderFactoryConfig) IBuilderFactory {
	return &BuilderFactory{
		image:                       config.Image,
		containerRegistries:         config.ContainerRegistries,
		buildImageNamespace:         config.BuildImageNamespace,
		buildImageContainerRegistry: config.BuildImageContainerRegistry,
		loggerFactory:               config.LoggerFactory,
		defaultWorkspaceImage:       config.DefaultWorkspaceImage,
		defaultWorkspaceUser:        config.DefaultWorkspaceUser,
	}
}

func (f *BuilderFactory) Create(build models.Build, workspaceDir string) (IBuilder, error) {
	// TODO: Implement factory logic after adding prebuilds and other builder types
	return f.newDevcontainerBuilder(workspaceDir)
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
			containerRegistries:         f.containerRegistries,
			buildImageContainerRegistry: f.buildImageContainerRegistry,
			buildImageNamespace:         f.buildImageNamespace,
			loggerFactory:               f.loggerFactory,
			defaultWorkspaceImage:       f.defaultWorkspaceImage,
			defaultWorkspaceUser:        f.defaultWorkspaceUser,
		},
		builderDockerPort: builderDockerPort,
	}, nil
}
