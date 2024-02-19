// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_git_provider

import (
	"context"
	"fmt"

	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/connection"
	server_proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/golang/protobuf/ptypes/empty"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var GitProviderCmd = &cobra.Command{
	Use:     "git-providers",
	Aliases: []string{"git-provider"},
	Short:   "Lists your registered Git providers",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		conn, err := connection.GetGrpcConn(nil)
		if err != nil {
			log.Fatal(err)
		}

		defer conn.Close()

		client := server_proto.NewServerClient(conn)

		serverConfig, err := client.GetConfig(ctx, &empty.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		if len(serverConfig.GitProviders) == 0 {
			fmt.Println("No Git providers registered")
			return
		}

		views_util.RenderMainTitle("Registered Git providers:")

		availableGitProviderViews := config.GetGitProviderList()
		var gitProviderViewList []config.GitProvider

		for _, gitProvider := range serverConfig.GitProviders {
			for _, availableGitProviderView := range availableGitProviderViews {
				if gitProvider.Id == availableGitProviderView.Id {
					gitProviderViewList = append(gitProviderViewList,
						config.GitProvider{
							Id:       gitProvider.Id,
							Name:     availableGitProviderView.Name,
							Username: gitProvider.Username,
						},
					)
				}
			}
		}

		for _, gitProviderView := range gitProviderViewList {
			views_util.RenderListLine(fmt.Sprintf("%s (%s)", gitProviderView.Name, gitProviderView.Username))
		}
	},
}

func init() {
	GitProviderCmd.AddCommand(gitProviderAddCmd)
	GitProviderCmd.AddCommand(gitProviderDeleteCmd)
}
