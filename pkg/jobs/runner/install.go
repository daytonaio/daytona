// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"context"
	"encoding/json"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
)

func (rj *RunnerJob) installProvider(ctx context.Context, j *models.Job) error {
	var installProviderDto services.InstallProviderDTO

	err := json.Unmarshal([]byte(*j.Metadata), &installProviderDto)
	if err != nil {
		return err
	}

	downloadPath, err := rj.providerManager.DownloadProvider(ctx, installProviderDto.DownloadUrls, installProviderDto.Name)
	if err != nil {
		return err
	}

	return rj.providerManager.RegisterProvider(downloadPath, false)
}
