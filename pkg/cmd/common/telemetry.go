// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/posthogservice"
	"github.com/daytonaio/daytona/pkg/telemetry"
)

var telemetryService telemetry.TelemetryService

func TrackTelemetryEvent(event telemetry.Event, clientId string) error {
	if telemetryService == nil {
		return nil
	}

	return telemetryService.Track(event, clientId)
}

func CloseTelemetryService() error {
	if telemetryService == nil {
		return nil
	}

	return telemetryService.Close()
}

func init() {
	telemetryEnabled := config.TelemetryEnabled()

	if !telemetryEnabled {
		return
	}

	source := telemetry.CLI_SOURCE
	if common.AgentMode() {
		source = telemetry.CLI_AGENT_MODE_SOURCE
	}

	telemetryService = posthogservice.NewTelemetryService(posthogservice.PosthogServiceConfig{
		ApiKey:   internal.PosthogApiKey,
		Endpoint: internal.PosthogEndpoint,
		Version:  internal.Version,
		Source:   source,
	})
}
