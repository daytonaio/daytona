// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

func OpenTerminalSsh(activeProfile config.Profile, workspaceId string, projectName string, gpgKey string, sshOptions []string, args ...string) error {
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName, gpgKey)
	if err != nil {
		return err
	}

	parsedOptions := make(map[string]string)
	for _, option := range sshOptions {
		parts := strings.SplitN(option, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("bad configuration option: %s, must be KEY=VALUE", option)
		}
		parsedOptions[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	cmdArgs := []string{projectHostname}

	for key, value := range parsedOptions {
		cmdArgs = append(cmdArgs, "-o", fmt.Sprintf("%s=%s", key, value))
	}

	cmdArgs = append(cmdArgs, args...)

	sshCommand := exec.Command("ssh", cmdArgs...)
	sshCommand.Stdin = os.Stdin
	sshCommand.Stdout = os.Stdout
	sshCommand.Stderr = os.Stderr

	return sshCommand.Run()
}
