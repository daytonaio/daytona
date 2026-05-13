// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

// GetContainerIpAddress returns the IP address of the container on its primary
// reachable network — the runner bridge network, falling back to the default
// Docker bridge. Secondary attachments such as per-owner link networks are
// deliberately ignored: callers use the returned IP as the iptables source
// address for per-sandbox network rules (block / allow / egress limiter), and
// the sandbox's outbound traffic always egresses via its default-route
// interface, which is the runner bridge.
//
// Returns an empty string when no known network is attached.
func GetContainerIpAddress(ctx context.Context, container *container.InspectResponse) string {
	if container == nil || container.NetworkSettings == nil || container.NetworkSettings.Networks == nil {
		return ""
	}

	if networkSettings, ok := container.NetworkSettings.Networks[RUNNER_BRIDGE_NETWORK_NAME]; ok && networkSettings != nil && networkSettings.IPAddress != "" {
		return networkSettings.IPAddress
	}

	if networkSettings, ok := container.NetworkSettings.Networks["bridge"]; ok && networkSettings != nil && networkSettings.IPAddress != "" {
		return networkSettings.IPAddress
	}

	return ""
}
