// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

type CliEvent string

const (
	CliEventCmdStart CliEvent = "cli_cmd_start"
	CliEventCmdEnd   CliEvent = "cli_cmd_end"
)
