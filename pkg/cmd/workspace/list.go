// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"encoding/json"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/cmd/format"
	"github.com/daytonaio/daytona/pkg/views/workspace/list"
	"github.com/spf13/cobra"
)

var labelFilters []string

var ListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List workspaces",
	Args:    cobra.NoArgs,
	GroupID: util.TARGET_GROUP,
	Aliases: common.GetAliases("list"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var specifyGitProviders bool

		apiClient, err := apiclient.GetApiClient(nil)
		if err != nil {
			return err
		}

		labels, err := common.MapKeyValue(labelFilters)
		if err != nil {
			return err
		}

		encoded, err := json.Marshal(labels)
		if err != nil {
			return err
		}

		workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Labels(string(encoded)).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			return apiclient.HandleErrorResponse(res, err)
		}

		if len(gitProviders) > 1 {
			specifyGitProviders = true
		}

		if format.FormatFlag != "" {
			formattedData := format.NewFormatter(workspaceList)
			formattedData.Print()
			return nil
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		list.ListWorkspaces(workspaceList, specifyGitProviders, activeProfile.Name)
		return nil
	},
}

func init() {
	ListCmd.Flags().StringSliceVarP(&labelFilters, "label", "l", nil, "Filter by label")
	format.RegisterFormatFlag(ListCmd)
}
