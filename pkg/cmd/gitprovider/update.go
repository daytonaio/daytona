// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	"github.com/spf13/cobra"
)

var gitProviderUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a Git provider",
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

		if len(gitProviders) == 0 {
			views.RenderInfoMessage("No git providers registered")
			return nil
		}

		selectedGitProvider, err := gitprovider_view.GetGitProviderFromPrompt(ctx, gitProviders, apiClient)
		if err != nil {
			return err
		}

		if selectedGitProvider == nil {
			return nil
		}

		existingAliases := util.ArrayMap(gitProviders, func(gp apiclient.GitProvider) string {
			return gp.Alias
		})

		setGitProviderConfig := apiclient.SetGitProviderConfig{
			Id:            &selectedGitProvider.Id,
			ProviderId:    selectedGitProvider.ProviderId,
			Token:         selectedGitProvider.Token,
			BaseApiUrl:    selectedGitProvider.BaseApiUrl,
			Username:      &selectedGitProvider.Username,
			Alias:         &selectedGitProvider.Alias,
			SigningMethod: selectedGitProvider.SigningMethod,
			SigningKey:    selectedGitProvider.SigningKey,
		}

		err = gitprovider_view.GitProviderCreationView(ctx, apiClient, &setGitProviderConfig, existingAliases)
		if err != nil {
			return err
		}

		res, err = apiClient.GitProviderAPI.SetGitProvider(ctx).GitProviderConfig(setGitProviderConfig).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Git provider has been updated")
		return nil
	},
}
