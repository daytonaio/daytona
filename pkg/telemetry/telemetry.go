// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import "io"

const ENABLED_HEADER = "X-Daytona-Telemetry-Enabled"
const SESSION_ID_HEADER = "X-Daytona-Session-Id"

type TelemetryService interface {
	io.Closer
	TrackCliEvent(event CliEvent, sessionId string, properties map[string]interface{}) error
	TrackServerEvent(event ServerEvent, sessionId string, properties map[string]interface{}) error
}

type AbstractTelemetryService struct {
	daytonaVersion string
	TelemetryService
}

func NewAbstractTelemetryService(daytonaVersion string) *AbstractTelemetryService {
	return &AbstractTelemetryService{
		daytonaVersion: daytonaVersion,
	}
}

func (t *AbstractTelemetryService) TrackCliEvent(event CliEvent, sessionId string, properties map[string]interface{}) error {
	properties["daytona_version"] = t.daytonaVersion
	return t.TelemetryService.TrackCliEvent(event, sessionId, properties)
}

func (t *AbstractTelemetryService) TrackServerEvent(event ServerEvent, sessionId string, properties map[string]interface{}) error {
	properties["daytona_version"] = t.daytonaVersion
	return t.TelemetryService.TrackServerEvent(event, sessionId, properties)
}
