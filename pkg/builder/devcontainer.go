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

	log "github.com/sirupsen/logrus"

	"github.com/daytonaio/daytona/pkg/builder/devcontainer"
	"github.com/daytonaio/daytona/pkg/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
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
	buildImageName     string
	user               string
	postCreateCommands []string
	postStartCommands  []string
}

func (b *DevcontainerBuilder) Build() (*BuildResult, error) {
	err := b.startContainer()
	if err != nil {
		return nil, err
	}

	err = b.startDocker()
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

	ctx := context.Background()
	cliBuilder, err := b.getBuilderDockerClient()
	if err != nil {
		return err
	}

	reader, err := cliBuilder.ImagePush(ctx, b.buildImageName, image.PushOptions{
		//	todo: registry auth (from container registry store)
		RegistryAuth: "empty", //	make sure that something is passed, as with "" will throw an X-Auth error
	})
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		projectLogger.Write([]byte(scanner.Text() + "\n"))
	}

	return nil
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
		//	todo: parse reason
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
	imageName := fmt.Sprintf("%s:%s", b.localContainerRegistryServer+"/p-"+b.id, tag)

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

	b.postCreateCommands = append(b.postCreateCommands, root.MergedConfiguration.PostCreateCommands...)
	b.postStartCommands = append(b.postStartCommands, root.MergedConfiguration.Entrypoints...)
	b.postStartCommands = append(b.postStartCommands, root.MergedConfiguration.PostStartCommands...)

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

	//	todo builder image from config
	reader, err := cli.ImagePull(ctx, "daytonaio/workspace-project", image.PullOptions{
		//	auth for builder image
	})
	if err != nil {
		return err
	}
	_, err = io.Copy(projectLogger, reader)
	if err != nil {
		return err
	}

	dir := "/tmp/" + b.id + "/var/lib/docker"
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	//	todo: mount folders
	_, err = cli.ContainerCreate(ctx, &container.Config{
		Image:      "daytonaio/workspace-project",
		Entrypoint: []string{"sleep", "infinity"},
	}, &container.HostConfig{
		NetworkMode: "host",
		Privileged:  true,
		Binds: []string{
			b.projectVolumePath + ":/project",
			filepath.Dir(b.getLocalDockerSocket()) + ":/tmp/docker",
			dir + ":/var/lib/docker",
		},
	}, nil, nil, b.id)
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, b.id, container.StartOptions{}); err != nil {
		return err
	}

	return nil
}

func (b *DevcontainerBuilder) startDocker() error {
	ctx := context.Background()

	projectLogger := b.loggerFactory.CreateProjectLogger(b.project.WorkspaceId, b.project.Name)
	defer projectLogger.Close()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"dockerd", "-H", "unix:///tmp/docker/docker.sock", "-H", "unix:///var/run/docker.sock", "--insecure-registry", b.localContainerRegistryServer},
	}

	execResp, err := cli.ContainerExecCreate(ctx, b.id, execConfig)
	if err != nil {
		return err
	}

	execStartCheck := types.ExecStartCheck{
		Detach: false,
		Tty:    false,
	}

	attachResp, err := cli.ContainerExecAttach(ctx, execResp.ID, execStartCheck)
	if err != nil {
		return err
	}
	defer attachResp.Close()

	go func() {
		_, err = io.Copy(projectLogger, attachResp.Reader)
		if err != nil {
			log.Errorf("error copying output: %v", err)
		}
	}()

	return nil
}

func (b *DevcontainerBuilder) getBuilderDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.WithHost(fmt.Sprintf("unix://%s", b.getLocalDockerSocket())), client.WithAPIVersionNegotiation())
}

func (b *DevcontainerBuilder) getLocalDockerSocket() string {
	return filepath.Join("/tmp", b.id, "docker", "docker.sock")
}
