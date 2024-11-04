// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/daytonaio/daytona/pkg/build/detect"
	"github.com/daytonaio/daytona/pkg/provider/util"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/docker/docker/api/types/container"
)

func (d *DockerClient) StartWorkspace(opts *CreateWorkspaceOptions, daytonaDownloadUrl string) error {
	var err error
	containerUser := opts.Workspace.User

	builderType, err := detect.DetectWorkspaceBuilderType(opts.Workspace.BuildConfig, opts.WorkspaceDir, opts.SshClient)
	if err != nil {
		return err
	}

	switch builderType {
	case detect.BuilderTypeDevcontainer:
		var remoteUser RemoteUser
		remoteUser, err = d.startDevcontainerWorkspace(opts)
		containerUser = string(remoteUser)
	case detect.BuilderTypeImage:
		err = d.startImageWorkspace(opts)
	default:
		return fmt.Errorf("unknown builder type: %s", builderType)
	}

	if err != nil {
		return err
	}

	return d.startDaytonaAgent(opts.Workspace, containerUser, daytonaDownloadUrl, opts.LogWriter)
}

func (d *DockerClient) startDaytonaAgent(w *workspace.Workspace, containerUser, daytonaDownloadUrl string, logWriter io.Writer) error {
	errChan := make(chan error)

	r, pipeW := io.Pipe()
	writer := io.MultiWriter(pipeW, logWriter)

	go func() {
		result, err := d.ExecSync(d.GetWorkspaceContainerName(w), container.ExecOptions{
			Cmd:          []string{"sh", "-c", util.GetWorkspaceStartScript(daytonaDownloadUrl, w.ApiKey)},
			AttachStdout: true,
			AttachStderr: true,
			User:         containerUser,
		}, writer)
		if err != nil {
			errChan <- err
			return
		}

		if result.ExitCode != 0 {
			errChan <- errors.New(result.StdErr)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "Daytona Agent started") {
				errChan <- nil
				return
			}
		}
	}()

	return <-errChan
}
