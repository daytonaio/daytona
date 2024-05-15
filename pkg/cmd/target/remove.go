// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var targetRemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove target",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"rm", "delete"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 1 {
			targetName := args[0]
			client, err := server.GetApiClient(nil)
			if err != nil {
				log.Fatal(err)
			}

			res, err := client.TargetAPI.RemoveTarget(context.Background(), targetName).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			views.RenderInfoMessageBold("Target removed successfully")
			return
		}
		
		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		targets, err := server.GetTargetList()
		if err != nil {
			log.Fatal(err)
		}

		selectedTarget, err := target.GetTargetFromPrompt(targets, activeProfile.Name, false)
		if err != nil {
			log.Fatal(err)
		}

		client, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		res, err := client.TargetAPI.RemoveTarget(context.Background(), *selectedTarget.Name).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessageBold("Target removed successfully")
	},
}
