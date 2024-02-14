// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	view "github.com/daytonaio/daytona/cli/cmd/views/server/configuration_prompt"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
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

		config, err := client.GetConfig(ctx, &empty.Empty{})
		if err != nil {
			log.Fatal(err)
		}

		view.ConfigurationForm(config)

		_, err = client.SetConfig(ctx, &types.ServerConfig{
			ProjectBaseImage:    config.ProjectBaseImage,
			DefaultWorkspaceDir: config.DefaultWorkspaceDir,
			PluginsDir:          config.PluginsDir,
			PluginRegistryUrl:   config.PluginRegistryUrl,
		})
		if err != nil {
			log.Fatal(err)
		}

		views_util.RenderInfoMessage("Server configuration updated. You might need to restart the server for the changes to take effect.")
	},
}
