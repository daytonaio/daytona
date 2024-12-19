// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"github.com/daytonaio/daytona/pkg/cmd/common/daemon"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the runner",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		views.RenderInfoMessageBold("Stopping the Daytona Runner daemon...")
		return daemon.Stop(svcConfig)
	},
}
