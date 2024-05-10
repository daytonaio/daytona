// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacemode

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the project",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		res, err := apiClient.WorkspaceAPI.StartProject(ctx, workspaceId, projectName).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		views.RenderInfoMessage("Project successfully started")
	},
}
