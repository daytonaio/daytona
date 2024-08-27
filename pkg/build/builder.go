// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
)

type IBuilder interface {
	Build(build Build) (string, string, error)
	CleanUp() error
	Publish(build Build) error
	SaveBuild(build Build) error
}

type Builder struct {
	id                       string
	hash                     string
	projectDir               string
	image                    string
	containerRegistryService containerregistries.IContainerRegistryService
	containerRegistryServer  string
	buildStore               Store
	buildImageNamespace      string
	loggerFactory            logs.LoggerFactory
	defaultProjectImage      string
	defaultProjectUser       string
}

func (b *Builder) SaveBuild(build Build) error {
	return b.buildStore.Save(&build)
}
