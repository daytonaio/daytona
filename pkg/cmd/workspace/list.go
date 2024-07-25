// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/views"
	list_view "github.com/daytonaio/daytona/pkg/views/workspace/list"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var verbose bool

var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List workspaces",
	Args:    cobra.ExactArgs(0),
	Aliases: []string{"ls"},
	GroupID: util.WORKSPACE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var specifyGitProviders bool

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Verbose(verbose).Execute()

		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		if len(gitProviders) > 1 {
			specifyGitProviders = true
		}

		if output.FormatFlag != "" {
			output.Output = workspaceList
			return
		}

		if len(workspaceList) == 0 {
			views.RenderInfoMessage("The workspace list is empty. Start off by running 'daytona create'.")
			return
		}

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		list_view.ListWorkspaces(workspaceList, specifyGitProviders, verbose, activeProfile.Name)
	},
}

func init() {
	ListCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
}
