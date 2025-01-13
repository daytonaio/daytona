// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"strings"

	"github.com/spf13/cobra"
)

type CliEventName string

const (
	CliEventCommandStarted      CliEventName = "cli_command_started"
	CliEventCommandCompleted    CliEventName = "cli_command_completed"
	CliEventCommandFailed       CliEventName = "cli_command_failed"
	CliEventCommandInvalid      CliEventName = "cli_command_invalid"
	CliEventCommandInterrupted  CliEventName = "cli_command_interrupted"
	CliEventTargetOpened        CliEventName = "cli_target_opened"
	CliEventTargetOpenFailed    CliEventName = "cli_target_open_failed"
	CliEventWorkspaceOpened     CliEventName = "cli_workspace_opened"
	CliEventWorkspaceOpenFailed CliEventName = "cli_workspace_open_failed"
	CliEventDefaultIdeSet       CliEventName = "cli_default_ide_set"
)

type cliEvent struct {
	AbstractEvent
	cmd   *cobra.Command
	flags []string
}

func NewCliEvent(name CliEventName, cmd *cobra.Command, flags []string, err error, extras map[string]interface{}) Event {
	return cliEvent{
		AbstractEvent: AbstractEvent{
			name:   string(name),
			extras: extras,
			err:    err,
		},
		cmd:   cmd,
		flags: flags,
	}
}

func (e cliEvent) Props() map[string]interface{} {
	props := e.AbstractEvent.Props()

	if e.cmd == nil {
		return props
	}

	path := e.cmd.CommandPath()

	// Trim daytona from the path if a non-root command was invoked
	// This prevents a `daytona` pileup in the telemetry data
	if path != "daytona" {
		path = strings.TrimPrefix(path, "daytona ")
	}

	calledAs := e.cmd.CalledAs()

	props["command"] = path
	props["called_as"] = calledAs
	props["flags"] = e.flags

	return props
}
