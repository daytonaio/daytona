// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"fmt"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/ports"
	"github.com/docker/docker/pkg/stringid"
)

type IBuilderFactory interface {
	Create(build Build, projectDir string) (IBuilder, error)
	CheckExistingBuild(build Build) (*Build, error)
}

type BuilderFactory struct {
	containerRegistry   *containerregistry.ContainerRegistry
	buildImageNamespace string
	buildStore          Store
	loggerFactory       logs.LoggerFactory
	image               string
	defaultProjectImage string
	defaultProjectUser  string
}

type BuilderFactoryConfig struct {
	Image               string
	ContainerRegistry   *containerregistry.ContainerRegistry
	BuildStore          Store
	BuildImageNamespace string // Namespace to be used when tagging and pushing the build image
	LoggerFactory       logs.LoggerFactory
	DefaultProjectImage string
	DefaultProjectUser  string
}

func NewBuilderFactory(config BuilderFactoryConfig) IBuilderFactory {
	return &BuilderFactory{
		image:               config.Image,
		containerRegistry:   config.ContainerRegistry,
		buildImageNamespace: config.BuildImageNamespace,
		buildStore:          config.BuildStore,
		loggerFactory:       config.LoggerFactory,
		defaultProjectImage: config.DefaultProjectImage,
		defaultProjectUser:  config.DefaultProjectUser,
	}
}

func (f *BuilderFactory) Create(build Build, projectDir string) (IBuilder, error) {
	// TODO: Implement factory logic after adding prebuilds and other builder types
	return f.newDevcontainerBuilder(projectDir)
}

func (f *BuilderFactory) CheckExistingBuild(b Build) (*Build, error) {
	if b.Repository == nil || b.Repository.Branch == nil {
		return nil, fmt.Errorf("repository and branch must be set")
	}

	build, err := f.buildStore.Find(&Filter{
		Branch:        b.Repository.Branch,
		RepositoryUrl: &b.Repository.Url,
		BuildConfig:   b.BuildConfig,
		EnvVars:       &b.EnvVars,
	})
	if err != nil {
		return nil, err
	}

	return build, nil
}

func (f *BuilderFactory) newDevcontainerBuilder(projectDir string) (*DevcontainerBuilder, error) {
	builderDockerPort, err := ports.GetAvailableEphemeralPort()
	if err != nil {
		return nil, err
	}

	id := stringid.GenerateRandomID()
	id = stringid.TruncateID(id)
	id = fmt.Sprintf("%s-%s", "devcontainer-builder", id)

	return &DevcontainerBuilder{
		Builder: &Builder{
			id:                  id,
			projectDir:          projectDir,
			image:               f.image,
			containerRegistry:   f.containerRegistry,
			buildImageNamespace: f.buildImageNamespace,
			buildStore:          f.buildStore,
			loggerFactory:       f.loggerFactory,
			defaultProjectImage: f.defaultProjectImage,
			defaultProjectUser:  f.defaultProjectUser,
		},
		builderDockerPort: builderDockerPort,
	}, nil
}
