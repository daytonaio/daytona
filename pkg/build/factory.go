// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/daytonaio/daytona/pkg/build/detect"
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/ports"
	"github.com/daytonaio/daytona/pkg/server/containerregistries"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/daytonaio/daytona/pkg/workspace/project/build"
	"github.com/docker/docker/pkg/stringid"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type IBuilderFactory interface {
	Create(p project.Project, gpc *gitprovider.GitProviderConfig) (IBuilder, error)
	CheckExistingBuild(p project.Project) (*BuildResult, error)
}

type BuilderFactory struct {
	serverConfigFolder       string
	containerRegistryServer  string
	buildImageNamespace      string
	buildResultStore         Store
	basePath                 string
	loggerFactory            logs.LoggerFactory
	image                    string
	containerRegistryService containerregistries.IContainerRegistryService
	defaultProjectImage      string
	defaultProjectUser       string
	createGitService         func(projectDir string, logWriter io.Writer) git.IGitService
}

type BuilderFactoryConfig struct {
	BuilderConfig
	CreateGitService func(projectDir string, logWriter io.Writer) git.IGitService
}

func NewBuilderFactory(config BuilderFactoryConfig) IBuilderFactory {
	return &BuilderFactory{
		image:                    config.Image,
		serverConfigFolder:       config.ServerConfigFolder,
		containerRegistryServer:  config.ContainerRegistryServer,
		buildImageNamespace:      config.BuildImageNamespace,
		buildResultStore:         config.BuildResultStore,
		containerRegistryService: config.ContainerRegistryService,
		basePath:                 config.BasePath,
		loggerFactory:            config.LoggerFactory,
		defaultProjectImage:      config.DefaultProjectImage,
		defaultProjectUser:       config.DefaultProjectUser,
		createGitService:         config.CreateGitService,
	}
}

func (f *BuilderFactory) Create(p project.Project, gpc *gitprovider.GitProviderConfig) (IBuilder, error) {
	buildId := stringid.GenerateRandomID()
	buildId = stringid.TruncateID(buildId)

	hash, err := p.GetConfigHash()
	if err != nil {
		return nil, err
	}
	projectDir := filepath.Join(f.basePath, hash, "project")

	err = os.RemoveAll(projectDir)
	if err != nil {
		return nil, err
	}

	var projectLogger logs.Logger

	if f.loggerFactory != nil {
		projectLogger = f.loggerFactory.CreateProjectLogger(p.WorkspaceId, p.Name, logs.LogSourceBuilder)
		defer projectLogger.Close()
	}

	gitservice := f.createGitService(projectDir, projectLogger)

	var auth *http.BasicAuth
	if gpc != nil {
		auth = &http.BasicAuth{
			Username: gpc.Username,
			Password: gpc.Token,
		}
	}

	err = gitservice.CloneRepository(&p, auth)
	if err != nil {
		return nil, err
	}

	if p.Build == nil || *p.Build != (build.ProjectBuildConfig{}) {
		if p.Build != nil && p.Build.Devcontainer != nil {
			return f.newDevcontainerBuilder(buildId, p, gpc, hash, projectDir)
		}

		return nil, nil
	}

	// Autodetect
	builderType, err := detect.DetectProjectBuilderType(&p, projectDir, nil)
	if err != nil {
		return nil, err
	}

	switch builderType {
	case detect.BuilderTypeDevcontainer:
		return f.newDevcontainerBuilder(buildId, p, gpc, hash, projectDir)
	case detect.BuilderTypeImage:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown builder type: %s", builderType)
	}
}

func (f *BuilderFactory) CheckExistingBuild(p project.Project) (*BuildResult, error) {
	hash, err := p.GetConfigHash()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(f.serverConfigFolder, "builds", hash, "build.json")

	_, err = os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result BuildResult
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}

	// If the builder registry changed, we need to rebuild and push again
	if !strings.HasPrefix(result.ImageName, fmt.Sprintf("%s%s", f.containerRegistryServer, f.buildImageNamespace)) {
		return nil, nil
	}

	return &result, nil
}

func (f *BuilderFactory) newDevcontainerBuilder(buildId string, p project.Project, gpc *gitprovider.GitProviderConfig, hash, projectDir string) (*DevcontainerBuilder, error) {
	builderDockerPort, err := ports.GetAvailableEphemeralPort()
	if err != nil {
		return nil, err
	}

	return &DevcontainerBuilder{
		Builder: &Builder{
			id:                       buildId,
			project:                  p,
			gitProviderConfig:        gpc,
			hash:                     hash,
			projectVolumePath:        projectDir,
			image:                    f.image,
			containerRegistryService: f.containerRegistryService,
			serverConfigFolder:       f.serverConfigFolder,
			containerRegistryServer:  f.containerRegistryServer,
			buildImageNamespace:      f.buildImageNamespace,
			buildResultStore:         f.buildResultStore,
			basePath:                 f.basePath,
			loggerFactory:            f.loggerFactory,
			defaultProjectImage:      f.defaultProjectImage,
			defaultProjectUser:       f.defaultProjectUser,
		},
		builderDockerPort: builderDockerPort,
	}, nil
}
