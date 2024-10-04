// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"os"
	"os/exec"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

func OpenTerminalSsh(activeProfile config.Profile, workspaceId string, projectName string, gpgForward bool, args ...string) error {
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName, gpgForward)
	if err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	cmdArgs := []string{projectHostname}
	cmdArgs = append(cmdArgs, args...)

	sshCommand := exec.Command("ssh", cmdArgs...)
	sshCommand.Stdin = os.Stdin
	sshCommand.Stdout = os.Stdout
	sshCommand.Stderr = os.Stderr

	return sshCommand.Run()
}
