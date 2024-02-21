// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_git_provider

import (
	"context"

	"github.com/daytonaio/daytona/cli/api"
	views_git_provider "github.com/daytonaio/daytona/cli/cmd/views/git_provider"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/common/api_client"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var gitProviderDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"remove"},
	Short:   "Unregister a Git providers",
	Run: func(cmd *cobra.Command, args []string) {
		apiClient := api.GetServerApiClient("http://localhost:3000", "")

		serverConfig, _, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			log.Fatal(err)
		}

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		gitProviderSelectView := views_git_provider.GitProviderSelectView{
			Id:       "",
			Username: "",
			Token:    "",
		}

		views_git_provider.GitProviderSelectionView(&gitProviderSelectView, serverConfig.GitProviders, true)

		if gitProviderSelectView.Id == "" {
			log.Fatal("Git provider id can not be blank")
			return
		}

		gitProviderList := serverConfig.GitProviders

		var providerExists bool

		for _, gitProvider := range gitProviderList {
			if *gitProvider.Id == gitProviderSelectView.Id {
				providerExists = true
			}
		}

		if providerExists {
			gitProviderList = removeGitProviderById(gitProviderSelectView.Id, gitProviderList)
		} else {
			views_util.RenderInfoMessage("Git provider is not registered")
			return
		}

		serverConfig.GitProviders = gitProviderList

		_, _, err = apiClient.ServerAPI.SetConfig(context.Background()).Config(*serverConfig).Execute()
		if err != nil {
			log.Fatal(err)
		}

		err = c.Save()
		if err != nil {
			log.Fatal(err)
		}

		views_util.RenderInfoMessage("Git provider has been removed")
	},
}

func removeGitProviderById(idToRemove string, gitProviderList []api_client.GitProvider) []api_client.GitProvider {
	var newList []api_client.GitProvider
	for _, provider := range gitProviderList {
		if *provider.Id != idToRemove {
			newList = append(newList, provider)
		}
	}
	return newList
}
