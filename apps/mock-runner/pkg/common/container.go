// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

// GetContainerIpAddress extracts the IP address from container inspect response
func GetContainerIpAddress(ctx context.Context, info container.InspectResponse) string {
	for _, network := range info.NetworkSettings.Networks {
		if network.IPAddress != "" {
			return network.IPAddress
		}
	}
	return ""
}



