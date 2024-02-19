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
	"github.com/daytonaio/daytona/pkg/git_provider"
	"github.com/golang/protobuf/ptypes/empty"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var gitProviderAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new", "register, update"},
	Short:   "Register a Git providers",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var providerExists bool

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
			if gitProvider.Id == gitProviderSelectView.Id {
				gitProvider.Token = gitProviderSelectView.Token
				gitProvider.Username = gitProviderSelectView.Username
				providerExists = true
			}
		}

		if !providerExists {
			gitProviderList = append(serverConfig.GitProviders, &types.GitProvider{
				Id:       gitProviderSelectView.Id,
				Username: gitProviderSelectView.Username,
				Token:    gitProviderSelectView.Token,
			})
		}

		serverConfig.GitProviders = gitProviderList

		_, err = client.SetConfig(ctx, serverConfig)
		if err != nil {
			log.Fatal(err)
		}

		err = c.Save()
		if err != nil {
			log.Fatal(err)
		}

		views_util.RenderInfoMessage("Git provider has been registered")
	},
}
