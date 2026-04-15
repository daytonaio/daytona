// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"fmt"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/runner/pkg/api/dto"
)

func (d *DockerClient) BuildSnapshot(ctx context.Context, req dto.BuildSnapshotRequestDTO) error {
	buildCtx, cancel := context.WithTimeout(ctx, time.Duration(d.buildTimeoutMin)*time.Minute)
	defer cancel()

	err := d.BuildImage(buildCtx, req)
	if err != nil {
		if buildCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("build timed out after %d minutes: %w", d.buildTimeoutMin, err)
		}
		return err
	}

	tag := req.Snapshot

	if req.PushToInternalRegistry {
		if req.Registry.Project == nil {
			return common_errors.NewBadRequestError(errors.New("project is required when pushing to internal registry"))
		}
		tag = fmt.Sprintf("%s/%s/%s", req.Registry.Url, *req.Registry.Project, req.Snapshot)
	}

	err = d.TagImage(ctx, req.Snapshot, tag)
	if err != nil {
		return err
	}

	if req.PushToInternalRegistry {
		err = d.PushImage(ctx, tag, req.Registry)
		if err != nil {
			return err
		}
	}

	return nil
}
