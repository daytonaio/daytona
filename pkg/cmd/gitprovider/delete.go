// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"

	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	"github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var gitProviderDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"remove", "rm"},
	Short:   "Unregister a Git providers",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		gitProviderList := serverConfig.GitProviders

		var gitProviderData serverapiclient.GitProvider
		gitProviderData.Id = new(string)
		gitProviderData.Username = new(string)
		gitProviderData.Token = new(string)
		gitProviderData.BaseApiUrl = new(string)

		if len(gitProviderList) == 0 {
			util.RenderInfoMessage("No git providers registered")
			return
		}

		gitprovider_view.GitProviderSelectionView(&gitProviderData, gitProviderList, true)

		if *gitProviderData.Id == "" {
			log.Fatal("Git provider id can not be blank")
			return
		}

		_, err = apiClient.GitProviderAPI.RemoveGitProvider(ctx, *gitProviderData.Id).Execute()
		if err != nil {
			log.Fatal(err)
		}

		util.RenderInfoMessage("Git provider has been removed")
	},
}
