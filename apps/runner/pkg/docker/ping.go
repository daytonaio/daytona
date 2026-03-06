// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
)

func (d *DockerClient) Ping(ctx context.Context) error {
	_, err := d.apiClient.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker ping failed: %w", err)
	}
	return nil
}
