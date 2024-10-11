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

var GitProviderAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new", "register"},
	Short:   "Register a Git provider",
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

		existingAliases := util.ArrayMap(gitProviders, func(gp apiclient.GitProvider) string {
			return gp.Alias
		})

		for _, gp := range gitProviders {
			existingAliases = append(existingAliases, gp.Alias)
		}

		setGitProviderConfig := apiclient.SetGitProviderConfig{}
		setGitProviderConfig.BaseApiUrl = new(string)
		setGitProviderConfig.Username = new(string)
		setGitProviderConfig.Alias = new(string)

		err = gitprovider_view.GitProviderCreationView(ctx, apiClient, &setGitProviderConfig, existingAliases)
		if err != nil {
			return err
		}

		if setGitProviderConfig.ProviderId == "" {
			return nil
		}

		res, err = apiClient.GitProviderAPI.SetGitProvider(ctx).GitProviderConfig(setGitProviderConfig).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Git provider has been registered")
		return nil
	},
}
