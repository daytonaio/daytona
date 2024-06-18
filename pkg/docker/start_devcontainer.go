// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package docker

import "github.com/daytonaio/daytona/pkg/ssh"

func (d *DockerClient) startDevcontainerProject(opts *CreateProjectOptions, sshClient *ssh.Client) (RemoteUser, error) {
	return d.createProjectFromDevcontainer(opts, false, sshClient)
}
