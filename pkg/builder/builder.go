// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/git"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logger"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/pkg/stringid"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type Catalog struct {
	Repositories []string `json:"repositories"`
}

type BuildResult struct {
	User               string
	ImageName          string
	ProjectVolumePath  string
	PostCreateCommands []string
	PostStartCommands  []string
}

type BuilderPlugin interface {
	Build() (*BuildResult, error)
	Publish() error
	CleanUp() error
}

type BuilderConfig struct {
	DaytonaServerConfigFolder       string
	LocalContainerRegistryServer    string
	BasePath                        string
	LoggerFactory                   logger.LoggerFactory
	DefaultProjectImage             string
	DefaultProjectUser              string
	DefaultProjectPostStartCommands []string
}

type IBuilder interface {
	Prepare() error
	LoadBuildResults() (*BuildResult, error)
	SaveBuildResults(r BuildResult) error
	GetBuilderPlugin() BuilderPlugin
}

type Builder struct {
	id                string
	plugin            BuilderPlugin
	project           workspace.Project
	containerRegistry *containerregistry.ContainerRegistry
	gitProviderConfig *gitprovider.GitProviderConfig
	hash              string

	daytonaServerConfigFolder       string
	localContainerRegistryServer    string
	basePath                        string
	loggerFactory                   logger.LoggerFactory
	defaultProjectImage             string
	defaultProjectUser              string
	defaultProjectPostStartCommands []string
}

type IBuilderFactory interface {
	Create(p workspace.Project, cr *containerregistry.ContainerRegistry, gpc *gitprovider.GitProviderConfig) IBuilder
}

type BuilderFactory struct {
	daytonaServerConfigFolder       string
	localContainerRegistryServer    string
	basePath                        string
	loggerFactory                   logger.LoggerFactory
	defaultProjectImage             string
	defaultProjectUser              string
	defaultProjectPostStartCommands []string
}

func NewBuilderFactory(config BuilderConfig) IBuilderFactory {
	return &BuilderFactory{
		daytonaServerConfigFolder:       config.DaytonaServerConfigFolder,
		localContainerRegistryServer:    config.LocalContainerRegistryServer,
		basePath:                        config.BasePath,
		loggerFactory:                   config.LoggerFactory,
		defaultProjectImage:             config.DefaultProjectImage,
		defaultProjectUser:              config.DefaultProjectUser,
		defaultProjectPostStartCommands: config.DefaultProjectPostStartCommands,
	}
}

func (f *BuilderFactory) Create(p workspace.Project, cr *containerregistry.ContainerRegistry, gpc *gitprovider.GitProviderConfig) IBuilder {
	buildId := stringid.GenerateRandomID()
	buildId = stringid.TruncateID(buildId)

	builder := &Builder{
		id:                buildId,
		plugin:            nil,
		project:           p,
		containerRegistry: cr,
		gitProviderConfig: gpc,

		daytonaServerConfigFolder:       f.daytonaServerConfigFolder,
		localContainerRegistryServer:    f.localContainerRegistryServer,
		basePath:                        f.basePath,
		loggerFactory:                   f.loggerFactory,
		defaultProjectImage:             f.defaultProjectImage,
		defaultProjectUser:              f.defaultProjectUser,
		defaultProjectPostStartCommands: f.defaultProjectPostStartCommands,
	}

	return builder
}

func (b *Builder) GetBuilderPlugin() BuilderPlugin {
	return b.plugin
}

func (b *Builder) Prepare() error {
	hash, err := b.project.GetConfigHash()
	if err != nil {
		return err
	}
	b.hash = hash
	projectDir := filepath.Join(b.basePath, hash, "project")

	err = os.RemoveAll(projectDir)
	if err != nil {
		return err
	}

	projectLogger := b.loggerFactory.CreateProjectLogger(b.project.WorkspaceId, b.project.Name)
	defer projectLogger.Close()

	gitservice := git.Service{
		ProjectDir:        projectDir,
		GitConfigFileName: "",
		LogWriter:         projectLogger,
	}

	var auth *http.BasicAuth
	if b.gitProviderConfig != nil {
		auth = &http.BasicAuth{
			Username: b.gitProviderConfig.Username,
			Password: b.gitProviderConfig.Token,
		}
	}

	err = gitservice.CloneRepository(&b.project, auth)
	if err != nil {
		return err
	}

	buildConfig := b.project.Build

	if buildConfig != nil && *buildConfig == (workspace.ProjectBuild{}) {
		//	detect is devcontainer
		devcontainerPath := ".devcontainer/devcontainer.json"
		isDevcontainer, err := fileExists(filepath.Join(projectDir, devcontainerPath))
		if err != nil {
			return err
		}
		if !isDevcontainer {
			devcontainerPath = ".devcontainer.json"
			isDevcontainer, err = fileExists(filepath.Join(projectDir, devcontainerPath))
			if err != nil {
				return err
			}
		}
		if isDevcontainer {
			buildConfig.Devcontainer = &workspace.ProjectBuildDevcontainer{
				DevContainerFilePath: devcontainerPath,
			}
			goto initPlugin
		}
		//	todo: detect dockerfile
		//	todo: detect nix

		//	no supported dev config standard found
		//	set default project image to ensure that project will run anyway
		b.project.Image = b.defaultProjectImage
		b.project.User = b.defaultProjectUser
		b.project.PostStartCommands = b.defaultProjectPostStartCommands
	}

initPlugin:

	if buildConfig.Devcontainer != nil {
		b.plugin = &DevcontainerBuilder{
			DevcontainerBuilderConfig: DevcontainerBuilderConfig{
				buildId:                      b.id,
				project:                      b.project,
				loggerFactory:                b.loggerFactory,
				localContainerRegistryServer: b.localContainerRegistryServer,
				projectVolumePath:            filepath.Join(b.basePath, b.hash, "project"),
			},
		}
	}

	return nil
}

func (b *Builder) LoadBuildResults() (*BuildResult, error) {
	hash, err := b.project.GetConfigHash()
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(b.daytonaServerConfigFolder, "builds", hash, "build.json")

	_, err = os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		if pathErr, ok := err.(*os.PathError); ok && pathErr.Err.Error() == "not a directory" {
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

	return &result, nil
}

func (b *Builder) SaveBuildResults(r BuildResult) error {
	err := os.MkdirAll(filepath.Join(b.daytonaServerConfigFolder, "builds", b.hash), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(b.daytonaServerConfigFolder, "builds", b.hash, "build.json"))
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

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		// There was an error checking for the file
		return false, err
	}
	return true, nil
}
