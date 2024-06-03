// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logger"
	"github.com/daytonaio/daytona/pkg/workspace"
)

type BuildResult struct {
	User               string
	ImageName          string
	ProjectVolumePath  string
	PostCreateCommands []string
	PostStartCommands  []string
}

type BuilderConfig struct {
	ServerConfigFolder              string
	LocalContainerRegistryServer    string
	BasePath                        string
	LoggerFactory                   logger.LoggerFactory
	DefaultProjectImage             string
	DefaultProjectUser              string
	DefaultProjectPostStartCommands []string
}

type IBuilder interface {
	Build() (*BuildResult, error)
	CleanUp() error
	Publish() error
	SaveBuildResults(r BuildResult) error
}

type Builder struct {
	id                string
	project           workspace.Project
	gitProviderConfig *gitprovider.GitProviderConfig
	hash              string
	projectVolumePath string

	serverConfigFolder              string
	localContainerRegistryServer    string
	basePath                        string
	loggerFactory                   logger.LoggerFactory
	defaultProjectImage             string
	defaultProjectUser              string
	defaultProjectPostStartCommands []string
}

func (b *Builder) SaveBuildResults(r BuildResult) error {
	err := os.MkdirAll(filepath.Join(b.serverConfigFolder, "builds", b.hash), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(b.serverConfigFolder, "builds", b.hash, "build.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(r)
	if err != nil {
		return err
	}

	return nil
}
