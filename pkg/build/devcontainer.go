// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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
	err := b.startContainer(build)
	if err != nil {
		return "", "", err
	}

	return b.buildDevcontainer(build)
}

func (b *DevcontainerBuilder) CleanUp() error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	err = cli.ContainerRemove(ctx, b.id, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		return err
	}

	err = os.RemoveAll(b.projectDir)
	if err != nil {
		return err
	}

	return nil
}

func (b *DevcontainerBuilder) Publish(build Build) error {
	buildLogger := b.loggerFactory.CreateBuildLogger(build.Project.Name, build.Id, logs.LogSourceBuilder)
	defer buildLogger.Close()

	cliBuilder, err := b.getBuilderDockerClient()
	if err != nil {
		return err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cliBuilder,
	})

	cr, err := b.containerRegistryService.Find(b.containerRegistryServer)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return err
	}

	return dockerClient.PushImage(build.Image, cr, buildLogger)
}

func (b *DevcontainerBuilder) buildDevcontainer(build Build) (string, string, error) {
	buildLogger := b.loggerFactory.CreateBuildLogger(build.Project.Name, build.Id, logs.LogSourceBuilder)
	defer buildLogger.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return b.defaultProjectImage, b.defaultProjectUser, err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	cmd := []string{"devcontainer", "up", "--prebuild", "--workspace-folder", "/project"}
	if build.Project.BuildConfig.Devcontainer.FilePath != "" {
		cmd = append(cmd, "--config", filepath.Join("/project", build.Project.BuildConfig.Devcontainer.FilePath))
	}

	execConfig := container.ExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		Tty:          true,
	}

	r, w := io.Pipe()
	var buildOutcome BuildOutcome

	go func(buildOutcome *BuildOutcome) {
		var lastLine string
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			lastLine = scanner.Text()
			buildLogger.Write([]byte(lastLine + "\n"))

			if strings.Contains(lastLine, `{"outcome"`) {
				start := strings.Index(lastLine, "{")
				end := strings.LastIndex(lastLine, "}")
				if start != -1 && end != -1 && start < end {
					lastLine = lastLine[start : end+1]
				}

				err = json.Unmarshal([]byte(lastLine), &buildOutcome)
				if err != nil {
					buildLogger.Write([]byte(err.Error() + "\n"))
				}
			}
		}
		if err := scanner.Err(); err != nil {
			buildLogger.Write([]byte(err.Error() + "\n"))
		}
	}(&buildOutcome)

	result, err := dockerClient.ExecSync(b.id, execConfig, w)
	if err != nil {
		return b.defaultProjectImage, b.defaultProjectUser, err
	}

	if buildOutcome.Outcome != "success" {
		return b.defaultProjectImage, b.defaultProjectUser, errors.New("devcontainer build failed")
	}

	if result.ExitCode != 0 {
		return b.defaultProjectImage, b.defaultProjectUser, errors.New(result.StdErr)
	}

	builderCli, err := b.getBuilderDockerClient()
	if err != nil {
		return b.defaultProjectImage, b.defaultProjectUser, err
	}

	tag := build.Project.Repository.Sha
	namespace := b.buildImageNamespace
	imageName := fmt.Sprintf("%s%s/p-%s:%s", b.containerRegistryServer, namespace, b.id, tag)

	_, err = builderCli.ContainerCommit(context.Background(), buildOutcome.ContainerId, container.CommitOptions{
		Reference: imageName,
	})
	if err != nil {
		return b.defaultProjectImage, b.defaultProjectUser, err
	}

	return imageName, buildOutcome.RemoteUser, nil
}

func (b *DevcontainerBuilder) startContainer(build Build) error {
	ctx := context.Background()

	buildLogger := b.loggerFactory.CreateBuildLogger(build.Project.Name, build.Id, logs.LogSourceBuilder)
	defer buildLogger.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	cr, err := b.containerRegistryService.FindByImageName(b.image)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return err
	}

	err = dockerClient.PullImage(b.image, cr, buildLogger)
	if err != nil {
		return err
	}

	serverHost, err := containerregistry.GetServerHostname(b.containerRegistryServer)
	if err != nil {
		return err
	}

	_, err = cli.ContainerCreate(ctx, &container.Config{
		Image:      b.image,
		Entrypoint: []string{"sudo", "dockerd", "-H", fmt.Sprintf("tcp://0.0.0.0:%d", b.builderDockerPort), "-H", "unix:///var/run/docker.sock", "--insecure-registry", serverHost},
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", b.builderDockerPort)): {},
		},
	}, &container.HostConfig{
		Privileged: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: b.projectDir,
				Target: "/project",
			},
		},
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", b.builderDockerPort)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprint(b.builderDockerPort),
				},
			},
		},
	}, nil, nil, b.id)
	if err != nil {
		return err
	}

	err = cli.ContainerStart(ctx, b.id, container.StartOptions{})
	if err != nil {
		return err
	}

	// Wait for docker to start
	builderCli, err := b.getBuilderDockerClient()
	if err != nil {
		return err
	}

	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		_, err = builderCli.Ping(ctx)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("timeout waiting for dockerd to start: %v", err)
	}

	return nil
}

func (b *DevcontainerBuilder) getBuilderDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.WithHost(fmt.Sprintf("tcp://127.0.0.1:%d", b.builderDockerPort)), client.WithAPIVersionNegotiation())
}
