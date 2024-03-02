// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	"github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var gitProviderAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new", "register", "update"},
	Short:   "Register a Git providers",
	Run: func(cmd *cobra.Command, args []string) {
		var providerExists bool

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

		gitprovider_view.GitProviderSelectionView(&gitProviderSelectView, serverConfig.GitProviders, false)

		if gitProviderSelectView.Id == "" {
			return
		}

		if gitProviderSelectView.Username == "" {
			gitUsername, err := gitprovider.GetUsernameFromToken(gitProviderSelectView.Id, config.GetGitProviderList(), gitProviderSelectView.Token)
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
			gitProviderList = append(serverConfig.GitProviders, serverapiclient.GitProvider{
				Id:       &gitProviderSelectView.Id,
				Username: &gitProviderSelectView.Username,
				Token:    &gitProviderSelectView.Token,
			})
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

		util.RenderInfoMessage("Git provider has been registered")
	},
}
