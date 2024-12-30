// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
	"fmt"
	"io"

	"github.com/daytonaio/daytona/internal"
	"github.com/google/uuid"
)

type TelemetryService interface {
	io.Closer
	TrackCliEvent(event CliEvent, clientId string, properties map[string]interface{}) error
	TrackServerEvent(event ServerEvent, clientId string, properties map[string]interface{}) error
	TrackBuildRunnerEvent(event BuildRunnerEvent, clientId string, properties map[string]interface{}) error
	SetCommonProps(properties map[string]interface{})
}

func TelemetryEnabled(ctx context.Context) bool {
	enabled, ok := ctx.Value(ENABLED_CONTEXT_KEY).(bool)
	if !ok {
		return false
	}

	return enabled
}

func ClientId(ctx context.Context) string {
	id, ok := ctx.Value(CLIENT_ID_CONTEXT_KEY).(string)
	if !ok {
		// To identify requests that had no client ID set
		return fmt.Sprintf("%s-invalid-client-id", uuid.NewString()[0:16])
	}

	return id
}

func SessionId(ctx context.Context) string {
	id, ok := ctx.Value(SESSION_ID_CONTEXT_KEY).(string)
	if !ok {
		return internal.SESSION_ID
	}

	return id
}

func ServerId(ctx context.Context) string {
	id, ok := ctx.Value(SERVER_ID_CONTEXT_KEY).(string)
	if !ok {
		// To identify requests that had no server ID set
		return fmt.Sprintf("%s-invalid-server-id", uuid.NewString()[0:16])
	}

	return id
}

type AbstractTelemetryService struct {
	daytonaVersion string
	TelemetryService
}

func NewAbstractTelemetryService(version string) *AbstractTelemetryService {
	return &AbstractTelemetryService{
		daytonaVersion: version,
	}
}

func (t *AbstractTelemetryService) SetCommonProps(properties map[string]interface{}) {
	properties["daytona_version"] = t.daytonaVersion
}
