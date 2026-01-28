// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"fmt"
	"strings"
)

// ParseSSHCommand parses the SSH command string returned by the API
// Expected formats:
// - "ssh token@host" (port 22)
// - "ssh -p port token@host"
func ParseSSHCommand(sshCommand string) ([]string, error) {
	parts := strings.Fields(sshCommand)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid SSH command format: %s", sshCommand)
	}

	// Skip the "ssh" part
	args := parts[1:]

	return args, nil
}
