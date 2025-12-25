// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import "context"

// GetContainerIpAddress extracts the IP address from a ContainerJSON
func GetContainerIpAddress(ctx context.Context, container ContainerJSON) string {
	if container.NetworkSettings == nil || container.NetworkSettings.Networks == nil {
		return ""
	}

	// Try to get IP from bridge network
	if networkSettings, ok := container.NetworkSettings.Networks["bridge"]; ok && networkSettings != nil {
		if ipMap, ok := networkSettings.(map[string]interface{}); ok {
			if ip, ok := ipMap["IPAddress"].(string); ok {
				return ip
			}
		}
	}

	// Fallback to the main IPAddress field
	return container.NetworkSettings.IPAddress
}
