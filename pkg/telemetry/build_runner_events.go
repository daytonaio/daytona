// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"
)

type BuildRunnerEvent string

const (
	// Purge events
	BuildRunnerEventPurgeStarted   BuildRunnerEvent = "build_runner_purge_started"
	BuildRunnerEventPurgeCompleted BuildRunnerEvent = "build_runner_purge_completed"
	BuildRunnerEventPurgeError     BuildRunnerEvent = "build_runner_purge_error"

	// Build events
	BuildRunnerEventRunBuild      BuildRunnerEvent = "build_runner_run_build"
	BuildRunnerEventRunBuildError BuildRunnerEvent = "build_runner_run_build_error"
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
