// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/pkg/cmd/server/daemon"
	"github.com/daytonaio/daytona/pkg/views"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the Daytona Server daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		views.RenderInfoMessageBold("Stopping the Daytona Server daemon...")
		return daemon.Stop()
	},
}
