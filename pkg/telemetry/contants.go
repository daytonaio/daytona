// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

const ENABLED_HEADER = "X-Daytona-Telemetry-Enabled"
const SESSION_ID_HEADER = "X-Daytona-Session-Id"
const SOURCE_HEADER = "X-Daytona-Source"
const CLIENT_ID_HEADER = "X-Daytona-Client-Id"

type TelemetryContextKey string

var (
	ENABLED_CONTEXT_KEY    TelemetryContextKey = "telemetry-enabled"
	CLIENT_ID_CONTEXT_KEY  TelemetryContextKey = "cli-id"
	SESSION_ID_CONTEXT_KEY TelemetryContextKey = "session-id"
	SERVER_ID_CONTEXT_KEY  TelemetryContextKey = "server-id"
)

type TelemetrySource string

var (
	CLI_SOURCE           TelemetrySource = "cli"
	CLI_WORKSPACE_SOURCE TelemetrySource = "cli-workspace"
	AGENT_SOURCE         TelemetrySource = "agent"
	RUNNER_SOURCE        TelemetrySource = "runner"
)
