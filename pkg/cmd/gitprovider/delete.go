// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
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
	Aliases: []string{"remove"},
	Short:   "Unregister a Git providers",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient, err := server.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		gitProviderSelectView := gitprovider_view.GitProviderSelectView{
			Id:       "",
			Username: "",
			Token:    "",
		}

		gitProviderList := serverConfig.GitProviders

		if len(gitProviderList) == 0 {
			util.RenderInfoMessage("No git providers registered")
			return
		}

		gitprovider_view.GitProviderSelectionView(&gitProviderSelectView, serverConfig.GitProviders, true)

		if gitProviderSelectView.Id == "" {
			log.Fatal("Git provider id can not be blank")
			return
		}

		var providerExists bool

		for _, gitProvider := range gitProviderList {
			if *gitProvider.Id == gitProviderSelectView.Id {
				providerExists = true
			}
		}

		if providerExists {
			gitProviderList = removeGitProviderById(gitProviderSelectView.Id, gitProviderList)
		} else {
			util.RenderInfoMessage("Git provider is not registered")
			return
		}

		serverConfig.GitProviders = gitProviderList

		_, res, err = apiClient.ServerAPI.SetConfig(context.Background()).Config(*serverConfig).Execute()
		if err != nil {
			log.Fatal(apiclient.HandleErrorResponse(res, err))
		}

		err = c.Save()
		if err != nil {
			log.Fatal(err)
		}

		util.RenderInfoMessage("Git provider has been removed")
	},
}

func removeGitProviderById(idToRemove string, gitProviderList []serverapiclient.GitProvider) []serverapiclient.GitProvider {
	var newList []serverapiclient.GitProvider
	for _, provider := range gitProviderList {
		if *provider.Id != idToRemove {
			newList = append(newList, provider)
		}
	}
	return newList
}
