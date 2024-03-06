// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views/target"
	"github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var targetRemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove target",
	Args:    cobra.NoArgs,
	Aliases: []string{"rm", "delete"},
	Run: func(cmd *cobra.Command, args []string) {
		targets, err := server.GetTargetList()
		if err != nil {
			log.Fatal(err)
		}

		selectedTarget, err := target.GetTargetFromPrompt(targets, false)
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

		util.RenderInfoMessageBold("Target removed successfully")
	},
}
