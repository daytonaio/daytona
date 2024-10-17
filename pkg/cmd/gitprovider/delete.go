// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var gitProviderDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"remove", "rm"},
	Short:   "Unregister a Git provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if allFlag {
			for _, gitProvider := range gitProviders {
				_, err := apiClient.GitProviderAPI.RemoveGitProvider(ctx, gitProvider.Id).Execute()
				if err != nil {
					return err
				}
			}

			views.RenderInfoMessage("All Git providers have been removed")
			return nil
		}

		if len(gitProviders) == 0 {
			views.RenderInfoMessage("No git providers registered")
			return nil
		}

		selectedGitProvider := selection.GetGitProviderConfigFromPrompt(selection.GetGitProviderConfigParams{
			GitProviderConfigs: gitProviders,
			ActionVerb:         "Remove",
		})

		if selectedGitProvider == nil {
			return nil
		}

		_, err = apiClient.GitProviderAPI.RemoveGitProvider(ctx, selectedGitProvider.Id).Execute()
		if err != nil {
			return err
		}

		views.RenderInfoMessage("Git provider has been removed")
		return nil
	},
}

var allFlag bool

func init() {
	gitProviderDeleteCmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Remove all Git providers")
}
