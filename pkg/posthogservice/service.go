// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package posthogservice

import (
	"log"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/posthog/posthog-go"
	"github.com/sirupsen/logrus"
)

type PosthogServiceConfig struct {
	ApiKey   string
	Endpoint string
}

func NewTelemetryService(config PosthogServiceConfig) telemetry.TelemetryService {
	client, _ := posthog.NewWithConfig(config.ApiKey, posthog.Config{
		Endpoint: config.Endpoint,
		Logger:   posthog.StdLogger(log.Default()),
		Verbose:  true,
	})
	posthogService := &posthogService{
		client:                   client,
		AbstractTelemetryService: telemetry.NewAbstractTelemetryService(internal.Version),
	}

	posthogService.AbstractTelemetryService.TelemetryService = posthogService

	return posthogService
}

type posthogService struct {
	*telemetry.AbstractTelemetryService

	client posthog.Client
}

func (p *posthogService) Close() error {
	return p.client.Close()
}

func (p *posthogService) TrackCliEvent(event telemetry.CliEvent, sessionId string, properties map[string]interface{}) error {
	logrus.Info(event)
	return p.client.Enqueue(posthog.Capture{
		DistinctId: sessionId,
		Event:      string(event),
		Properties: properties,
	})
}

func (p *posthogService) TrackServerEvent(event telemetry.ServerEvent, sessionId string, properties map[string]interface{}) error {
	return p.client.Enqueue(posthog.Capture{
		DistinctId: sessionId,
		Event:      string(event),
		Properties: properties,
	})
}
