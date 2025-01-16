// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
)

func (rj *RunnerJob) uninstallProvider(_ context.Context, j *models.Job) error {
	if j.Metadata == nil {
		return errors.New("metadata is required")
	}

	return rj.providerManager.UninstallProvider(*j.Metadata)
}
