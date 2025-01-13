// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package apikeys

import (
	"context"

	"github.com/daytonaio/daytona/pkg/models"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/telemetry"
	log "github.com/sirupsen/logrus"
)

func (s *ApiKeyService) ListClientKeys(ctx context.Context) ([]*services.ApiKeyDTO, error) {
	keys, err := s.apiKeyStore.List(ctx)
	if err != nil {
		return nil, err
	}

	clientKeys := []*services.ApiKeyDTO{}

	for _, key := range keys {
		if key.Type == models.ApiKeyTypeClient {
			clientKeys = append(clientKeys, &services.ApiKeyDTO{
				Type: key.Type,
				Name: key.Name,
			})
		}
	}

	return clientKeys, nil
}

func (s *ApiKeyService) Revoke(ctx context.Context, name string) error {
	apiKey, err := s.apiKeyStore.FindByName(ctx, name)
	if err != nil {
		return err
	}

	return s.apiKeyStore.Delete(ctx, apiKey)
}

func (s *ApiKeyService) Generate(ctx context.Context, keyType models.ApiKeyType, name string) (string, error) {
	key := s.generateRandomKey(name)

	apiKey := &models.ApiKey{
		KeyHash: s.getKeyHash(key),
		Type:    keyType,
		Name:    name,
	}

	err := s.apiKeyStore.Save(ctx, apiKey)
	if err != nil {
		return "", s.handleGenerateApiKeyError(ctx, apiKey, err)
	}

	return key, s.handleGenerateApiKeyError(ctx, apiKey, nil)
}

func (s *ApiKeyService) GetApiKeyName(ctx context.Context, apiKey string) (string, error) {
	key, err := s.apiKeyStore.Find(ctx, s.getKeyHash(apiKey))
	if err != nil {
		return "", err
	}

	return key.Name, nil
}

func (s *ApiKeyService) handleGenerateApiKeyError(ctx context.Context, key *models.ApiKey, err error) error {
	if key.Type != models.ApiKeyTypeClient {
		return err
	}

	if !telemetry.TelemetryEnabled(ctx) {
		return err
	}

	eventName := telemetry.ApiKeyEventLifecycleCreated
	if err != nil {
		eventName = telemetry.ApiKeyEventLifecycleCreationFailed
	}

	event := telemetry.NewApiKeyEvent(eventName, key, err, nil)

	telemetryErr := s.trackTelemetryEvent(event, telemetry.ClientId(ctx))
	if telemetryErr != nil {
		log.Trace(telemetryErr)
	}

	return err
}
