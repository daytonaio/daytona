// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"

	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var gitProviderDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"remove", "rm"},
	Short:   "Unregister a Git providers",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			log.Fatal(apiclient_util.HandleErrorResponse(res, err))
		}

		var gitProviderData apiclient.GitProvider
		gitProviderData.Id = new(string)
		gitProviderData.Identity = new(string)
		gitProviderData.Username = new(string)
		gitProviderData.Token = new(string)
		gitProviderData.BaseApiUrl = new(string)

		if len(gitProviders) == 0 {
			views.RenderInfoMessage("No git providers registered")
			return
		}

		gitprovider_view.GitProviderSelectionView(&gitProviderData, gitProviders, true)

		if *gitProviderData.Id == "" || *gitProviderData.Identity == "" {
			log.Fatal("Git provider id and identity can not be blank")
			return
		}

		_, err = apiClient.GitProviderAPI.RemoveGitProvider(ctx, *gitProviderData.Id).Execute()
		if err != nil {
			log.Fatal(err)
		}

		views.RenderInfoMessage("Git provider has been removed")
	},
}
