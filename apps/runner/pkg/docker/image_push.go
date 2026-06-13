// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"io"
	"time"

	"github.com/daytonaio/common-go/pkg/log"
	"github.com/daytonaio/runner/pkg/api/dto"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/jsonmessage"
	"go.opentelemetry.io/otel/codes"
)

// PushImage pushes a tagged local image to a registry. sandboxId, when
// non-nil, is attached as the sandbox.id attribute on the emitted span and
// metrics so push activity can be attributed back to the sandbox that
// triggered it (sandbox snapshot, backup, etc.). Pass nil for paths that have
// no sandbox context (snapshot pull-through, snapshot build).
func (d *DockerClient) PushImage(ctx context.Context, imageName string, reg *dto.RegistryDTO, sandboxId *string) (retErr error) {
	ctx, span := StartRegistrySpan(ctx, "docker.PushImage", RegistryOpPush, sandboxId, imageName)
	defer span.End()

	start := time.Now()
	defer func() {
		RecordRegistryOp(ctx, RegistryOpPush, sandboxId, imageName, start, retErr)
		if retErr != nil {
			span.RecordError(retErr)
			span.SetStatus(codes.Error, retErr.Error())
		}
	}()

	d.logger.InfoContext(ctx, "Pushing image", "imageName", imageName)

	responseBody, err := d.apiClient.ImagePush(ctx, imageName, image.PushOptions{
		RegistryAuth: getRegistryAuth(reg),
	})
	if err != nil {
		return err
	}
	defer responseBody.Close()

	err = jsonmessage.DisplayJSONMessagesStream(responseBody, io.Writer(&log.DebugLogWriter{}), 0, true, nil)
	if err != nil {
		return err
	}

	d.logger.InfoContext(ctx, "Image pushed successfully", "imageName", imageName)

	return nil
}
