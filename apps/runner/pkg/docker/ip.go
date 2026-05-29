// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
)

// GetContainerIpAddress returns the IP address of the container on its primary
// reachable network.
//
// Resolution order:
//  1. Android-device containers prefer their link network IP. Their ADB/emulator
//     socat forwarders bind to eth0, which for followers is the link network.
//  2. The runner bridge network (daytona-internal inter-sandbox network).
//  3. The default Docker bridge.
//
// Returns an empty string when no known network is attached.
func GetContainerIpAddress(ctx context.Context, container *container.InspectResponse) string {
	if container == nil || container.NetworkSettings == nil || container.NetworkSettings.Networks == nil {
		return ""
	}

	// Only Android followers (primary NetworkMode set to the link network) need
	// the link IP — that's where their eth0-bound ADB/emulator forwarders listen.
	// Android containers attached to a link network as a secondary network must
	// still report their primary bridge IP.
	if isAndroidDeviceContainer(container) && container.HostConfig != nil &&
		strings.HasPrefix(string(container.HostConfig.NetworkMode), linkNetworkPrefix) {
		if ip := linkNetworkIP(container); ip != "" {
			return ip
		}
	}

	if networkSettings, ok := container.NetworkSettings.Networks[RUNNER_BRIDGE_NETWORK_NAME]; ok && networkSettings != nil && networkSettings.IPAddress != "" {
		return networkSettings.IPAddress
	}

	if networkSettings, ok := container.NetworkSettings.Networks["bridge"]; ok && networkSettings != nil && networkSettings.IPAddress != "" {
		return networkSettings.IPAddress
	}

	return ""
}

// linkNetworkIP returns the IP address of the first attached network whose name
// starts with linkNetworkPrefix, or an empty string when none is attached.
func linkNetworkIP(container *container.InspectResponse) string {
	for name, settings := range container.NetworkSettings.Networks {
		if settings == nil || settings.IPAddress == "" {
			continue
		}
		if strings.HasPrefix(name, linkNetworkPrefix) {
			return settings.IPAddress
		}
	}
	return ""
}
