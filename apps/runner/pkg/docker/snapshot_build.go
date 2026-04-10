// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package docker

import (
	"context"
	"errors"
	"fmt"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/daytonaio/runner/pkg/api/dto"
)

func (d *DockerClient) BuildSnapshot(ctx context.Context, req dto.BuildSnapshotRequestDTO) error {
	err := d.BuildImage(ctx, req)
	if err != nil {
		return err
	}

	tag := req.GetSnapshot()

	if req.GetPushToInternalRegistry() {
		if req.GetRegistry().GetProject() == "" {
			return common_errors.NewBadRequestError(errors.New("project is required when pushing to internal registry"))
		}
		tag = fmt.Sprintf("%s/%s/%s", req.GetRegistry().GetUrl(), req.GetRegistry().GetProject(), req.GetSnapshot())
	}

	err = d.TagImage(ctx, req.GetSnapshot(), tag)
	if err != nil {
		return err
	}

	if req.GetPushToInternalRegistry() {
		err = d.PushImage(ctx, tag, req.GetRegistry())
		if err != nil {
			return err
		}
	}

	return nil
}
