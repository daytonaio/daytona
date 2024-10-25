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

func OpenTerminalSsh(activeProfile config.Profile, targetId string, projectName string, gpgKey string, sshOptions []string, args ...string) error {
	if err := config.EnsureSshConfigEntryAdded(activeProfile.Id, targetId, projectName, gpgKey); err != nil {
		return err
	}

	// Parse SSH options
	parsedOptions, err := parseSshOptions(sshOptions)
	if err != nil {
		return err
	}

	projectHostname := config.GetProjectHostname(activeProfile.Id, targetId, projectName)
	cmdArgs := buildCommandArgs(projectHostname, parsedOptions, args...)

	sshCommand := exec.Command("ssh", cmdArgs...)
	sshCommand.Stdin = os.Stdin
	sshCommand.Stdout = os.Stdout
	sshCommand.Stderr = os.Stderr

	return sshCommand.Run()
}

// parseSshOptions validates and parses the SSH options.
func parseSshOptions(sshOptions []string) (map[string]string, error) {
	parsedOptions := make(map[string]string)
	for _, option := range sshOptions {
		parts := strings.SplitN(option, "=", 2)
		if len(parts) == 1 {
			return nil, fmt.Errorf("no argument after keyword %q", parts[0])
		}
		if len(parts) != 2 || strings.Count(option, "=") > 1 {
			return nil, fmt.Errorf("bad configuration option: %s", option)
		}
		parsedOptions[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return parsedOptions, nil
}

func buildCommandArgs(projectHostname string, parsedOptions map[string]string, args ...string) []string {
	cmdArgs := []string{projectHostname}
	for key, value := range parsedOptions {
		cmdArgs = append(cmdArgs, "-o", fmt.Sprintf("%s=%s", key, value))
	}
	return append(cmdArgs, args...)
}
