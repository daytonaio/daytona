// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"encoding/json"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
)

func (pj *RunnerJob) updateProvider(ctx context.Context, j *models.Job) error {
	metadata := j.Metadata

	var metadataJson services.ProviderJobMetadata
	err := json.Unmarshal([]byte(*j.Metadata), &metadataJson)
	if err != nil {
		return err
	}

	uninstallMetadataJson, err := json.Marshal(metadataJson.Name)
	if err != nil {
		return err
	}

	uninstallMetadata := string(uninstallMetadataJson)
	j.Metadata = &uninstallMetadata

	err = pj.uninstallProvider(ctx, j)
	if err != nil {
		return err
	}

	j.Metadata = metadata

	return pj.installProvider(ctx, j)
}
