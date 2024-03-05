// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	list_view "github.com/daytonaio/daytona/pkg/views/workspace/list"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List workspaces",
	Args:    cobra.ExactArgs(0),
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var specifyGitProviders bool

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfig(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		if len(serverConfig.GitProviders) > 1 {
			specifyGitProviders = true
		}

		if output.FormatFlag != "" {
			output.Output = workspaceList
			return
		}

		if len(workspaceList) == 0 {
			views_util.RenderInfoMessage("The workspace list is empty. Start off by running 'daytona create'.")
			return
		}

		list_view.ListWorkspaces(workspaceList, specifyGitProviders)
	},
}
