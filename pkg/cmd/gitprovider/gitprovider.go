// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var GitProviderCmd = &cobra.Command{
	Use:     "git-providers",
	Aliases: []string{"git-provider"},
	Short:   "Lists your registered Git providers",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		if len(serverConfig.GitProviders) == 0 {
			util.RenderInfoMessage("No git providers registered. Add a new git provider by preparing a Personal Access Token and running 'daytona git-providers add'")
			return
		}

		util.RenderMainTitle("Registered Git providers:")

		availableGitProviderViews := config.GetGitProviderList()
		var gitProviderViewList []config.GitProvider

		for _, gitProvider := range serverConfig.GitProviders {
			for _, availableGitProviderView := range availableGitProviderViews {
				if *gitProvider.Id == availableGitProviderView.Id {
					gitProviderViewList = append(gitProviderViewList,
						config.GitProvider{
							Id:       *gitProvider.Id,
							Name:     availableGitProviderView.Name,
							Username: *gitProvider.Username,
						},
					)
				}
			}
		}

		for _, gitProviderView := range gitProviderViewList {
			util.RenderListLine(fmt.Sprintf("%s (%s)", gitProviderView.Name, gitProviderView.Username))
		}
	},
}

func init() {
	GitProviderCmd.AddCommand(gitProviderAddCmd)
	GitProviderCmd.AddCommand(gitProviderDeleteCmd)
}
