// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/workspace/project"
)

type BuilderConfig struct {
	Image                    string
	ContainerRegistryService containerregistries.IContainerRegistryService
	ServerConfigFolder       string
	ContainerRegistryServer  string
	BuildResultStore         Store
	// Namespace to be used when tagging and pushing the build image
	BuildImageNamespace string
	BasePath            string
	LoggerFactory       logs.LoggerFactory
	DefaultProjectImage string
	DefaultProjectUser  string
}

type IBuilder interface {
	Build() (*BuildResult, error)
	CleanUp() error
	Publish() error
	SaveBuildResults(r BuildResult) error
}

type Builder struct {
	id                string
	project           project.Project
	gitProviderConfig *gitprovider.GitProviderConfig
	hash              string
	projectVolumePath string

	image                    string
	containerRegistryService containerregistries.IContainerRegistryService
	containerRegistryServer  string
	buildImageNamespace      string
	buildResultStore         Store
	serverConfigFolder       string
	basePath                 string
	loggerFactory            logs.LoggerFactory
	defaultProjectImage      string
	defaultProjectUser       string
}

func (b *Builder) SaveBuildResults(r BuildResult) error {
	return b.buildResultStore.Save(&r)
}
