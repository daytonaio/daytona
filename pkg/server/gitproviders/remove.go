// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitproviders

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *GitProviderService) RemoveGitProvider(ctx context.Context, gitProviderId string) error {
	gitProvider, err := s.configStore.Find(ctx, gitProviderId)
	if err != nil {
		return s.handleRemoveGitProviderConfigError(ctx, nil, err)
	}

	err = s.detachWorkspaceTemplates(ctx, gitProvider.Id)
	if err != nil {
		return s.handleRemoveGitProviderConfigError(ctx, gitProvider, err)
	}

	err = s.configStore.Delete(ctx, gitProvider)
	return s.handleRemoveGitProviderConfigError(ctx, gitProvider, err)
}

func (s *GitProviderService) handleRemoveGitProviderConfigError(ctx context.Context, gpc *models.GitProviderConfig, err error) error {
	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	clientId := telemetry.ClientId(ctx)

	eventName := telemetry.GitProviderConfigEventLifecycleDeleted
	if err != nil {
		eventName = telemetry.GitProviderConfigEventLifecycleDeletionFailed
	}
	event := telemetry.NewGitProviderConfigEvent(eventName, gpc, err, nil)

	telemetryError := s.trackTelemetryEvent(event, clientId)
	if telemetryError != nil {
		log.Trace(telemetryError)
	}

	return err
}
