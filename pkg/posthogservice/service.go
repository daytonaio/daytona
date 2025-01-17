// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package posthogservice

import (
	"log"

	"github.com/daytonaio/daytona/pkg/telemetry"
	"github.com/posthog/posthog-go"
	"github.com/sirupsen/logrus"
)

type PosthogServiceConfig struct {
	ApiKey   string
	Endpoint string
	Version  string
	Source   telemetry.TelemetrySource
}

func NewTelemetryService(config PosthogServiceConfig) telemetry.TelemetryService {
	client, _ := posthog.NewWithConfig(config.ApiKey, posthog.Config{
		Endpoint: config.Endpoint,
		Logger:   posthog.StdLogger(log.New(logrus.StandardLogger().WriterLevel(logrus.TraceLevel), "", 0)),
		Verbose:  true,
	})
	posthogService := &posthogService{
		client:  client,
		version: config.Version,
		source:  config.Source,
	}

	return posthogService
}

type posthogService struct {
	client  posthog.Client
	version string
	source  telemetry.TelemetrySource
}

func (p *posthogService) Close() error {
	return p.client.Close()
}

func (p *posthogService) Track(event telemetry.Event, clientId string) error {
	props := event.Props()

	telemetry.SetCommonProps(p.version, p.source, props)
	return p.client.Enqueue(posthog.Capture{
		DistinctId: clientId,
		Event:      event.Name(),
		Properties: props,
	})
}
