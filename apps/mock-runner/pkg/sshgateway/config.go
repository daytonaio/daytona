// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sshgateway

import "os"

// IsSSHGatewayEnabled checks if SSH gateway is enabled via environment variable
func IsSSHGatewayEnabled() bool {
	return os.Getenv("SSH_GATEWAY_ENABLE") == "true"
}
