// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_git_provider

import (
	"context"

	views_git_provider "github.com/daytonaio/daytona/cli/cmd/views/git_provider"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/connection"
	server_proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/common/grpc/proto/types"
	"github.com/golang/protobuf/ptypes/empty"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var gitProviderDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"remove"},
	Short:   "Unregister a Git providers",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		conn, err := connection.Get(nil)
		if err != nil {
			log.Fatal(err)
		}

		defer conn.Close()

		client := server_proto.NewServerClient(conn)

		serverConfig, err := client.GetConfig(ctx, &empty.Empty{})
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

		views_git_provider.GitProviderSelectionView(&gitProviderSelectView, true)

		if gitProviderSelectView.Id == "" {
			log.Fatal("Git provider id can not be blank")
			return
		}

		gitProviderList := serverConfig.GitProviders

		var providerExists bool

		for _, gitProvider := range gitProviderList {
			if gitProvider.Id == gitProviderSelectView.Id {
				providerExists = true
			}
		}

		if providerExists {
			gitProviderList = removeGitProviderById(gitProviderSelectView.Id, gitProviderList)
		} else {
			views_util.RenderInfoMessage("Git provider is not registered")
			return
		}

		_, err = client.SetConfig(ctx, &types.ServerConfig{
			ProjectBaseImage:    serverConfig.ProjectBaseImage,
			DefaultWorkspaceDir: serverConfig.DefaultWorkspaceDir,
			PluginsDir:          serverConfig.PluginsDir,
			GitProviders:        gitProviderList,
		})
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

func removeGitProviderById(idToRemove string, gitProviderList []*types.GitProvider) []*types.GitProvider {
	var newList []*types.GitProvider
	for _, provider := range gitProviderList {
		if provider.Id != idToRemove {
			newList = append(newList, provider)
		}
	}
	return newList
}
