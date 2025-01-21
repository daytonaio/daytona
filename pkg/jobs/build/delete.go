// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
)

func (bj *BuildJob) delete(ctx context.Context, j *models.Job, force bool) error {
	b, err := bj.findBuild(ctx, j.ResourceId)
	if err != nil {
		return err
	}

	if b.Image != nil {
		return bj.deleteImage(ctx, *b.Image, force)
	}

	return nil
}
