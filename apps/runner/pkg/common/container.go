// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

func GetContainerIpAddress(ctx context.Context, container *container.InspectResponse) string {
	if container.NetworkSettings == nil {
		return ""
	}

	if container.NetworkSettings.Networks == nil {
		return ""
	}

	if networkSettings, ok := container.NetworkSettings.Networks["bridge"]; ok && networkSettings != nil {
		return networkSettings.IPAddress
	}

	return ""
}
