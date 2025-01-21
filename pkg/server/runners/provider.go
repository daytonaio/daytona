// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runners

import (
	"context"
	"encoding/json"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/runner"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *RunnerService) ListProviders(ctx context.Context, runnerId *string) ([]models.ProviderInfo, error) {
	var metadatas []*models.RunnerMetadata

	if runnerId == nil {
		var err error
		metadatas, err = s.runnerMetadataStore.List(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		metadata, err := s.runnerMetadataStore.Find(ctx, *runnerId)
		if err != nil {
			return nil, err
		}
		metadatas = []*models.RunnerMetadata{metadata}
	}

	providers := []models.ProviderInfo{}
	for _, metadata := range metadatas {
		providers = append(providers, metadata.Providers...)
	}

	return providers, nil
}

func (s *RunnerService) ListProvidersForInstall(ctx context.Context, registryUrl string) ([]services.ProviderDTO, error) {
	providersManifest, err := runner.GetProvidersManifest(registryUrl)
	if err != nil {
		return nil, err
	}

	return providersManifest.GetProviderListFromManifest(), nil
}

func (s *RunnerService) InstallProvider(ctx context.Context, runnerId, name, version, registryUrl string) error {
	params := providerActionParams{
		providerName:    name,
		providerVersion: &version,
		eventName:       telemetry.RunnerEventProviderInstalled,
		errEventName:    telemetry.RunnerEventProviderInstallationFailed,
	}

	downloadUrls, err := runner.GetProviderDownloadUrls(name, version, registryUrl)
	if err != nil {
		return s.handleProviderActionError(ctx, params, err)
	}

	runner, err := s.runnerStore.Find(ctx, runnerId)
	if err != nil {
		return s.handleProviderActionError(ctx, params, err)
	}

	params.runner = runner

	metadata, err := json.Marshal(services.ProviderMetadata{
		Name:         name,
		Version:      version,
		DownloadUrls: downloadUrls,
	})
	if err != nil {
		return s.handleProviderActionError(ctx, params, err)
	}

	err = s.createJob(ctx, runnerId, models.JobActionInstallProvider, string(metadata))
	return s.handleProviderActionError(ctx, params, err)
}

func (s *RunnerService) UninstallProvider(ctx context.Context, runnerId string, providerName string) error {
	params := providerActionParams{
		providerName: providerName,
		eventName:    telemetry.RunnerEventProviderUninstalled,
		errEventName: telemetry.RunnerEventProviderUninstallationFailed,
	}

	runner, err := s.runnerStore.Find(ctx, runnerId)
	if err != nil {
		return s.handleProviderActionError(ctx, params, err)
	}

	params.runner = runner

	err = s.createJob(ctx, runnerId, models.JobActionUninstallProvider, providerName)
	return s.handleProviderActionError(ctx, params, err)
}

func (s *RunnerService) UpdateProvider(ctx context.Context, runnerId, name, version, registryUrl string) error {
	params := providerActionParams{
		providerName:    name,
		providerVersion: &version,
		eventName:       telemetry.RunnerEventProviderUpdated,
		errEventName:    telemetry.RunnerEventProviderUpdateFailed,
	}

	r, err := s.runnerStore.Find(ctx, runnerId)
	if err != nil {
		return s.handleProviderActionError(ctx, params, err)
	}

	params.runner = r

	downloadUrls, err := runner.GetProviderDownloadUrls(name, version, registryUrl)
	if err != nil {
		return s.handleProviderActionError(ctx, params, err)
	}

	providerDto := services.ProviderMetadata{
		DownloadUrls: downloadUrls,
		Version:      version,
		Name:         name,
	}

	metadata, err := json.Marshal(providerDto)
	if err != nil {
		return s.handleProviderActionError(ctx, params, err)
	}

	err = s.createJob(ctx, r.Id, models.JobActionUpdateProvider, string(metadata))
	return s.handleProviderActionError(ctx, params, err)
}

type providerActionParams struct {
	runner          *models.Runner
	eventName       telemetry.RunnerEventName
	errEventName    telemetry.RunnerEventName
	providerName    string
	providerVersion *string
}

func (s *RunnerService) handleProviderActionError(ctx context.Context, params providerActionParams, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	eventName := params.eventName
	if err != nil {
		eventName = params.errEventName
	}

	clientId := telemetry.ClientId(ctx)

	extras := map[string]interface{}{
		"provider_name": params.providerName,
	}
	if params.providerVersion != nil {
		extras["provider_version"] = *params.providerVersion
	}

	event := telemetry.NewRunnerEvent(eventName, params.runner, err, extras)
	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
