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
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"
)

type Catalog struct {
	Repositories []string `json:"repositories"`
}

type BuildResult struct {
	User              string
	ImageName         string
	ProjectVolumePath string
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
	BuilderConfig
	id                string
	plugin            BuilderPlugin
	project           workspace.Project
	containerRegistry *containerregistry.ContainerRegistry
	gitProviderConfig *gitprovider.GitProviderConfig
	hash              string
}

type IBuilderFactory interface {
	Create(p workspace.Project, cr *containerregistry.ContainerRegistry, gpc *gitprovider.GitProviderConfig) IBuilder
}

type BuilderFactory struct {
	BuilderConfig
}

func (f *BuilderFactory) Create(p workspace.Project, cr *containerregistry.ContainerRegistry, gpc *gitprovider.GitProviderConfig) IBuilder {
	uuid := uuid.New()
	buildId := uuid.String()[:8]

	builder := &Builder{
		BuilderConfig: BuilderConfig{
			DaytonaServerConfigFolder:       f.DaytonaServerConfigFolder,
			LocalContainerRegistryServer:    f.LocalContainerRegistryServer,
			BasePath:                        f.BasePath,
			LoggerFactory:                   f.LoggerFactory,
			DefaultProjectImage:             f.DefaultProjectImage,
			DefaultProjectUser:              f.DefaultProjectUser,
			DefaultProjectPostStartCommands: f.DefaultProjectPostStartCommands,
		},
		id:                buildId,
		plugin:            nil,
		project:           p,
		containerRegistry: cr,
		gitProviderConfig: gpc,
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
	projectDir := filepath.Join(b.BasePath, hash, "project")

	err = os.RemoveAll(projectDir)
	if err != nil {
		return err
	}

	gitservice := git.Service{
		ProjectDir:        projectDir,
		GitConfigFileName: "",
		//	todo: write to project log
		LogWriter: os.Stdout,
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
		b.project.Image = b.DefaultProjectImage
		b.project.User = b.DefaultProjectUser
		b.project.PostStartCommands = b.DefaultProjectPostStartCommands
	}

initPlugin:

	if buildConfig.Devcontainer != nil {
		b.plugin = &DevcontainerBuilder{
			DevcontainerBuilderConfig: DevcontainerBuilderConfig{
				buildId:                      b.id,
				project:                      b.project,
				loggerFactory:                b.LoggerFactory,
				localContainerRegistryServer: b.LocalContainerRegistryServer,
				projectVolumePath:            filepath.Join(b.BasePath, b.hash, "project"),
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

	filePath := filepath.Join(b.DaytonaServerConfigFolder, "builds", hash, "build.json")

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
	err := os.MkdirAll(filepath.Join(b.DaytonaServerConfigFolder, "builds", b.hash), 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(b.DaytonaServerConfigFolder, "builds", b.hash, "build.json"))
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
