// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runners

import (
	"context"
	"encoding/json"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
)

func (s *RunnerService) ListProviders(ctx context.Context) ([]models.ProviderInfo, error) {
	metadatas, err := s.runnerMetadataStore.List(ctx)
	if err != nil {
		return nil, err
	}

	providers := []models.ProviderInfo{}
	for _, metadata := range metadatas {
		providers = append(providers, metadata.Providers...)
	}

	return providers, nil
}

func (s *RunnerService) InstallProvider(ctx context.Context, runnerId string, providerMetadata services.InstallProviderDTO) error {
	metadata, err := json.Marshal(providerMetadata)
	if err != nil {
		return err
	}

	return s.createJob(ctx, runnerId, models.JobActionInstallProvider, string(metadata))
}

func (s *RunnerService) UninstallProvider(ctx context.Context, runnerId string, providerName string) error {
	return s.createJob(ctx, runnerId, models.JobActionUninstallProvider, providerName)
}

func (s *RunnerService) UpdateProvider(ctx context.Context, runnerId string, providerName string, downloadUrls services.ProviderDownloadUrlsDTO) error {
	metadata, err := json.Marshal(services.InstallProviderDTO{
		Name:                    providerName,
		ProviderDownloadUrlsDTO: downloadUrls,
	})
	if err != nil {
		return err
	}

	return s.createJob(ctx, runnerId, models.JobActionUpdateProvider, string(metadata))
}
