// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package posthogservice

import (
	"fmt"
	"log"

	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/posthog/posthog-go"
	"github.com/sirupsen/logrus"
)

type PosthogServiceConfig struct {
	ApiKey   string
	Endpoint string
	Version  string
}

func NewTelemetryService(config PosthogServiceConfig) telemetry.TelemetryService {
	client, _ := posthog.NewWithConfig(config.ApiKey, posthog.Config{
		Endpoint: config.Endpoint,
		Logger:   posthog.StdLogger(log.New(logrus.StandardLogger().WriterLevel(logrus.TraceLevel), "", 0)),
		Verbose:  true,
	})
	posthogService := &posthogService{
		client:                   client,
		AbstractTelemetryService: telemetry.NewAbstractTelemetryService(config.Version),
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

func (p *posthogService) TrackCliEvent(event telemetry.CliEvent, clientId string, properties map[string]interface{}) error {
	p.AbstractTelemetryService.SetCommonProps(properties)
	return p.client.Enqueue(posthog.Capture{
		DistinctId: clientId,
		Event:      string(event),
		Properties: properties,
	})
}

func (p *posthogService) TrackServerEvent(event telemetry.ServerEvent, clientId string, properties map[string]interface{}) error {
	p.AbstractTelemetryService.SetCommonProps(properties)
	return p.client.Enqueue(posthog.Capture{
		DistinctId: clientId,
		Event:      string(event),
		Properties: properties,
	})
}

func (p *posthogService) TrackBuildRunnerEvent(event telemetry.BuildRunnerEvent, buildRunnerId string, properties map[string]interface{}) error {
	p.AbstractTelemetryService.SetCommonProps(properties)
	return p.client.Enqueue(posthog.Capture{
		DistinctId: fmt.Sprintf("build-runner-%s", buildRunnerId),
		Event:      string(event),
		Properties: properties,
	})
}
