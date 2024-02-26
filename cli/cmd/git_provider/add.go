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
	"github.com/daytonaio/daytona/pkg/git_provider"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var gitProviderAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new", "register", "update"},
	Short:   "Register a Git providers",
	Run: func(cmd *cobra.Command, args []string) {
		var providerExists bool

		apiClient, err := api.GetServerApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		serverConfig, res, err := apiClient.ServerAPI.GetConfig(context.Background()).Execute()
		if err != nil {
			log.Fatal(api.HandleErrorResponse(res, err))
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

		views_git_provider.GitProviderSelectionView(&gitProviderSelectView, serverConfig.GitProviders, false)

		if gitProviderSelectView.Id == "" {
			return
		}

		if gitProviderSelectView.Username == "" {
			gitUsername, err := git_provider.GetUsernameFromToken(gitProviderSelectView.Id, config.GetGitProviderList(), gitProviderSelectView.Token)
			if err != nil {
				log.Fatal(err)
			}
			gitProviderSelectView.Username = gitUsername
		}

		gitProviderList := serverConfig.GitProviders

		for _, gitProvider := range gitProviderList {
			if *gitProvider.Id == gitProviderSelectView.Id {
				*gitProvider.Username = gitProviderSelectView.Username
				*gitProvider.Token = gitProviderSelectView.Token
				providerExists = true
			}
		}

		if !providerExists {
			gitProviderList = append(serverConfig.GitProviders, api_client.GitProvider{
				Id:       &gitProviderSelectView.Id,
				Username: &gitProviderSelectView.Username,
				Token:    &gitProviderSelectView.Token,
			})
		}

		serverConfig.GitProviders = gitProviderList

		_, res, err = apiClient.ServerAPI.SetConfig(context.Background()).Config(*serverConfig).Execute()
		if err != nil {
			log.Fatal(api.HandleErrorResponse(res, err))
		}

		err = c.Save()
		if err != nil {
			log.Fatal(err)
		}

		views_util.RenderInfoMessage("Git provider has been registered")
	},
}
