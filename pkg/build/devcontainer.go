// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"errors"
	"os"

	"github.com/daytonaio/daytona/pkg/build/detect"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type BuildOutcome struct {
	Outcome               string `json:"outcome"`
	ContainerId           string `json:"containerId"`
	RemoteUser            string `json:"remoteUser"`
	RemoteWorkspaceFolder string `json:"remoteWorkspaceFolder"`
}

type DevcontainerBuilder struct {
	*Builder
	builderDockerPort uint16
}

func (b *DevcontainerBuilder) Build(build Build) (string, string, error) {
	builderType, err := detect.DetectWorkspaceBuilderType(build.BuildConfig, b.workspaceDir, nil)
	if err != nil {
		return "", "", err
	}

	if builderType != detect.BuilderTypeDevcontainer {
		return "", "", errors.New("failed to detect devcontainer config")
	}

	return b.buildDevcontainer(build)
}

func (b *DevcontainerBuilder) CleanUp() error {
	return os.RemoveAll(b.workspaceDir)
}

func (b *DevcontainerBuilder) Publish(build Build) error {
	buildLogger := b.loggerFactory.CreateBuildLogger(build.Id, logs.LogSourceBuilder)
	defer buildLogger.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	if build.Image == nil {
		return errors.New("build image is nil")
	}

	return dockerClient.PushImage(*build.Image, b.buildImageContainerRegistry, buildLogger)
}

func (b *DevcontainerBuilder) buildDevcontainer(build Build) (string, string, error) {
	buildLogger := b.loggerFactory.CreateBuildLogger(build.Id, logs.LogSourceBuilder)
	defer buildLogger.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return b.defaultWorkspaceImage, b.defaultWorkspaceUser, err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	err = dockerClient.PullImage(b.image, b.containerRegistry, buildLogger)
	if err != nil {
		return b.defaultWorkspaceImage, b.defaultWorkspaceUser, err
	}

	containerId, remoteUser, err := dockerClient.CreateFromDevcontainer(docker.CreateDevcontainerOptions{
		BuildConfig:              build.BuildConfig,
		WorkspaceFolderName:      build.Id,
		ContainerRegistry:        b.buildImageContainerRegistry,
		BuilderImage:             b.image,
		BuilderContainerRegistry: b.containerRegistry,
		Prebuild:                 true,
		IdLabels: map[string]string{
			"daytona.build.id": build.Id,
		},
		WorkspaceDir: b.workspaceDir,
		LogWriter:    buildLogger,
		EnvVars:      build.EnvVars,
	})
	if err != nil {
		return b.defaultWorkspaceImage, b.defaultWorkspaceUser, err
	}

	defer dockerClient.RemoveContainer(containerId) // nolint: errcheck

	imageName, err := b.GetImageName(build)
	if err != nil {
		return b.defaultWorkspaceImage, b.defaultWorkspaceUser, err
	}

	_, err = cli.ContainerCommit(context.Background(), containerId, container.CommitOptions{
		Reference: imageName,
	})
	if err != nil {
		return b.defaultWorkspaceImage, b.defaultWorkspaceUser, err
	}

	return imageName, string(remoteUser), err
}
