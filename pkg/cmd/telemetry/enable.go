// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/views"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable telemetry collection",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		err = c.EnableTelemetry()
		if err != nil {
			log.Fatal(err)
		}

		views.RenderInfoMessage("Telemetry collection enabled")
	},
}
