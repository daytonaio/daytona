// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/daytonaio/daytona/pkg/builder/detect"
	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/daytonaio/daytona/pkg/ssh"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/api/types"
)

func (d *DockerClient) StartProject(opts *CreateProjectOptions, daytonaDownloadUrl string) error {
	var err error
	containerUser := opts.Project.User

	var sshClient *ssh.Client
	if opts.SshSessionConfig != nil {
		sshClient, err = ssh.NewClient(opts.SshSessionConfig)
		if err != nil {
			return err
		}
		defer sshClient.Close()
	}

	builderType, err := detect.DetectProjectBuilderType(opts.Project, opts.ProjectDir, sshClient)
	if err != nil {
		return err
	}

	switch builderType {
	case detect.BuilderTypeDevcontainer:
		var remoteUser RemoteUser
		remoteUser, err = d.startDevcontainerProject(opts, sshClient)
		containerUser = string(remoteUser)
	case detect.BuilderTypeImage:
		err = d.startImageProject(opts)
	default:
		return fmt.Errorf("unknown builder type: %s", builderType)
	}

	if err != nil {
		return err
	}

	return d.startDaytonaAgent(opts.Project, containerUser, daytonaDownloadUrl, opts.LogWriter)
}

func (d *DockerClient) startDaytonaAgent(project *workspace.Project, containerUser, daytonaDownloadUrl string, logWriter io.Writer) error {
	errChan := make(chan error)

	go func() {
		result, err := d.ExecSync(d.GetProjectContainerName(project), types.ExecConfig{
			Cmd:          []string{"bash", "-c", util.GetProjectStartScript(daytonaDownloadUrl, project.ApiKey)},
			AttachStdout: true,
			AttachStderr: true,
			User:         containerUser,
		}, logWriter)
		if err != nil {
			errChan <- err
		}

		if result.ExitCode != 0 {
			errChan <- errors.New(result.StdErr)
		}
	}()

	go func() {
		// TODO: Figure out how to check if the agent is running here
		time.Sleep(5 * time.Second)
		errChan <- nil
	}()

	return <-errChan
}
