// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/ports"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/workspace/project"
)

type IBuilderFactory interface {
	Create(build Build) (IBuilder, error)
	CheckExistingBuild(p project.Project) (*Build, error)
}

type BuilderFactory struct {
	containerRegistryServer  string
	buildImageNamespace      string
	buildStore               Store
	basePath                 string
	loggerFactory            logs.LoggerFactory
	image                    string
	containerRegistryService containerregistries.IContainerRegistryService
	defaultProjectImage      string
	defaultProjectUser       string
}

type BuilderFactoryConfig struct {
	Image                    string
	ContainerRegistryService containerregistries.IContainerRegistryService
	ContainerRegistryServer  string
	BuildStore               Store
	BuildImageNamespace      string // Namespace to be used when tagging and pushing the build image
	LoggerFactory            logs.LoggerFactory
	DefaultProjectImage      string
	DefaultProjectUser       string
	BasePath                 string
}

func NewBuilderFactory(config BuilderFactoryConfig) IBuilderFactory {
	return &BuilderFactory{
		image:                    config.Image,
		containerRegistryServer:  config.ContainerRegistryServer,
		buildImageNamespace:      config.BuildImageNamespace,
		buildStore:               config.BuildStore,
		containerRegistryService: config.ContainerRegistryService,
		loggerFactory:            config.LoggerFactory,
		defaultProjectImage:      config.DefaultProjectImage,
		defaultProjectUser:       config.DefaultProjectUser,
		basePath:                 config.BasePath,
	}
}

func (f *BuilderFactory) Create(build Build) (IBuilder, error) {
	// TODO: Implement factory logic after adding prebuilds and other builder types
	return f.newDevcontainerBuilder(build)
}

func (f *BuilderFactory) CheckExistingBuild(p project.Project) (*Build, error) {
	hash, err := p.GetConfigHash()
	if err != nil {
		return nil, err
	}

	build, err := f.buildStore.Find(hash)
	if err != nil {
		return nil, err
	}

	return build, nil
}

func (f *BuilderFactory) newDevcontainerBuilder(build Build) (*DevcontainerBuilder, error) {
	builderDockerPort, err := ports.GetAvailableEphemeralPort()
	if err != nil {
		return nil, err
	}

	return &DevcontainerBuilder{
		Builder: &Builder{
			hash:                     build.Hash,
			projectDir:               filepath.Join(f.basePath, build.Hash, "project"),
			image:                    f.image,
			containerRegistryService: f.containerRegistryService,
			containerRegistryServer:  f.containerRegistryServer,
			buildImageNamespace:      f.buildImageNamespace,
			buildStore:               f.buildStore,
			loggerFactory:            f.loggerFactory,
			defaultProjectImage:      f.defaultProjectImage,
			defaultProjectUser:       f.defaultProjectUser,
		},
		builderDockerPort: builderDockerPort,
	}, nil
}
