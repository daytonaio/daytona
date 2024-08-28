// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"
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
	builderType, err := detect.DetectProjectBuilderType(build.BuildConfig, b.projectDir, nil)
	if err != nil {
		return "", "", err
	}

	if builderType != detect.BuilderTypeDevcontainer {
		return "", "", fmt.Errorf("failed to detect devcontainer config")
	}

	return b.buildDevcontainer(build)
}

func (b *DevcontainerBuilder) CleanUp() error {
	return os.RemoveAll(b.projectDir)
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

	return dockerClient.PushImage(build.Image, b.containerRegistry, buildLogger)
}

func (b *DevcontainerBuilder) buildDevcontainer(build Build) (string, string, error) {
	buildLogger := b.loggerFactory.CreateBuildLogger(build.Id, logs.LogSourceBuilder)
	defer buildLogger.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return b.defaultProjectImage, b.defaultProjectUser, err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	containerId, remoteUser, err := dockerClient.CreateFromDevcontainer(docker.CreateDevcontainerOptions{
		BuildConfig:       build.BuildConfig,
		ProjectName:       build.Id,
		ContainerRegistry: b.containerRegistry,
		Prebuild:          true,
		IdLabels: map[string]string{
			"daytona.build.id": build.Id,
		},
		ProjectDir: b.projectDir,
		LogWriter:  buildLogger,
	})
	if err != nil {
		return b.defaultProjectImage, b.defaultProjectUser, err
	}

	defer dockerClient.RemoveContainer(containerId) // nolint: errcheck

	imageName, err := b.GetImageName(build)
	if err != nil {
		return b.defaultProjectImage, b.defaultProjectUser, err
	}

	_, err = cli.ContainerCommit(context.Background(), containerId, container.CommitOptions{
		Reference: imageName,
	})
	if err != nil {
		return b.defaultProjectImage, b.defaultProjectUser, err
	}

	return imageName, string(remoteUser), err
}