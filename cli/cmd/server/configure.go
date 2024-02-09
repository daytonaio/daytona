// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	view "github.com/daytonaio/daytona/cli/cmd/views/server/configuration_prompt"
	"github.com/daytonaio/daytona/cli/connection"
	server_proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/common/grpc/proto/types"
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Daytona Server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		conn, err := connection.Get(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		client := server_proto.NewServerClient(conn)

		response, err := client.GetConfig(ctx, &empty.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		config := types.ServerConfig{
			ProjectBaseImage:    response.ProjectBaseImage,
			DefaultWorkspaceDir: response.DefaultWorkspaceDir,
			PluginsDir:          response.PluginsDir,
		}

		view.ConfigurationForm(&config)

		_, err = client.SetConfig(ctx, &server_proto.SetConfigRequest{
			ProjectBaseImage:    config.ProjectBaseImage,
			DefaultWorkspaceDir: config.DefaultWorkspaceDir,
			PluginsDir:          config.PluginsDir,
		})
		if err != nil {
			log.Fatal(err)
		}
	},
}
