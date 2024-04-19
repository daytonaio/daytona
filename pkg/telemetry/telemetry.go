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

const ENABLED_HEADER = "X-Daytona-Telemetry-Enabled"
const SESSION_ID_HEADER = "X-Daytona-Session-Id"
const SOURCE_HEADER = "X-Daytona-Source"
const CLI_ID_HEADER = "X-Daytona-CLI-Id"

type TelemetryContextKey string

var (
	ENABLED_CONTEXT_KEY    TelemetryContextKey = "telemetry-enabled"
	CLI_ID_CONTEXT_KEY     TelemetryContextKey = "cli-id"
	SESSION_ID_CONTEXT_KEY TelemetryContextKey = "session-id"
	SERVER_ID_CONTEXT_KEY  TelemetryContextKey = "server-id"
)

type TelemetryService interface {
	io.Closer
	TrackCliEvent(event CliEvent, cliId string, properties map[string]interface{}) error
	TrackServerEvent(event ServerEvent, cliId string, properties map[string]interface{}) error
	SetCommonProps(properties map[string]interface{})
}

func TelemetryEnabled(ctx context.Context) bool {
	enabled, ok := ctx.Value(ENABLED_CONTEXT_KEY).(bool)
	if !ok {
		return false
	}

	return enabled
}

func CliId(ctx context.Context) string {
	id, ok := ctx.Value(CLI_ID_CONTEXT_KEY).(string)
	if !ok {
		// To identify requests that had no CLI ID set
		return fmt.Sprintf("%s-invalid-cli-id", uuid.NewString()[0:16])
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

func NewAbstractTelemetryService() *AbstractTelemetryService {
	return &AbstractTelemetryService{
		daytonaVersion: internal.Version,
	}
}

func (t *AbstractTelemetryService) SetCommonProps(properties map[string]interface{}) {
	properties["daytona_version"] = t.daytonaVersion
}
