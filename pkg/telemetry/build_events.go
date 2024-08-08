// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
)

type BuildEvent string

const (
	// Build events
	BuildEventRunBuild      BuildEvent = "server_run_build"
	BuildEventRunBuildError BuildEvent = "server_run_build_error"
)

func NewBuildRunnerEventProps(ctx context.Context, buildId, buildState string) map[string]interface{} {
	props := map[string]interface{}{}

	sessionId := SessionId(ctx)
	serverId := ServerId(ctx)

	props["session_id"] = sessionId
	props["server_id"] = serverId

	props["build_id"] = buildId
	props["build_state"] = buildState

	return props
}
