// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"github.com/spf13/cobra"
)

var TelemetryCmd = &cobra.Command{
	Use:   "telemetry",
	Short: "Manage telemetry collection",
	Args:  cobra.NoArgs,
}

func init() {
	TelemetryCmd.AddCommand(enableCmd)
	TelemetryCmd.AddCommand(disableCmd)
}
