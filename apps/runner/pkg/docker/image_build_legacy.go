// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/pkg/jsonmessage"
)

func (d *DockerClient) runDockerImageBuildLegacy(
	ctx context.Context,
	dockerBuildContext io.Reader,
	buildOpts build.ImageBuildOptions,
	writer io.Writer,
) error {
	resp, err := d.apiClient.ImageBuild(ctx, dockerBuildContext, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer resp.Body.Close()

	return jsonmessage.DisplayJSONMessagesStream(resp.Body, writer, 0, true, nil)
}
