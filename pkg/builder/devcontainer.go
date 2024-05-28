// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package builder

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

	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/pkg/builder/devcontainer"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/docker/docker/api/types"
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
	buildImageName     string
	user               string
	builderDockerPort  uint16
	postCreateCommands []string
	postStartCommands  []string
}

func (b *DevcontainerBuilder) Build() (*BuildResult, error) {
	err := b.startContainer()
	if err != nil {
		return nil, err
	}

	err = b.buildDevcontainer()
	if err != nil {
		return nil, err
	}

	err = b.readConfiguration()
	if err != nil {
		return nil, err
	}

	return &BuildResult{
		User:               b.user,
		ImageName:          b.buildImageName,
		ProjectVolumePath:  b.projectVolumePath,
		PostCreateCommands: b.postCreateCommands,
		PostStartCommands:  b.postStartCommands,
	}, nil
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

	err = os.RemoveAll(b.projectVolumePath)
	if err != nil {
		return err
	}

	return nil
}

func (b *DevcontainerBuilder) Publish() error {
	projectLogger := b.loggerFactory.CreateProjectLogger(b.project.WorkspaceId, b.project.Name)
	defer projectLogger.Close()

	cliBuilder, err := b.getBuilderDockerClient()
	if err != nil {
		return err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cliBuilder,
	})

	//	todo: registry auth (from container registry store)
	return dockerClient.PushImage(b.buildImageName, nil, projectLogger)
}

func (b *DevcontainerBuilder) buildDevcontainer() error {
	projectLogger := b.loggerFactory.CreateProjectLogger(b.project.WorkspaceId, b.project.Name)
	defer projectLogger.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	cmd := []string{"devcontainer", "up", "--prebuild", "--workspace-folder", "/project"}
	if b.project.Build.Devcontainer.DevContainerFilePath != "" {
		cmd = append(cmd, "--config", filepath.Join("/project", b.project.Build.Devcontainer.DevContainerFilePath))
	}

	execConfig := types.ExecConfig{
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
			projectLogger.Write([]byte(lastLine + "\n"))

			if strings.Contains(lastLine, `{"outcome"`) {
				start := strings.Index(lastLine, "{")
				end := strings.LastIndex(lastLine, "}")
				if start != -1 && end != -1 && start < end {
					lastLine = lastLine[start : end+1]
				}

				err = json.Unmarshal([]byte(lastLine), &buildOutcome)
				if err != nil {
					//	todo: handle properly
					log.Error(err)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			//	todo: handle properly
			log.Error(err)
		}
	}(&buildOutcome)

	result, err := dockerClient.ExecSync(b.id, execConfig, w)
	if err != nil {
		return err
	}

	if buildOutcome.Outcome != "success" {
		return errors.New("devcontainer build failed")
	}

	b.user = buildOutcome.RemoteUser

	if result.ExitCode != 0 {
		return errors.New(result.StdErr)
	}

	builderCli, err := b.getBuilderDockerClient()
	if err != nil {
		return err
	}

	tag := b.project.Repository.Sha
	imageName := fmt.Sprintf("%s/p-%s:%s", b.localContainerRegistryServer, b.id, tag)

	_, err = builderCli.ContainerCommit(context.Background(), buildOutcome.ContainerId, container.CommitOptions{
		Reference: imageName,
	})
	if err != nil {
		return err
	}

	b.buildImageName = imageName

	return nil
}

func (b *DevcontainerBuilder) readConfiguration() error {
	projectLogger := b.loggerFactory.CreateProjectLogger(b.project.WorkspaceId, b.project.Name)
	defer projectLogger.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	cmd := []string{"devcontainer", "read-configuration", "--include-features-configuration", "--include-merged-configuration", "--workspace-folder", "/project"}
	if b.project.Build.Devcontainer.DevContainerFilePath != "" {
		cmd = append(cmd, "--config", filepath.Join("/project", b.project.Build.Devcontainer.DevContainerFilePath))
	}

	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
		Tty:          true,
	}

	result, err := dockerClient.ExecSync(b.id, execConfig, nil)
	if err != nil {
		return err
	}

	if result.ExitCode != 0 {
		return errors.New(result.StdErr)
	}

	// Convert result.Stdout to string
	stdoutStr := string(result.StdOut)
	stdoutStr = strings.TrimSuffix(stdoutStr, "\n")

	// Find the index of the last newline character
	lastNewline := strings.LastIndex(stdoutStr, "\n")

	// If there is a newline character, slice the string from the index after the newline to the end
	if lastNewline != -1 {
		stdoutStr = stdoutStr[lastNewline+1:]
	}

	// Create a new Root object
	root := &devcontainer.Root{}

	// Unmarshal the JSON into the Root object
	err = json.Unmarshal([]byte(stdoutStr), root)
	if err != nil {
		return err
	}

	postCreateCommands, err := devcontainer.ConvertCommands(root.MergedConfiguration.PostCreateCommands)
	if err != nil {
		projectLogger.Write([]byte(fmt.Sprintf("Error converting post create commands: %v\n", err)))
	}

	postStartCommands, err := devcontainer.ConvertCommands(root.MergedConfiguration.PostStartCommands)
	if err != nil {
		projectLogger.Write([]byte(fmt.Sprintf("Error converting post start commands: %v\n", err)))
	}

	b.postCreateCommands = append(b.postCreateCommands, postCreateCommands...)
	b.postStartCommands = append(b.postStartCommands, root.MergedConfiguration.Entrypoints...)
	b.postStartCommands = append(b.postStartCommands, postStartCommands...)

	return nil
}

func (b *DevcontainerBuilder) startContainer() error {
	ctx := context.Background()

	projectLogger := b.loggerFactory.CreateProjectLogger(b.project.WorkspaceId, b.project.Name)
	defer projectLogger.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	dockerClient := docker.NewDockerClient(docker.DockerClientConfig{
		ApiClient: cli,
	})

	//	todo: builder image from config
	err = dockerClient.PullImage("daytonaio/workspace-project", nil, projectLogger)
	if err != nil {
		return err
	}

	//	todo: mount folders
	_, err = cli.ContainerCreate(ctx, &container.Config{
		Image:      "daytonaio/workspace-project",
		Entrypoint: []string{"dockerd", "-H", fmt.Sprintf("tcp://0.0.0.0:%d", b.builderDockerPort), "-H", "unix:///var/run/docker.sock", "--insecure-registry", b.localContainerRegistryServer},
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", b.builderDockerPort)): {},
		},
	}, &container.HostConfig{
		Privileged: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: b.projectVolumePath,
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
