// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
)

func (pj *RunnerJob) updateProvider(ctx context.Context, j *models.Job) error {
	if j.Metadata == nil {
		return errors.New("metadata is required")
	}

	installMetadata := j.Metadata

	var metadata services.ProviderMetadata
	err := json.Unmarshal([]byte(*j.Metadata), &metadata)
	if err != nil {
		return err
	}

	j.Metadata = &metadata.Name

	err = pj.uninstallProvider(ctx, j)
	if err != nil {
		return err
	}

	j.Metadata = installMetadata

	return pj.installProvider(ctx, j)
}
