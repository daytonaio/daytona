// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package runner

import (
	"github.com/daytonaio/daytona/pkg/cmd/common/daemon"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restarts the runner",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		views.RenderInfoMessage("Stopping the Daytona Runner daemon...")
		err := daemon.Stop(svcConfig)
		if err != nil {
			return err
		}

		c, err := server.GetConfig()
		if err != nil {
			return err
		}

		views.RenderInfoMessage("Starting the Daytona Runner daemon...")
		err = daemon.Start(c.LogFile.Path, svcConfig)
		if err != nil {
			return err
		}
		views.RenderContainerLayout(views.GetBoldedInfoMessage("Daytona Runner daemon restarted successfully"))
		return nil
	},
}
