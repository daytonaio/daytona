// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"fmt"

	view "github.com/daytonaio/daytona/cli/cmd/views/workspace/select_prompt"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/connection"
	server_proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/pkg/git_provider"
	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CreateTempCmd = &cobra.Command{
	Use:   "create-temp",
	Short: "Create a workspace",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var providerId string
		var namespaceId string
		var gitProvider git_provider.GitProvider

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

		if serverConfig.GitProviders == nil || len(serverConfig.GitProviders) == 0 {
			fmt.Println("No Git providers registered")
			return
		}

		if len(serverConfig.GitProviders) == 1 {
			providerId = serverConfig.GitProviders[0].Id
		} else {

			availableGitProviderViews := config.GetGitProviderList()
			var gitProviderViewList []config.GitProvider

			for _, gitProvider := range serverConfig.GitProviders {
				for _, availableGitProviderView := range availableGitProviderViews {
					if gitProvider.Id == availableGitProviderView.Id {
						gitProviderViewList = append(gitProviderViewList, availableGitProviderView)
					}
				}
			}
			providerId = view.GetProviderIdFromPrompt(gitProviderViewList)
		}

		gitProvider = git_provider.CreateGitProvider(providerId, serverConfig.GitProviders)
		user, err := gitProvider.GetUserData()
		if err != nil {
			log.Fatal(err)
		}

		namespaceList, err := gitProvider.GetNamespaces()
		if err != nil {
			log.Fatal(err)
		}

		if len(namespaceList) == 1 {
			namespaceId = namespaceList[0].Id
		} else {
			var namespaceViewList []git_provider.GitNamespace
			namespaceViewList = append(namespaceViewList, namespaceList...)
			namespaceId = view.GetNamespaceIdFromPrompt(namespaceViewList)
		}

		repos, err := gitProvider.GetRepositories(namespaceId)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(repos)
		fmt.Println(len(repos))
		fmt.Println(user.Username)
	},
}
